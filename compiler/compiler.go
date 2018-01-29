package compiler

import (
	"errors"
	"fmt"
	"sort"

	"github.com/influxdata/ifql/semantic"
)

type Type int

const (
	TInvalid Type = iota
	TBool
	TInt
	TUInt
	TFloat
	TString
	TTime
	TMap
)

type Time int64

func Compile(f *semantic.ArrowFunctionExpression, types map[ReferencePath]Type) (Func, error) {
	references, err := FindReferences(f)
	if err != nil {
		return nil, err
	}
	// Validate we have types for each reference
	for _, r := range references {
		rp := r.Path()
		if types[rp] == TInvalid {
			return nil, fmt.Errorf("missing type information for %q", rp)
		}
	}

	root, err := compile(f.Body, types)
	if err != nil {
		return nil, err
	}
	cpy := make(map[ReferencePath]Type, len(types))
	for k, v := range types {
		cpy[k] = v
	}
	return compiledFn{
		root:  root,
		types: cpy,
	}, nil
}

// FindReferences returns all references in the expression.
func FindReferences(f *semantic.ArrowFunctionExpression) ([]Reference, error) {
	return findReferences(f.Body.(semantic.Expression))
}

func findReferences(n semantic.Expression) ([]Reference, error) {
	switch n := n.(type) {
	case *semantic.ObjectExpression:
		var refs []Reference
		for _, p := range n.Properties {
			r, err := findReferences(p.Value)
			if err != nil {
				return nil, err
			}
			refs = append(refs, r...)
		}
		return refs, nil
	case *semantic.MemberExpression:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		return []Reference{r}, nil
	case *semantic.Identifier:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		return []Reference{r}, nil
	case *semantic.UnaryExpression:
		return findReferences(n.Argument)
	case *semantic.LogicalExpression:
		l, err := findReferences(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := findReferences(n.Right)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	case *semantic.BinaryExpression:
		l, err := findReferences(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := findReferences(n.Right)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	default:
		return nil, nil
	}
}

func determineReference(n semantic.Expression) (Reference, error) {
	switch n := n.(type) {
	case *semantic.MemberExpression:
		r, err := determineReference(n.Object)
		if err != nil {
			return nil, err
		}
		r = append(r, n.Property)
		return r, nil
	case *semantic.Identifier:
		return Reference{n.Name}, nil
	default:
		return nil, fmt.Errorf("unexpected reference expression type %T", n)
	}
}

// CompilationCache caches compilation results based on the types of the input parameters.
type CompilationCache struct {
	fn        *semantic.ArrowFunctionExpression
	pathOrder []ReferencePath
	root      *compilationCacheNode
}

func NewCompilationCache(fn *semantic.ArrowFunctionExpression, referencePaths []ReferencePath) *CompilationCache {
	pathOrder := make([]ReferencePath, len(referencePaths))
	copy(pathOrder, referencePaths)
	sort.Slice(pathOrder, func(i, j int) bool { return pathOrder[i] < pathOrder[j] })
	return &CompilationCache{
		fn:        fn,
		pathOrder: pathOrder,
		root:      new(compilationCacheNode),
	}
}

// Compile returnes a compiled function bsaed on the provided types.
// The result will be cached for subsequent calls.
func (c *CompilationCache) Compile(types map[ReferencePath]Type) (Func, error) {
	return c.root.compile(c.fn, c.pathOrder, types)
}

type compilationCacheNode struct {
	children map[Type]*compilationCacheNode

	fn  Func
	err error
}

// compile recursively searches for a matching child node that has compiled the function.
// If the compilation has not been performed previously its result is cached and returned.
func (c *compilationCacheNode) compile(fn *semantic.ArrowFunctionExpression, order []ReferencePath, types map[ReferencePath]Type) (Func, error) {
	if len(order) == 0 {
		// We are the matching child, return the cached result or do the compilation.
		if c.fn == nil && c.err == nil {
			c.fn, c.err = Compile(fn, types)
		}
		return c.fn, c.err
	}
	// Find the matching child based on the order.
	next := order[0]
	t := types[next]
	child := c.children[t]
	if child == nil {
		child = new(compilationCacheNode)
		if c.children == nil {
			c.children = make(map[Type]*compilationCacheNode)
		}
		c.children[t] = child
	}
	return child.compile(fn, order[1:], types)
}

type compiledType Type

func (c compiledType) Type() Type {
	return Type(c)
}
func (c compiledType) MapMeta() MapMeta {
	panic(c.error(TMap))
}
func (c compiledType) error(exp Type) error {
	return typeErr{Actual: Type(c), Expected: exp}
}

func compile(n semantic.Node, types map[ReferencePath]Type) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.BlockStatement:
		body := make([]Evaluator, len(n.Body))
		hasReturn := false
		for i, s := range n.Body {
			node, err := compile(s, types)
			if err != nil {
				return nil, err
			}
			body[i] = node
			if _, ok := node.(returnEvaluator); ok {
				// The returnEvaluator indicates the last statement
				hasReturn = true
				break
			}
		}
		if !hasReturn {
			return nil, errors.New("block has no return statement")
		}
		return &blockEvaluator{
			compiledType: compiledType(TBool),
			body:         body,
		}, nil
	case *semantic.ExpressionStatement:
		//TODO(nathanielc): I belive sideffects are not possible in record functions.
		// Maybe we can skip these or throw an error?
		return compile(n.Expression, types)
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, types)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.VariableDeclaration:
		if len(n.Declarations) != 1 {
			return nil, errors.New("var declaration must have exactly one declaration")
		}
		d := n.Declarations[0]
		node, err := compile(d.Init, types)
		if err != nil {
			return nil, err
		}
		r, err := determineReference(d.ID)
		if err != nil {
			return nil, err
		}
		rp := r.Path()
		// Update type information with new type
		types[rp] = node.Type()
		return &declarationEvaluator{
			compiledType: compiledType(node.Type()),
			id:           rp,
			init:         node,
		}, nil
	case *semantic.ObjectExpression:
		meta := MapMeta{
			Properties: make([]MapPropertyMeta, len(n.Properties)),
		}
		properties := make(map[string]Evaluator, len(n.Properties))
		for i, p := range n.Properties {
			node, err := compile(p.Value, types)
			if err != nil {
				return nil, err
			}
			meta.Properties[i].Key = p.Key.Name
			meta.Properties[i].Type = node.Type()

			properties[p.Key.Name] = node
		}
		return &mapEvaluator{
			compiledType: compiledType(TMap),
			meta:         meta,
			properties:   properties,
		}, nil
	case *semantic.Identifier:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		rp := r.Path()
		return &referenceEvaluator{
			compiledType:  compiledType(types[rp]),
			referencePath: rp,
		}, nil
	case *semantic.MemberExpression:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		rp := r.Path()
		return &referenceEvaluator{
			compiledType:  compiledType(types[rp]),
			referencePath: rp,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			compiledType: compiledType(TBool),
			b:            n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			compiledType: compiledType(TInt),
			i:            n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			compiledType: compiledType(TFloat),
			f:            n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			compiledType: compiledType(TString),
			s:            n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			compiledType: compiledType(TTime),
			t:            Time(n.Value.UnixNano()),
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, types)
		if err != nil {
			return nil, err
		}
		nt := node.Type()
		if nt != TBool && nt != TInt && nt != TFloat {
			return nil, fmt.Errorf("invalid unary operator %v on type %v", n.Operator, nt)
		}
		return &unaryEvaluator{
			compiledType: compiledType(nt),
			node:         node,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, types)
		if err != nil {
			return nil, err
		}
		if l.Type() != TBool {
			return nil, fmt.Errorf("invalid left operand type %v in logical expression", l.Type())
		}
		r, err := compile(n.Right, types)
		if err != nil {
			return nil, err
		}
		if r.Type() != TBool {
			return nil, fmt.Errorf("invalid right operand type %v in logical expression", r.Type())
		}
		return &logicalEvaluator{
			compiledType: compiledType(TBool),
			operator:     n.Operator,
			left:         l,
			right:        r,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, types)
		if err != nil {
			return nil, err
		}
		lt := l.Type()
		r, err := compile(n.Right, types)
		if err != nil {
			return nil, err
		}
		rt := r.Type()
		sig := binarySignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
		}
		f, ok := binaryFuncs[sig]
		if !ok {
			return nil, fmt.Errorf("unsupported binary expression with types %v %v %v", sig.Left, sig.Operator, sig.Right)
		}
		return &binaryEvaluator{
			compiledType: compiledType(f.ResultType),
			left:         l,
			right:        r,
			f:            f.Func,
		}, nil
	default:
		return nil, fmt.Errorf("unknown semantic node of type %T", n)
	}
}

type compiledFn struct {
	root  Evaluator
	types map[ReferencePath]Type
}

func (c compiledFn) validate(scope Scope) error {
	// Validate scope
	for k, t := range c.types {
		if scope.Type(k) != t {
			return fmt.Errorf("missing or incorrectly typed value found in scope for name %q", k)
		}
	}
	return nil
}

func (c compiledFn) Type() Type {
	return c.root.Type()
}
func (c compiledFn) MapMeta() MapMeta {
	return c.root.MapMeta()
}
func (c compiledFn) Eval(scope Scope) (Value, error) {
	if err := c.validate(scope); err != nil {
		return Value{}, err
	}
	var val interface{}
	switch c.Type() {
	case TBool:
		val = c.root.EvalBool(scope)
	case TInt:
		val = c.root.EvalInt(scope)
	case TUInt:
		val = c.root.EvalUInt(scope)
	case TFloat:
		val = c.root.EvalFloat(scope)
	case TString:
		val = c.root.EvalString(scope)
	case TTime:
		val = c.root.EvalTime(scope)
	case TMap:
		val = c.root.EvalMap(scope)
	default:
		return Value{}, fmt.Errorf("unsupported type %s", c.Type())
	}
	return Value{
		Type:  c.Type(),
		Value: val,
	}, nil
}

func (c compiledFn) EvalBool(scope Scope) (bool, error) {
	if err := c.validate(scope); err != nil {
		return false, err
	}
	return c.root.EvalBool(scope), nil
}
func (c compiledFn) EvalInt(scope Scope) (int64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalInt(scope), nil
}
func (c compiledFn) EvalUInt(scope Scope) (uint64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalUInt(scope), nil
}
func (c compiledFn) EvalFloat(scope Scope) (float64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalFloat(scope), nil
}
func (c compiledFn) EvalString(scope Scope) (string, error) {
	if err := c.validate(scope); err != nil {
		return "", err
	}
	return c.root.EvalString(scope), nil
}
func (c compiledFn) EvalTime(scope Scope) (Time, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalTime(scope), nil
}
func (c compiledFn) EvalMap(scope Scope) (Map, error) {
	if err := c.validate(scope); err != nil {
		return Map{}, err
	}
	return c.root.EvalMap(scope), nil
}
