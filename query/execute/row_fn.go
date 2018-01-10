package execute

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/influxdata/ifql/ast"
	"github.com/pkg/errors"
)

type rowFn struct {
	references       []Reference
	referencePaths   []ReferencePath
	compilationCache *CompilationCache
	scope            Scope

	scopeCols map[ReferencePath]int
	types     map[ReferencePath]DataType

	preparedFn CompiledFn
}

func newRowFn(fn *ast.ArrowFunctionExpression) (rowFn, error) {
	if len(fn.Params) != 1 {
		return rowFn{}, fmt.Errorf("function should only have a single parameter, got %v", fn.Params)
	}
	references, err := FindReferences(fn)
	if err != nil {
		return rowFn{}, err
	}

	referencePaths := make([]ReferencePath, len(references))
	for i, r := range references {
		referencePaths[i] = r.Path()
	}

	return rowFn{
		references:       references,
		referencePaths:   referencePaths,
		compilationCache: NewCompilationCache(fn, referencePaths),
		scope:            make(Scope, len(references)),
		scopeCols:        make(map[ReferencePath]int, len(references)),
		types:            make(map[ReferencePath]DataType, len(references)),
	}, nil
}

func (f *rowFn) prepare(cols []ColMeta) error {
	// Prepare types and scopeCols
	for i, r := range f.references {
		rp := f.referencePaths[i]
		found := false
		for j, c := range cols {
			if r[1] == c.Label {
				f.scopeCols[rp] = j
				f.types[rp] = c.Type
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("function references unknown value %q", rp)
		}
	}
	// Compile fn for given types
	fn, err := f.compilationCache.Compile(f.types)
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

func (f *rowFn) eval(row int, rr RowReader) (Value, error) {
	for _, rp := range f.referencePaths {
		f.scope[rp] = ValueForRow(row, f.scopeCols[rp], rr)
	}
	return f.preparedFn.Eval(f.scope)
}

type RowPredicateFn struct {
	rowFn
}

func NewRowPredicateFn(fn *ast.ArrowFunctionExpression) (*RowPredicateFn, error) {
	r, err := newRowFn(fn)
	if err != nil {
		return nil, err
	}
	return &RowPredicateFn{
		rowFn: r,
	}, nil
}

func (f *RowPredicateFn) Prepare(cols []ColMeta) error {
	err := f.rowFn.prepare(cols)
	if err != nil {
		return err
	}
	if f.preparedFn.Type() != TBool {
		return errors.New("row predicate function does not evaluate to a boolean")
	}
	return nil
}

func (f *RowPredicateFn) Eval(row int, rr RowReader) (bool, error) {
	v, err := f.rowFn.eval(row, rr)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}

type RowMapFn struct {
	rowFn

	isWrap  bool
	wrapMap Map
}

func NewRowMapFn(fn *ast.ArrowFunctionExpression) (*RowMapFn, error) {
	r, err := newRowFn(fn)
	if err != nil {
		return nil, err
	}
	return &RowMapFn{
		rowFn: r,
		wrapMap: Map{
			Meta: MapMeta{
				Properties: []MapPropertyMeta{{Key: DefaultValueColLabel}},
			},
			Values: make(map[string]Value, 1),
		},
	}, nil
}

func (f *RowMapFn) Prepare(cols []ColMeta) error {
	err := f.rowFn.prepare(cols)
	if err != nil {
		return err
	}
	t := f.preparedFn.Type()
	f.isWrap = t != TMap
	if f.isWrap {
		f.wrapMap.Meta.Properties[0].Type = t
	}
	return nil
}

func (f *RowMapFn) MapMeta() MapMeta {
	if f.isWrap {
		return f.wrapMap.Meta
	}
	return f.preparedFn.MapMeta()
}

func (f *RowMapFn) Eval(row int, rr RowReader) (Map, error) {
	v, err := f.rowFn.eval(row, rr)
	if err != nil {
		return Map{}, err
	}
	if f.isWrap {
		f.wrapMap.Values[DefaultValueColLabel] = v
		return f.wrapMap, nil
	}
	return v.Map(), nil
}

// FindReferences returns all references in the expression.
func FindReferences(f *ast.ArrowFunctionExpression) ([]Reference, error) {
	return findReferences(f.Body.(ast.Expression))
}

func findReferences(n ast.Expression) ([]Reference, error) {
	switch n := n.(type) {
	case *ast.ObjectExpression:
		var refs []Reference
		for _, p := range n.Properties {
			r, err := findReferences(p.Value)
			if err != nil {
				return nil, err
			}
			refs = append(refs, r...)
		}
		return refs, nil
	case *ast.MemberExpression:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		return []Reference{r}, nil
	case *ast.Identifier:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		return []Reference{r}, nil
	case *ast.UnaryExpression:
		return findReferences(n.Argument)
	case *ast.LogicalExpression:
		l, err := findReferences(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := findReferences(n.Right)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	case *ast.BinaryExpression:
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

func determineReference(n ast.Expression) (Reference, error) {
	switch n := n.(type) {
	case *ast.MemberExpression:
		r, err := determineReference(n.Object)
		if err != nil {
			return nil, err
		}
		name, err := propertyName(n)
		if err != nil {
			return nil, err
		}
		r = append(r, name)
		return r, nil
	case *ast.Identifier:
		return Reference{n.Name}, nil
	default:
		return nil, fmt.Errorf("unexpected reference expression type %T", n)
	}
}

func propertyName(m *ast.MemberExpression) (string, error) {
	switch p := m.Property.(type) {
	case *ast.Identifier:
		return p.Name, nil
	case *ast.StringLiteral:
		return p.Value, nil
	default:
		return "", fmt.Errorf("unsupported member property expression of type %T", m.Property)
	}
}

func CompileFn(f *ast.ArrowFunctionExpression, types map[ReferencePath]DataType) (CompiledFn, error) {
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
	cpy := make(map[ReferencePath]DataType, len(types))
	for k, v := range types {
		cpy[k] = v
	}
	return compiledFn{
		root:  root,
		types: cpy,
	}, nil
}

type Reference []string

type ReferencePath string

func (r Reference) Path() ReferencePath {
	return ReferencePath(strings.Join([]string(r), "."))
}
func (rp ReferencePath) Reference() Reference {
	return Reference(strings.Split(string(rp), "."))
}

func (r Reference) String() string {
	return string(r.Path())
}

type compiledFn struct {
	root  DataTypeEvaluator
	types map[ReferencePath]DataType
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

func (c compiledFn) Type() DataType {
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

func compile(n ast.Node, types map[ReferencePath]DataType) (DataTypeEvaluator, error) {
	switch n := n.(type) {
	case *ast.BlockStatement:
		body := make([]DataTypeEvaluator, len(n.Body))
		hasReturn := false
		var typ DataType
		for i, s := range n.Body {
			node, err := compile(s, types)
			if err != nil {
				return nil, err
			}
			body[i] = node
			if _, ok := node.(returnEvaluator); ok {
				// The returnEvaluator indicates the last statement
				hasReturn = true
				typ = node.Type()
				break
			}
		}
		if !hasReturn {
			return nil, errors.New("block has no return statement")
		}
		return &blockEvaluator{
			compiledType: compiledType(typ),
			body:         body,
		}, nil
	case *ast.ExpressionStatement:
		//TODO(nathanielc): I belive sideffects are not possible in record functions.
		// Maybe we can skip these or throw an error?
		return compile(n.Expression, types)
	case *ast.ReturnStatement:
		node, err := compile(n.Argument, types)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			DataTypeEvaluator: node,
		}, nil
	case *ast.CallExpression:
		if len(n.Arguments) != 1 {
			return nil, errors.New("call expressions only support a single object argument")
		}
		obj, ok := n.Arguments[0].(*ast.ObjectExpression)
		if !ok {
			return nil, errors.New("call expression argument must be an ObjectExpression")
		}

		//TODO(nathanielc): We need to not flatten references into reference paths so that we can accurately track
		// type information through call expressions.
		callTypes := make(map[ReferencePath]DataType, len(types)+len(obj.Properties))
		for k, v := range types {
			callTypes[k] = v
		}

		args := make(map[ReferencePath]DataTypeEvaluator, len(obj.Properties))
		for _, p := range obj.Properties {
			n, err := compile(p.Value, types)
			if err != nil {
				return nil, err
			}
			rp := ReferencePath(p.Key.Name)
			args[rp] = n
			callTypes[rp] = n.Type()
		}

		node, err := compile(n.Callee, callTypes)
		if err != nil {
			return nil, err
		}
		log.Println("callEvaluator", callTypes, node, args)

		return &callEvaluator{
			compiledType: compiledType(node.Type()),
			arguments:    args,
			fn:           node,
			callScope:    make(Scope),
		}, nil
	case *ast.ArrowFunctionExpression:
		return compile(n.Body, types)
	case *ast.VariableDeclaration:
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
	case *ast.ObjectExpression:
		meta := MapMeta{
			Properties: make([]MapPropertyMeta, len(n.Properties)),
		}
		properties := make(map[string]DataTypeEvaluator, len(n.Properties))
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
	case *ast.Identifier:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		rp := r.Path()
		return &referenceEvaluator{
			compiledType:  compiledType(types[rp]),
			referencePath: rp,
		}, nil
	case *ast.MemberExpression:
		r, err := determineReference(n)
		if err != nil {
			return nil, err
		}
		rp := r.Path()
		return &referenceEvaluator{
			compiledType:  compiledType(types[rp]),
			referencePath: rp,
		}, nil
	case *ast.BooleanLiteral:
		return &booleanEvaluator{
			compiledType: compiledType(TBool),
			b:            n.Value,
		}, nil
	case *ast.IntegerLiteral:
		return &integerEvaluator{
			compiledType: compiledType(TInt),
			i:            n.Value,
		}, nil
	case *ast.FloatLiteral:
		return &floatEvaluator{
			compiledType: compiledType(TFloat),
			f:            n.Value,
		}, nil
	case *ast.StringLiteral:
		return &stringEvaluator{
			compiledType: compiledType(TString),
			s:            n.Value,
		}, nil
	case *ast.DateTimeLiteral:
		return &timeEvaluator{
			compiledType: compiledType(TTime),
			t:            Time(n.Value.UnixNano()),
		}, nil
	case *ast.UnaryExpression:
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
	case *ast.LogicalExpression:
		l, err := compile(n.Left, types)
		if err != nil {
			return nil, err
		}
		if l.Type() != TBool {
			return nil, fmt.Errorf("invalid left operand type %v in logical ast", l.Type())
		}
		r, err := compile(n.Right, types)
		if err != nil {
			return nil, err
		}
		if r.Type() != TBool {
			return nil, fmt.Errorf("invalid right operand type %v in logical ast", r.Type())
		}
		return &logicalEvaluator{
			compiledType: compiledType(TBool),
			operator:     n.Operator,
			left:         l,
			right:        r,
		}, nil
	case *ast.BinaryExpression:
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
		return nil, fmt.Errorf("unknown ast node of type %T", n)
	}
}

type Scope map[ReferencePath]Value

func (s Scope) Type(rp ReferencePath) DataType {
	return s[rp].Type
}
func (s Scope) Set(rp ReferencePath, v Value) {
	s[rp] = v
}
func (s Scope) GetBool(rp ReferencePath) bool {
	return s[rp].Bool()
}
func (s Scope) GetInt(rp ReferencePath) int64 {
	return s[rp].Int()
}
func (s Scope) GetUInt(rp ReferencePath) uint64 {
	return s[rp].UInt()
}
func (s Scope) GetFloat(rp ReferencePath) float64 {
	return s[rp].Float()
}
func (s Scope) GetString(rp ReferencePath) string {
	return s[rp].Str()
}
func (s Scope) GetTime(rp ReferencePath) Time {
	return s[rp].Time()
}
func (s Scope) GetMap(rp ReferencePath) Map {
	return s[rp].Map()
}

type DataTypeEvaluator interface {
	Type() TypeMeta
	EvalBool(scope Scope) bool
	EvalInt(scope Scope) int64
	EvalUInt(scope Scope) uint64
	EvalFloat(scope Scope) float64
	EvalString(scope Scope) string
	EvalTime(scope Scope) Time
	EvalMap(scope Scope) Map
}

type CompiledFn interface {
	Type() TypeMeta
	Eval(scope Scope) (Value, error)
	EvalBool(scope Scope) (bool, error)
	EvalInt(scope Scope) (int64, error)
	EvalUInt(scope Scope) (uint64, error)
	EvalFloat(scope Scope) (float64, error)
	EvalString(scope Scope) (string, error)
	EvalTime(scope Scope) (Time, error)
	EvalMap(scope Scope) (Map, error)
}

type TypeMeta interface {
	DataType() DataType
	Property(name string) TypeMeta
}

func (t DataType) DataType() DataType {
	return t
}
func (t DataType) Property() TypeMeta {
	return nil
}

type MapMeta struct {
	Properties []MapPropertyMeta
}

func (mm MapMeta) DataType() DataType {
	return TMap
}
func (mm MapMeta) Property(name string) TypeMeta {
	for _, p := range mm.Properties {
		if p.Key == name {
			return p.Type
		}
	}
	return nil
}

type MapPropertyMeta struct {
	Key  string
	Type TypeMeta
}

// CompilationCache caches compilation results based on the types of the input parameters.
type CompilationCache struct {
	fn        *ast.ArrowFunctionExpression
	pathOrder []ReferencePath
	root      *compilationCacheNode
}

func NewCompilationCache(fn *ast.ArrowFunctionExpression, referencePaths []ReferencePath) *CompilationCache {
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
func (c *CompilationCache) Compile(types map[ReferencePath]DataType) (CompiledFn, error) {
	return c.root.compile(c.fn, c.pathOrder, types)
}

type compilationCacheNode struct {
	children map[DataType]*compilationCacheNode

	fn  CompiledFn
	err error
}

// compile recursively searches for a matching child node that has compiled the function.
// If the compilation has not been performed previously its result is cached and returned.
func (c *compilationCacheNode) compile(fn *ast.ArrowFunctionExpression, order []ReferencePath, types map[ReferencePath]DataType) (CompiledFn, error) {
	if len(order) == 0 {
		// We are the matching child, return the cached result or do the compilation.
		if c.fn == nil && c.err == nil {
			c.fn, c.err = CompileFn(fn, types)
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
			c.children = make(map[DataType]*compilationCacheNode)
		}
		c.children[t] = child
	}
	return child.compile(fn, order[1:], types)
}

type compiledType DataType

func (c compiledType) Type() DataType {
	return DataType(c)
}
func (c compiledType) MapMeta() MapMeta {
	panic(c.error(TMap))
}
func (c compiledType) error(exp DataType) error {
	return typeErr{Actual: DataType(c), Expected: exp}
}

type typeErr struct {
	Actual, Expected DataType
}

func (t typeErr) Error() string {
	return fmt.Sprintf("unexpected type: got %q want %q", t.Actual, t.Expected)
}

type Value struct {
	Type  DataType
	Value interface{}
}

func (v Value) error(exp DataType) error {
	return typeErr{Actual: v.Type, Expected: exp}
}

func (v Value) Bool() bool {
	return v.Value.(bool)
}
func (v Value) Int() int64 {
	return v.Value.(int64)
}
func (v Value) UInt() uint64 {
	return v.Value.(uint64)
}
func (v Value) Float() float64 {
	return v.Value.(float64)
}
func (v Value) Str() string {
	return v.Value.(string)
}
func (v Value) Time() Time {
	return v.Value.(Time)
}
func (v Value) Map() Map {
	return v.Value.(Map)
}

func eval(e DataTypeEvaluator, scope Scope) (v Value) {
	v.Type = e.Type()
	switch v.Type {
	case TBool:
		v.Value = e.EvalBool(scope)
	case TInt:
		v.Value = e.EvalInt(scope)
	case TUInt:
		v.Value = e.EvalUInt(scope)
	case TFloat:
		v.Value = e.EvalFloat(scope)
	case TString:
		v.Value = e.EvalString(scope)
	case TTime:
		v.Value = e.EvalTime(scope)
	}
	return
}

type blockEvaluator struct {
	compiledType
	body  []DataTypeEvaluator
	value Value
}

func (e *blockEvaluator) eval(scope Scope) {
	for _, b := range e.body {
		e.value = eval(b, scope)
	}
}

func (e *blockEvaluator) EvalBool(scope Scope) bool {
	if DataType(e.compiledType) != TBool {
		panic(e.error(TBool))
	}
	e.eval(scope)
	return e.value.Bool()
}

func (e *blockEvaluator) EvalInt(scope Scope) int64 {
	if DataType(e.compiledType) != TInt {
		panic(e.error(TInt))
	}
	e.eval(scope)
	return e.value.Int()
}

func (e *blockEvaluator) EvalUInt(scope Scope) uint64 {
	if DataType(e.compiledType) != TUInt {
		panic(e.error(TUInt))
	}
	e.eval(scope)
	return e.value.UInt()
}

func (e *blockEvaluator) EvalFloat(scope Scope) float64 {
	if DataType(e.compiledType) != TFloat {
		panic(e.error(TFloat))
	}
	e.eval(scope)
	return e.value.Float()
}

func (e *blockEvaluator) EvalString(scope Scope) string {
	if DataType(e.compiledType) != TString {
		panic(e.error(TString))
	}
	e.eval(scope)
	return e.value.Str()
}

func (e *blockEvaluator) EvalTime(scope Scope) Time {
	if DataType(e.compiledType) != TTime {
		panic(e.error(TTime))
	}
	e.eval(scope)
	return e.value.Time()
}
func (e *blockEvaluator) EvalMap(scope Scope) Map {
	if DataType(e.compiledType) != TMap {
		panic(e.error(TMap))
	}
	e.eval(scope)
	return e.value.Map()
}

type returnEvaluator struct {
	DataTypeEvaluator
}

type declarationEvaluator struct {
	compiledType
	id   ReferencePath
	init DataTypeEvaluator
}

func (e *declarationEvaluator) eval(scope Scope) {
	scope.Set(e.id, eval(e.init, scope))
}

func (e *declarationEvaluator) EvalBool(scope Scope) bool {
	e.eval(scope)
	return scope.GetBool(e.id)
}

func (e *declarationEvaluator) EvalInt(scope Scope) int64 {
	e.eval(scope)
	return scope.GetInt(e.id)
}

func (e *declarationEvaluator) EvalUInt(scope Scope) uint64 {
	e.eval(scope)
	return scope.GetUInt(e.id)
}

func (e *declarationEvaluator) EvalFloat(scope Scope) float64 {
	e.eval(scope)
	return scope.GetFloat(e.id)
}

func (e *declarationEvaluator) EvalString(scope Scope) string {
	e.eval(scope)
	return scope.GetString(e.id)
}

func (e *declarationEvaluator) EvalTime(scope Scope) Time {
	e.eval(scope)
	return scope.GetTime(e.id)
}

func (e *declarationEvaluator) EvalMap(scope Scope) Map {
	e.eval(scope)
	return scope.GetMap(e.id)
}

type callEvaluator struct {
	compiledType
	fn        DataTypeEvaluator
	arguments map[ReferencePath]DataTypeEvaluator
	callScope Scope
}

func (e *callEvaluator) eval(scope Scope) Value {
	for k := range e.callScope {
		delete(e.callScope, k)
	}
	for k, v := range scope {
		e.callScope.Set(k, v)
	}
	for k, node := range e.arguments {
		e.callScope.Set(k, eval(node, scope))
	}
	return eval(e.fn, e.callScope)
}

func (e *callEvaluator) EvalBool(scope Scope) bool {
	return e.eval(scope).Bool()
}

func (e *callEvaluator) EvalInt(scope Scope) int64 {
	return e.eval(scope).Int()
}

func (e *callEvaluator) EvalUInt(scope Scope) uint64 {
	return e.eval(scope).UInt()
}

func (e *callEvaluator) EvalFloat(scope Scope) float64 {
	return e.eval(scope).Float()
}

func (e *callEvaluator) EvalString(scope Scope) string {
	return e.eval(scope).Str()
}

func (e *callEvaluator) EvalTime(scope Scope) Time {
	return e.eval(scope).Time()
}

func (e *callEvaluator) EvalMap(scope Scope) Map {
	return e.eval(scope).Map()
}

type mapEvaluator struct {
	compiledType
	meta       MapMeta
	properties map[string]DataTypeEvaluator
}

func (e *mapEvaluator) MapMeta() MapMeta {
	return e.meta
}

func (e *mapEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *mapEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *mapEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *mapEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *mapEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *mapEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *mapEvaluator) EvalMap(scope Scope) Map {
	values := make(map[string]Value, len(e.properties))
	for k, node := range e.properties {
		values[k] = eval(node, scope)
	}
	return Map{
		Meta:   e.meta,
		Values: values,
	}
}

type Map struct {
	Meta   MapMeta
	Values map[string]Value
}

type logicalEvaluator struct {
	compiledType
	operator    ast.LogicalOperatorKind
	left, right DataTypeEvaluator
}

func (e *logicalEvaluator) EvalBool(scope Scope) bool {
	switch e.operator {
	case ast.AndOperator:
		return e.left.EvalBool(scope) && e.right.EvalBool(scope)
	case ast.OrOperator:
		return e.left.EvalBool(scope) || e.right.EvalBool(scope)
	default:
		panic(fmt.Errorf("unknown logical operator %v", e.operator))
	}
}

func (e *logicalEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *logicalEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *logicalEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *logicalEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *logicalEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *logicalEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type binaryFunc func(scope Scope, left, right DataTypeEvaluator) Value

type binarySignature struct {
	Operator    ast.OperatorKind
	Left, Right DataType
}

type binaryEvaluator struct {
	compiledType
	left, right DataTypeEvaluator
	f           binaryFunc
}

func (e *binaryEvaluator) EvalBool(scope Scope) bool {
	return e.f(scope, e.left, e.right).Bool()
}

func (e *binaryEvaluator) EvalInt(scope Scope) int64 {
	return e.f(scope, e.left, e.right).Int()
}

func (e *binaryEvaluator) EvalUInt(scope Scope) uint64 {
	return e.f(scope, e.left, e.right).UInt()
}

func (e *binaryEvaluator) EvalFloat(scope Scope) float64 {
	return e.f(scope, e.left, e.right).Float()
}

func (e *binaryEvaluator) EvalString(scope Scope) string {
	return e.f(scope, e.left, e.right).Str()
}

func (e *binaryEvaluator) EvalTime(scope Scope) Time {
	return e.f(scope, e.left, e.right).Time()
}
func (e *binaryEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type unaryEvaluator struct {
	compiledType
	node DataTypeEvaluator
}

func (e *unaryEvaluator) EvalBool(scope Scope) bool {
	// There is only one boolean unary operator
	return !e.node.EvalBool(scope)
}

func (e *unaryEvaluator) EvalInt(scope Scope) int64 {
	// There is only one integer unary operator
	return -e.node.EvalInt(scope)
}

func (e *unaryEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *unaryEvaluator) EvalFloat(scope Scope) float64 {
	// There is only one float unary operator
	return -e.node.EvalFloat(scope)
}

func (e *unaryEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *unaryEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *unaryEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type integerEvaluator struct {
	compiledType
	i int64
}

func (e *integerEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *integerEvaluator) EvalInt(scope Scope) int64 {
	return e.i
}

func (e *integerEvaluator) EvalUInt(scope Scope) uint64 {
	return uint64(e.i)
}

func (e *integerEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *integerEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *integerEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *integerEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type stringEvaluator struct {
	compiledType
	s string
}

func (e *stringEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *stringEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *stringEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *stringEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *stringEvaluator) EvalString(scope Scope) string {
	return e.s
}

func (e *stringEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *stringEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type booleanEvaluator struct {
	compiledType
	b bool
}

func (e *booleanEvaluator) EvalBool(scope Scope) bool {
	return e.b
}

func (e *booleanEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *booleanEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *booleanEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *booleanEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *booleanEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *booleanEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type floatEvaluator struct {
	compiledType
	f float64
}

func (e *floatEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *floatEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *floatEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *floatEvaluator) EvalFloat(scope Scope) float64 {
	return e.f
}

func (e *floatEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *floatEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *floatEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type timeEvaluator struct {
	compiledType
	t Time
}

func (e *timeEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *timeEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *timeEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *timeEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *timeEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *timeEvaluator) EvalTime(scope Scope) Time {
	return e.t
}
func (e *timeEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type referenceEvaluator struct {
	compiledType
	referencePath ReferencePath
}

func (e *referenceEvaluator) EvalBool(scope Scope) bool {
	return scope.GetBool(e.referencePath)
}

func (e *referenceEvaluator) EvalInt(scope Scope) int64 {
	return scope.GetInt(e.referencePath)
}

func (e *referenceEvaluator) EvalUInt(scope Scope) uint64 {
	return scope.GetUInt(e.referencePath)
}

func (e *referenceEvaluator) EvalFloat(scope Scope) float64 {
	return scope.GetFloat(e.referencePath)
}

func (e *referenceEvaluator) EvalString(scope Scope) string {
	return scope.GetString(e.referencePath)
}

func (e *referenceEvaluator) EvalTime(scope Scope) Time {
	return scope.GetTime(e.referencePath)
}
func (e *referenceEvaluator) EvalMap(scope Scope) Map {
	return scope.GetMap(e.referencePath)
}

// Map of binary functions
var binaryFuncs = map[binarySignature]struct {
	Func       binaryFunc
	ResultType DataType
}{
	//---------------
	// Math Operators
	//---------------
	{Operator: ast.AdditionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l + r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.AdditionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l + r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.AdditionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l + r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.SubtractionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l - r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.SubtractionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l - r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.SubtractionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l - r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.MultiplicationOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l * r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.MultiplicationOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l * r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.MultiplicationOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l * r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.DivisionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l / r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.DivisionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l / r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.DivisionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l / r,
			}
		},
		ResultType: TFloat,
	},

	//---------------------
	// Comparison Operators
	//---------------------

	// LessThanEqualOperator

	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l <= uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},

	// LessThanOperator

	{Operator: ast.LessThanOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l < uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l < float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l < float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},

	// GreaterThanEqualOperator

	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l >= uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},

	// GreaterThanOperator

	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l > uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l > float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l > float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},

	// EqualOperator

	{Operator: ast.EqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l == uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l == float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l == float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TString, Right: TString}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalString(scope)
			r := right.EvalString(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},

	// NotEqualOperator

	{Operator: ast.NotEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: l != uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l != float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l != float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
}

func FindValueKeys(f *ast.ArrowFunctionExpression) ([]string, error) {
	switch n := f.Body.(type) {
	case ast.Expression:
		if o, ok := n.(*ast.ObjectExpression); ok {
			return determineObjectKeys(o), nil
		}
		// No Key was specified, use default ValueColLabel
		return []string{DefaultValueColLabel}, nil
	case *ast.BlockStatement:
		if len(n.Body) == 0 {
			return nil, errors.New("arrow function has empty body")
		}

		//TODO(nathanielc): When we have conditional this needs to actually follow the program flow, not just skip to the last statement.
		last := n.Body[len(n.Body)-1]
		returnStmt, ok := last.(*ast.ReturnStatement)
		if !ok {
			return nil, fmt.Errorf("arrow function does not end with return statement, found %T", last)
		}
		return findValueKeys(returnStmt, n)
	default:
		return nil, fmt.Errorf("unsupported arrow body node %T", f.Body)
	}
}

// findValueKeys is a form of static analysis of the code. It searches the block for the ObjectExpression that will be returned.
func findValueKeys(n ast.Node, block *ast.BlockStatement) ([]string, error) {
	switch n := n.(type) {
	case *ast.ObjectExpression:
		return determineObjectKeys(n), nil
	case *ast.Identifier:
		// Search back for var def of the Identifier.
		for i := len(block.Body) - 1; i >= 0; i-- {
			if vd, ok := block.Body[i].(*ast.VariableDeclaration); ok {
				for _, d := range vd.Declarations {
					if d.ID.Name == n.Name {
						return findValueKeys(d.Init, block)
					}
				}
			}
		}
		return nil, fmt.Errorf("could not find identifier %q", n.Name)
	case *ast.ReturnStatement:
		return findValueKeys(n.Argument, block)
	default:
		return nil, fmt.Errorf("cannot find value keys from %T", n)
	}
}

func determineObjectKeys(m *ast.ObjectExpression) []string {
	keys := make([]string, len(m.Properties))
	for i, p := range m.Properties {
		keys[i] = p.Key.Name
	}
	return keys
}
