package ifql

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query"
)

// Evaluate validates and converts an ast.Program to a query
func Evaluate(program *ast.Program) (*query.QuerySpec, error) {
	ev := evaluator{
		scope: newScope(),
	}
	err := ev.eval(program)
	if err != nil {
		return nil, err
	}
	return ev.spec(), nil
}

type Context interface {
	// LookupIDFromIdentifier returns the operation ID of an existing operation for a given identifier.
	LookupIDFromIdentifier(string) (query.OperationID, error)
	// AdditionalParent indicates that additional parents IDs should be added to this operation.
	AdditionalParent(id query.OperationID)
}

type CreateOperationSpec func(args map[string]Value, ctx Context) (query.OperationSpec, error)

var functionsMap = make(map[string]CreateOperationSpec)

func RegisterFunction(name string, c CreateOperationSpec) {
	if functionsMap[name] != nil {
		panic(fmt.Errorf("duplicate registration for function %q", name))
	}
	functionsMap[name] = c
}

type evaluator struct {
	id    int
	scope *scope

	operations []*query.Operation
	edges      []query.Edge
}

func (ev *evaluator) eval(program *ast.Program) error {
	// TODO: There are other possible expression/variable statements
	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			if err := ev.doVariableDeclaration(s); err != nil {
				return err
			}
		case *ast.ExpressionStatement:
			value, err := ev.doExpression(s.Expression)
			if err != nil {
				return err
			}
			if value.Type == TChain {
				chain := value.Value.(*CallChain)
				ev.addChain(chain)
			}
		default:
			return fmt.Errorf("unsupported program statement expression type %t", s)
		}
	}
	return nil
}

// TODO: There are other possible expression/variable statements
func (ev *evaluator) spec() *query.QuerySpec {
	return &query.QuerySpec{
		Operations: ev.operations,
		Edges:      ev.edges,
	}
}

func (ev *evaluator) nextID() int {
	id := ev.id
	ev.id++
	return id
}

func (ev *evaluator) addChain(chain *CallChain) {
	ev.operations = append(ev.operations, chain.Operations...)
	ev.edges = append(ev.edges, chain.Edges...)
	chain.Operations = chain.Operations[0:0]
	chain.Edges = chain.Edges[0:0]
}

func (ev *evaluator) doVariableDeclaration(declarations *ast.VariableDeclaration) error {
	for _, vd := range declarations.Declarations {
		value, err := ev.doExpression(vd.Init)
		if err != nil {
			return err
		}
		ev.scope.Set(vd.ID.Name, value)
		if value.Type == TChain {
			chain := value.Value.(*CallChain)
			ev.addChain(chain)
		}
	}
	return nil
}

func (ev *evaluator) doExpression(expr ast.Expression) (Value, error) {
	switch e := expr.(type) {
	case ast.Literal:
		return ev.doLiteral(e)
	case *ast.Identifier:
		value, ok := ev.scope.Get(e.Name)
		if !ok {
			return Value{}, fmt.Errorf("undefined identifier %q", e.Name)
		}
		return value, nil
	case *ast.CallExpression:
		chain, err := ev.callFunction(e, nil)
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type:  TChain,
			Value: chain,
		}, nil
	case *ast.BinaryExpression:
		root, err := ev.binaryExpression(e)
		if err != nil {
			return Value{}, err
		}
		// TODO(nathanielc): Attempt to resolve the binary expression
		return Value{
			Type:  TExpression,
			Value: root,
		}, nil
	case *ast.LogicalExpression:
		root, err := ev.logicalExpression(e)
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type:  TExpression,
			Value: root,
		}, nil
	case *ast.FunctionExpression:
		return ev.doExpression(e.Function)
	case *ast.ArrayExpression:
		return ev.doArray(e)
	default:
		return Value{}, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (ev *evaluator) doArray(a *ast.ArrayExpression) (Value, error) {
	array := Array{
		Type: TInvalid,
	}
	elements := make([]Value, len(a.Elements))
	for i, el := range a.Elements {
		v, err := ev.doExpression(el)
		if err != nil {
			return Value{}, err
		}
		if array.Type == TInvalid {
			array.Type = v.Type
		}
		if array.Type != v.Type {
			return Value{}, fmt.Errorf("cannot mix types in an array, found both %v and %v", array.Type, v.Type)
		}
		elements[i] = v
	}
	switch array.Type {
	case TString:
		value := make([]string, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(string)
		}
		array.Elements = value
	case TInt:
		value := make([]int64, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(int64)
		}
		array.Elements = value
	case TFloat:
		value := make([]float64, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(float64)
		}
		array.Elements = value
	case TBool:
		value := make([]bool, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(bool)
		}
		array.Elements = value
	case TTime:
		value := make([]time.Time, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(time.Time)
		}
		array.Elements = value
	case TDuration:
		value := make([]time.Duration, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(time.Duration)
		}
		array.Elements = value
	case TArray:
		value := make([]Array, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(Array)
		}
		array.Elements = value
	case TMap:
		value := make([]Map, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(Map)
		}
		array.Elements = value
	case TChain:
		value := make([]*CallChain, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(*CallChain)
		}
		array.Elements = value
	case TExpression:
		value := make([]expression.Node, len(elements))
		for i, el := range elements {
			value[i] = el.Value.(expression.Node)
		}
		array.Elements = value
	default:
		return Value{}, fmt.Errorf("cannot define an array with elements of type %v", array.Type)
	}
	return Value{
		Type:  TArray,
		Value: array,
	}, nil
}

func (ev *evaluator) doLiteral(lit ast.Literal) (Value, error) {
	switch l := lit.(type) {
	case *ast.DateTimeLiteral:
		return Value{
			Type:  TTime,
			Value: l.Value,
		}, nil
	case *ast.DurationLiteral:
		return Value{
			Type:  TDuration,
			Value: l.Value,
		}, nil
	case *ast.NumberLiteral:
		return Value{
			Type:  TFloat,
			Value: l.Value,
		}, nil
	case *ast.IntegerLiteral:
		return Value{
			Type:  TInt,
			Value: l.Value,
		}, nil
	case *ast.StringLiteral:
		return Value{
			Type:  TString,
			Value: l.Value,
		}, nil
	case *ast.BooleanLiteral:
		return Value{
			Type:  TBool,
			Value: l.Value,
		}, nil
	// TODO(nathanielc): Support lists and maps
	default:
		return Value{}, fmt.Errorf("unknown literal type %T", lit)
	}

}

// CallChain is an intermediate structure to build QuerySpecs during recursion through the AST
type CallChain struct {
	Operations []*query.Operation
	Edges      []query.Edge
	Parent     query.OperationID
}

func (ev *evaluator) callFunction(call *ast.CallExpression, chain *CallChain) (*CallChain, error) {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		op, parents, err := ev.function(callee.Name, call.Arguments)
		if err != nil {
			return nil, err
		}
		chain := &CallChain{
			Operations: []*query.Operation{op},
			Parent:     op.ID,
		}
		// Add any additional parents
		for _, p := range parents {
			chain.Edges = append(chain.Edges, query.Edge{
				Parent: p,
				Child:  op.ID,
			})
		}
		return chain, nil
	case *ast.MemberExpression:
		chain, name, err := ev.memberFunction(callee, chain)
		if err != nil {
			return nil, err
		}

		op, parents, err := ev.function(name, call.Arguments)
		if err != nil {
			return nil, err
		}

		// Update chain
		chain.Operations = append(chain.Operations, op)
		chain.Edges = append(chain.Edges, query.Edge{
			Parent: chain.Parent,
			Child:  op.ID,
		})

		// Add any additional parents
		for _, p := range parents {
			if p != chain.Parent {
				chain.Edges = append(chain.Edges, query.Edge{
					Parent: p,
					Child:  op.ID,
				})
			}
		}

		// Update ParentID
		chain.Parent = op.ID
		return chain, nil
	default:
		return nil, fmt.Errorf("Unsupported callee expression type %t", callee)
	}
}

func (ev *evaluator) memberFunction(member *ast.MemberExpression, chain *CallChain) (*CallChain, string, error) {
	switch obj := member.Object.(type) {
	case *ast.CallExpression:
		var err error
		chain, err = ev.callFunction(obj, chain)
		if err != nil {
			return nil, "", err
		}
	case *ast.Identifier:
		value, ok := ev.scope.Get(obj.Name)
		if !ok {
			return nil, "", fmt.Errorf("undefined identifier %q", obj.Name)
		}
		if value.Type != TChain {
			return nil, "", fmt.Errorf("variable %q is not a function chain", obj.Name)
		}
		// Create a copy of the chain since we do not want to mutate the version stored in the scope.
		if chain == nil {
			chain = new(CallChain)
		}
		*chain = *(value.Value.(*CallChain))
	default:
		return nil, "", fmt.Errorf("unsupported member expression object type %t", obj)
	}

	return chain, member.Property.Name, nil
}

func (ev *evaluator) function(name string, args []ast.Expression) (*query.Operation, []query.OperationID, error) {
	op := &query.Operation{
		ID: query.OperationID(fmt.Sprintf("%s%d", name, ev.nextID())),
	}
	createOpSpec, ok := functionsMap[name]
	if !ok {
		return nil, nil, fmt.Errorf("unknown function %q", name)
	}
	var paramMap map[string]Value
	if len(args) == 1 {
		params, ok := args[0].(*ast.ObjectExpression)
		if !ok {
			return nil, nil, fmt.Errorf("arguments not a valid object expression")
		}
		var err error
		paramMap, err = ev.resolveParameters(params)
		if err != nil {
			return nil, nil, err
		}
	}
	ctx := &context{scope: ev.scope}
	spec, err := createOpSpec(paramMap, ctx)
	if err != nil {
		return nil, nil, err
	}
	op.Spec = spec
	return op, ctx.parents, nil
}

func (ev *evaluator) resolveParameters(params *ast.ObjectExpression) (map[string]Value, error) {
	paramsMap := make(map[string]Value, len(params.Properties))
	for _, p := range params.Properties {
		value, err := ev.doExpression(p.Value)
		if err != nil {
			return nil, err
		}
		paramsMap[p.Key.Name] = value
	}
	return paramsMap, nil
}

func ToQueryTime(value Value) (query.Time, error) {
	switch v := value.Value.(type) {
	case time.Time:
		return query.Time{
			Absolute: v,
		}, nil
	case time.Duration:
		return query.Time{
			Relative:   v,
			IsRelative: true,
		}, nil
	case int64:
		return query.Time{
			Absolute: time.Unix(v, 0),
		}, nil
	default:
		return query.Time{}, fmt.Errorf("unknown time type %t", value.Value)
	}
}

func (ev *evaluator) binaryOperation(expr ast.Expression) (expression.Node, error) {
	switch op := expr.(type) {
	case *ast.BinaryExpression:
		return ev.binaryExpression(op)
	case *ast.LogicalExpression:
		return ev.logicalExpression(op)
	default:
		return nil, fmt.Errorf("Expression type expected to be arithmatic, relational or logical, but is %t", op)
	}
}

func (ev *evaluator) binaryExpression(expr *ast.BinaryExpression) (expression.Node, error) {
	lhs, err := ev.primaryNode(expr.Left, true /* isLeft */)
	if err != nil {
		return nil, err
	}

	rhs, err := ev.primaryNode(expr.Right, false /* isLeft */)
	if err != nil {
		return nil, err
	}

	isRegexp := lhs.Type() == expression.RegexpLiteral || rhs.Type() == expression.RegexpLiteral
	op, err := expressionOperator(expr.Operator, isRegexp)
	if err != nil {
		return nil, err
	}

	return &expression.BinaryNode{
		Operator: op,
		Left:     lhs,
		Right:    rhs,
	}, nil
}

func expressionOperator(op ast.OperatorKind, isRegexp bool) (expression.Operator, error) {
	switch op {
	case ast.EqualOperator:
		if isRegexp {
			return expression.RegexpMatchOperator, nil
		}
		return expression.EqualOperator, nil
	case ast.NotEqualOperator:
		if isRegexp {
			return expression.RegexpNotMatchOperator, nil
		}
		return expression.NotEqualOperator, nil
	case ast.StartsWithOperator:
		return expression.StartsWithOperator, nil
	case ast.MultiplicationOperator:
		return expression.MultiplicationOperator, nil
	case ast.DivisionOperator:
		return expression.DivisionOperator, nil
	case ast.AdditionOperator:
		return expression.AdditionOperator, nil
	case ast.SubtractionOperator:
		return expression.SubtractionOperator, nil
	case ast.LessThanEqualOperator:
		return expression.LessThanEqualOperator, nil
	case ast.LessThanOperator:
		return expression.LessThanOperator, nil
	case ast.GreaterThanEqualOperator:
		return expression.GreaterThanEqualOperator, nil
	case ast.GreaterThanOperator:
		return expression.GreaterThanOperator, nil
	case ast.InOperator:
		return expression.InOperator, nil
	case ast.NotEmptyOperator:
		return expression.NotEmptyOperator, nil
	default:
		return 0, fmt.Errorf("unknown operator %s", op)
	}
}

func logicalOperator(op ast.LogicalOperatorKind) (expression.Operator, error) {
	switch op {
	case ast.AndOperator:
		return expression.AndOperator, nil
	case ast.OrOperator:
		return expression.OrOperator, nil
	default:
		return 0, fmt.Errorf("unknown logical operator %s", op)
	}
}

const (
	tagRefKind   = "tag"
	fieldRefKind = "field"
	identRefKind = "identifier"
)

func (ev *evaluator) primaryNode(expr ast.Expression, isLeft bool) (expression.Node, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return ev.binaryExpression(e)
	case *ast.StringLiteral:
		if isLeft {
			return &expression.ReferenceNode{
				Name: e.Value,
				Kind: tagRefKind,
			}, nil
		}
		return &expression.StringLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.BooleanLiteral:
		return &expression.BooleanLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.NumberLiteral:
		return &expression.FloatLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.IntegerLiteral:
		return &expression.IntegerLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.DurationLiteral:
		return &expression.DurationLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.DateTimeLiteral:
		return &expression.TimeLiteralNode{
			Value: e.Value,
		}, nil
	case *ast.RegexpLiteral:
		return &expression.RegexpLiteralNode{
			Value: e.Value.String(),
		}, nil
	case *ast.FieldLiteral:
		return &expression.ReferenceNode{
			Name: e.Value,
			Kind: fieldRefKind,
		}, nil
	case *ast.Identifier:
		return &expression.ReferenceNode{
			Name: e.Name,
			Kind: identRefKind,
		}, nil
	default:
		return nil, fmt.Errorf("unknown primary type: %T", expr)
	}
}

func (ev *evaluator) logicalExpression(expr *ast.LogicalExpression) (expression.Node, error) {
	lhs, err := ev.binaryOperation(expr.Left)
	if err != nil {
		return nil, err
	}

	rhs, err := ev.binaryOperation(expr.Right)
	if err != nil {
		return nil, err
	}
	op, err := logicalOperator(expr.Operator)
	if err != nil {
		return nil, err
	}

	return &expression.BinaryNode{
		Operator: op,
		Left:     lhs,
		Right:    rhs,
	}, nil
}

type scope struct {
	vars map[string]Value
}

func newScope() *scope {
	return &scope{
		vars: make(map[string]Value),
	}
}

func (s *scope) Get(name string) (Value, bool) {
	v, ok := s.vars[name]
	return v, ok
}
func (s *scope) Set(name string, value Value) {
	s.vars[name] = value
}

// TODO(nathanielc): Maybe we want Value to be an interface instead of a struct?
// Value represents any value that can be the result of evaluating any expression.
type Value struct {
	Type  Type
	Value interface{}
}

// Type represents the supported types within IFQL
type Type int

const (
	TInvalid    Type = iota // Go type nil
	TString                 // Go type string
	TInt                    // Go type int64
	TFloat                  // Go type float64
	TBool                   // Go type bool
	TTime                   // Go type time.Time
	TDuration               // Go type time.Duration
	TArray                  // Go type Array
	TMap                    // Go type Map
	TChain                  // Go type *CallChain
	TExpression             // Go type expression.Node
)

// String converts Type into a string representation of the type's name
func (t Type) String() string {
	switch t {
	case TString:
		return "string"
	case TInt:
		return "int"
	case TFloat:
		return "float"
	case TBool:
		return "bool"
	case TTime:
		return "time"
	case TDuration:
		return "duration"
	case TArray:
		return "list"
	case TMap:
		return "map"
	case TChain:
		return "chain"
	case TExpression:
		return "expression"
	default:
		return fmt.Sprintf("unknown type %d", int(t))
	}
}

// Array represents an IFQL sequence of elements
type Array struct {
	Type Type
	// Elements will be a typed slice of any other type
	// []string, []float64, or possibly even []Array and []Map
	Elements interface{}
}

// Map represents an IFQL association of keys to values of Type
type Map struct {
	Type Type
	// Elements will be a typed map of any other type, keys are always strings
	// map[string]string, map[string]float64, or possibly even map[string]Array and map[string]Map
	Elements interface{}
}

type context struct {
	scope   *scope
	parents []query.OperationID
}

func (c *context) LookupIDFromIdentifier(ident string) (id query.OperationID, err error) {
	v, ok := c.scope.Get(ident)
	if !ok {
		err = fmt.Errorf("unknown identifier %q", ident)
		return
	}
	if v.Type != TChain {
		err = fmt.Errorf("identifier not a function chain %q, got %v", ident, v.Type)
		return
	}
	id = v.Value.(*CallChain).Parent
	return
}
func (c *context) AdditionalParent(id query.OperationID) {
	for _, p := range c.parents {
		if p == id {
			return
		}
	}
	c.parents = append(c.parents, id)
}
