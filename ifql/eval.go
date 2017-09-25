package ifql

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
)

// Evaluate validates and converts an ast.Program to a query
func Evaluate(program *ast.Program) (*query.QuerySpec, error) {
	ev := evaluator{
		scope: newScope(),
	}
	// TODO: There are other possible expression/variable statements
	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			if err := ev.doVariableDeclaration(s); err != nil {
				return nil, err
			}
		case *ast.ExpressionStatement:
			value, err := ev.doExpression(s.Expression)
			if err != nil {
				return nil, err
			}
			if value.Type == TChain {
				chain := value.Value.(*CallChain)
				ev.operations = append(ev.operations, chain.Operations...)
				ev.edges = append(ev.edges, chain.Edges...)
			}
		default:
			return nil, fmt.Errorf("Unsupported program statement expression type %t", s)
		}
	}
	// TODO: There are other possible expression/variable statements
	return ev.spec(), nil
}

type CreateOperationSpec func(args map[string]Value) (query.OperationSpec, error)

var functionsMap = make(map[string]CreateOperationSpec)

func RegisterFunction(name string, c CreateOperationSpec) {
	if functionsMap[name] != nil {
		panic(fmt.Errorf("duplicate registration for function %q", name))
	}
	functionsMap[name] = c
}

// TODO(nathanielc): Maybe we want Value to be an interface instead of a struct?
// Value represents any value that can be the result of evaluating any expression.
type Value struct {
	Type  Type
	Value interface{}
}

type Type int

const (
	TString     Type = iota // Go type string
	TInt                    // Go type int64
	TFloat                  // Go type float64
	TBool                   // Go type bool
	TTime                   // Go type time.Time
	TDuration               // Go type time.Duration
	TList                   // Go type List
	TMap                    // Go type Map
	TChain                  // Go type *CallChain
	TExpression             // Go type *storage.Node TODO(nathanielc): create a type for this in this package
)

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
	case TList:
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

type List struct {
	Type Type
	// Elements will be a typed slice of any other type
	// []string, []float64, or possibly even []List and []Map
	Elements interface{}
}
type Map struct {
	Type Type
	// Elements will be a typed map of any other type, keys are always strings
	// map[strin]string, map[string]float64, or possibly even map[string]List and map[string]Map
	Elements interface{}
}

type evaluator struct {
	id    int
	scope *scope

	operations []*query.Operation
	edges      []query.Edge
}

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

func (ev *evaluator) doVariableDeclaration(declarations *ast.VariableDeclaration) error {
	for _, vd := range declarations.Declarations {
		value, err := ev.doExpression(vd.Init)
		if err != nil {
			return err
		}
		ev.scope.Set(vd.ID.Name, value)
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
		root, err := ev.comparisonNode(e)
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type:  TExpression,
			Value: root,
		}, nil
	case *ast.LogicalExpression:
		root, err := ev.logicalNode(e)
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type:  TExpression,
			Value: root,
		}, nil
	default:
		return Value{}, fmt.Errorf("unsupported expression %T", expr)
	}
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
		op, err := ev.function(callee.Name, call.Arguments)
		if err != nil {
			return nil, err
		}
		return &CallChain{
			Operations: []*query.Operation{op},
			Parent:     op.ID,
		}, nil
	case *ast.MemberExpression:
		chain, name, err := ev.memberFunction(callee, chain)
		if err != nil {
			return nil, err
		}

		op, err := ev.function(name, call.Arguments)
		if err != nil {
			return nil, err
		}

		chain.Operations = append(chain.Operations, op)
		chain.Edges = append(chain.Edges, query.Edge{
			Parent: chain.Parent,
			Child:  op.ID,
		})
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
		return nil, "", fmt.Errorf("Variables not support yet in member expression object type")
	default:
		return nil, "", fmt.Errorf("Unsupported member expression object type %t", obj)
	}

	return chain, member.Property.Name, nil
}

func (ev *evaluator) function(name string, args []ast.Expression) (*query.Operation, error) {
	op := &query.Operation{
		ID: query.OperationID(fmt.Sprintf("%s%d", name, ev.nextID())),
	}
	createOpSpec, ok := functionsMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown function %q", name)
	}
	var paramMap map[string]Value
	if len(args) == 1 {
		params, ok := args[0].(*ast.ObjectExpression)
		if !ok {
			return nil, fmt.Errorf("arguments not a valid object expression")
		}
		var err error
		paramMap, err = ev.resolveParameters(params)
		if err != nil {
			return nil, err
		}
	}
	spec, err := createOpSpec(paramMap)
	if err != nil {
		return nil, err
	}
	op.Spec = spec
	return op, nil
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
			Relative: v,
		}, nil
	case int64:
		return query.Time{
			Absolute: time.Unix(v, 0),
		}, nil
	default:
		return query.Time{}, fmt.Errorf("unknown time type %t", value.Value)
	}
}

func (ev *evaluator) binaryOperation(expr ast.Expression) (*storage.Node, error) {
	switch op := expr.(type) {
	case *ast.BinaryExpression:
		return ev.comparisonNode(op)
	case *ast.LogicalExpression:
		return ev.logicalNode(op)
	default:
		return nil, fmt.Errorf("Expression type expected to be arithmatic, relational or logical, but is %t", op)
	}
}

func (ev *evaluator) comparisonNode(expr *ast.BinaryExpression) (*storage.Node, error) {
	lhs, err := primaryNode(expr.Left.(ast.Literal), true /* isLeft */)
	if err != nil {
		return nil, err
	}

	rhs, err := primaryNode(expr.Right.(ast.Literal), false /* isLeft */)
	if err != nil {
		return nil, err
	}

	isRegex := rhs.GetRegexValue() != "" || lhs.GetRegexValue() != ""
	op, err := comparisonOp(expr.Operator, isRegex)
	if err != nil {
		return nil, err
	}

	return &storage.Node{
		NodeType: storage.NodeTypeComparisonExpression,
		Value:    &storage.Node_Comparison_{Comparison: op},
		Children: []*storage.Node{lhs, rhs},
	}, nil
}

func comparisonOp(op ast.OperatorKind, isRegex bool) (storage.Node_Comparison, error) {
	switch op {
	case ast.EqualOperator:
		if isRegex {
			return storage.ComparisonRegex, nil
		}
		return storage.ComparisonEqual, nil
	case ast.NotEqualOperator:
		if isRegex {
			return storage.ComparisonNotRegex, nil
		}
		return storage.ComparisonNotEqual, nil
	case ast.StartsWithOperator:
		return storage.ComparisonStartsWith, nil
	case ast.MultiplicationOperator, ast.DivisionOperator, ast.AdditionOperator, ast.SubtractionOperator:
		fallthrough
	case ast.LessThanEqualOperator, ast.LessThanOperator, ast.GreaterThanEqualOperator, ast.GreaterThanOperator:
		fallthrough
	case ast.InOperator, ast.NotEmptyOperator:
		return 0, fmt.Errorf("Unimplemented comparison operator %s", op)
	default:
		return 0, fmt.Errorf("Unknown comparison operator %s", op)
	}
}

func primaryNode(expr ast.Literal, isLeft bool) (*storage.Node, error) {
	switch lit := expr.(type) {
	case *ast.StringLiteral:
		if isLeft {
			return &storage.Node{
				NodeType: storage.NodeTypeTagRef,
				Value: &storage.Node_TagRefValue{
					TagRefValue: lit.Value,
				},
			}, nil
		}
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: lit.Value,
			},
		}, nil

	case *ast.NumberLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: lit.Value,
			},
		}, nil
	case *ast.IntegerLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_IntegerValue{
				IntegerValue: lit.Value,
			},
		}, nil
	case *ast.RegexpLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: lit.Value.String(),
			},
		}, nil
	case *ast.FieldLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeTagRef,
			Value: &storage.Node_TagRefValue{
				TagRefValue: lit.Value,
			},
		}, nil
	case *ast.DurationLiteral, *ast.DateTimeLiteral, *ast.BooleanLiteral:
		return nil, fmt.Errorf("durations, datetimes and booleans not yet support in expressions")
	default:
		return nil, fmt.Errorf("unknown primary literal type: %T", expr)
	}
}

func (ev *evaluator) logicalNode(expr *ast.LogicalExpression) (*storage.Node, error) {
	lhs, err := ev.binaryOperation(expr.Left)
	if err != nil {
		return nil, err
	}

	rhs, err := ev.binaryOperation(expr.Right)
	if err != nil {
		return nil, err
	}

	value := func(op ast.LogicalOperatorKind) *storage.Node_Logical_ {
		if op == ast.AndOperator {
			return &storage.Node_Logical_{Logical: storage.LogicalAnd}
		}
		return &storage.Node_Logical_{Logical: storage.LogicalOr}
	}

	return &storage.Node{
		NodeType: storage.NodeTypeLogicalExpression,
		Value:    value(expr.Operator),
		Children: []*storage.Node{lhs, rhs},
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
