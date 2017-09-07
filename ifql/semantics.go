package ifql

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
)

// NewQuerySpec validates and converts an ast.Program to a query
func NewQuerySpec(program *ast.Program) (*query.QuerySpec, error) {
	// TODO: There are other possible expression/variable statements
	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			return nil, fmt.Errorf("Variables not support yet")
		case *ast.ExpressionStatement:
			return fromExpression(s.Expression)
		default:
			return nil, fmt.Errorf("Unsupported program statement expression type %t", s)
		}
	}
	// TODO: There are other possible expression/variable statements
	return nil, nil
}

func fromExpression(expr ast.Expression) (*query.QuerySpec, error) {
	switch e := expr.(type) {
	case *ast.CallExpression:
		fn, err := callFunction(e, nil)
		if err != nil {
			return nil, err
		}
		return fn.QuerySpec(), nil
	default:
		return nil, fmt.Errorf("Unsupport expression %t", e)
	}
}

// CallChain is an intermediate structure to build QuerySpecs during recursion through the AST
type CallChain struct {
	Operators  []*query.Operation
	Edges      []query.Edge
	ParentName string
	ParentID   string
}

// QuerySpec converts the CallChain into a query.QuerySpec
func (c *CallChain) QuerySpec() *query.QuerySpec {
	return &query.QuerySpec{
		Operations: c.Operators,
		Edges:      c.Edges,
	}
}

func callFunction(call *ast.CallExpression, chain *CallChain) (*CallChain, error) {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		op, err := operation(callee.Name, call.Arguments)
		if err != nil {
			return nil, err
		}
		return &CallChain{
			Operators:  []*query.Operation{op},
			Edges:      []query.Edge{},
			ParentName: callee.Name,
			ParentID:   callee.Name, // TODO make this a real ID
		}, nil
	case *ast.MemberExpression:
		chain, err := memberFunction(callee, chain)
		if err != nil {
			return nil, err
		}

		op, err := operation(chain.ParentName, call.Arguments)
		if err != nil {
			return nil, err
		}

		chain.Operators = append(chain.Operators, op)
		return chain, nil
	default:
		return nil, fmt.Errorf("Unsupported callee expression type %t", callee)
	}
}

func memberFunction(member *ast.MemberExpression, chain *CallChain) (*CallChain, error) {
	chain, err := callFunction(member.Object, chain)
	if err != nil {
		return nil, err
	}

	child := member.Property.Name
	// TODO: make these IDs uniquer-er-er
	childID := child

	edge := query.Edge{
		Parent: query.OperationID(chain.ParentID),
		Child:  query.OperationID(childID),
	}

	chain.Edges = append(chain.Edges, edge)
	chain.ParentID = childID
	chain.ParentName = child
	return chain, nil
}

func operation(name string, args []ast.Expression) (*query.Operation, error) {
	switch strings.ToLower(name) {
	case "select":
		return selectOperation(args)
	case "sum":
		return sumOperation(args)
	case "range":
		return rangeOperation(args)
	case "count":
		return countOperation(args)
	case "clear":
		return clearOperation(args)
	case "where":
		return whereOperation(args)
	case "window":
		return windowOperation(args)
	default:
		return nil, fmt.Errorf("Unknown function %s", name)
	}
}

func selectOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(`select operation requires argument "database"`)
	}

	params, ok := args[0].(*ast.ObjectExpression)
	if !ok {
		return nil, fmt.Errorf("Arguments not a valid object expression")
	}

	if len(params.Properties) != 1 {
		return nil, fmt.Errorf(`select operation requires argument "database"`)
	}

	arg := params.Properties[0]
	param := strings.ToLower(arg.Key.Name)
	if param != "database" && param != "db" {
		return nil, fmt.Errorf("Argument is not database: %s", param)
	}

	db, ok := arg.Value.(*ast.StringLiteral)
	if !ok {
		return nil, fmt.Errorf("Argument to database parameter is not a string but is %t", arg.Value)
	}

	return &query.Operation{
		ID: "select", // TODO: Change this to a UUID
		Spec: &query.SelectOpSpec{
			Database: db.Value,
		},
	}, nil
}

func sumOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`sum operation requires no arguments`)
	}

	return &query.Operation{
		ID:   "sum",
		Spec: &query.SumOpSpec{},
	}, nil
}

func countOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`count operation requires no arguments`)
	}

	return &query.Operation{
		ID:   "count",
		Spec: &query.CountOpSpec{},
	}, nil
}

func clearOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`clear operation requires no arguments`)
	}

	return &query.Operation{
		ID:   "clear",
		Spec: &query.ClearOpSpec{},
	}, nil
}

func rangeOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(`range operation requires the argument "start"`)
	}

	params, ok := args[0].(*ast.ObjectExpression)
	if !ok {
		return nil, fmt.Errorf("Arguments not a valid object expression")
	}

	if len(params.Properties) != 1 && len(params.Properties) != 2 {
		return nil, fmt.Errorf(`range operation requires argument "start" with optional argument "stop"`)
	}

	times := map[string]query.Time{}
	for _, p := range params.Properties {
		fn := strings.ToLower(p.Key.Name)
		if fn != "start" && fn != "stop" {
			return nil, fmt.Errorf(`range operation requires argument "start" with optional argument "stop"`)
		}

		tm, err := queryTime(p.Value.(ast.Literal))
		if err != nil {
			return nil, err
		}

		times[fn] = tm
	}

	start, ok := times["start"]
	if !ok {
		return nil, fmt.Errorf(`range operation requires the argument "start"`)
	}

	stop, ok := times["stop"]
	if !ok {
		stop = query.Time{
			Absolute: time.Now(),
		}
	}

	return &query.Operation{
		ID: "range", // TODO: Change this to a UUID
		Spec: &query.RangeOpSpec{
			Start: start,
			Stop:  stop,
		},
	}, nil
}

func queryTime(arg ast.Literal) (query.Time, error) {
	switch t := arg.(type) {
	case *ast.DateTimeLiteral:
		return query.Time{
			Absolute: t.Value,
		}, nil
	case *ast.DurationLiteral:
		return query.Time{
			Relative: t.Value,
		}, nil
	case *ast.NumberLiteral:
		// TODO:  Should this be nano?
		// FIXME: Should this be an integer?
		return query.Time{
			Absolute: time.Unix(int64(t.Value), 0),
		}, nil
	default:
		return query.Time{}, fmt.Errorf("Unknown time type")
	}
}

func whereOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(`where operation requires argument "exp"`)
	}

	params, ok := args[0].(*ast.ObjectExpression)
	if !ok {
		return nil, fmt.Errorf("Arguments not a valid object expression")
	}

	if len(params.Properties) != 1 {
		return nil, fmt.Errorf(`where operation requires argument "exp"`)
	}

	arg := params.Properties[0]
	param := strings.ToLower(arg.Key.Name)
	if param != "exp" {
		return nil, fmt.Errorf(`where operation requires argument "exp", but has %s`, param)
	}

	expr, ok := arg.Value.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf(`Argument to "exp" parameter is not an expression but is %t`, arg.Value)
	}

	root, err := binaryOperation(expr)
	if err != nil {
		return nil, err
	}

	return &query.Operation{
		ID: "where", // TODO: Change this to a UUID
		Spec: &query.WhereOpSpec{
			Exp: &query.WhereExpressionSpec{
				Predicate: &storage.Predicate{
					Root: root,
				},
			},
		},
	}, nil
}

func binaryOperation(expr ast.Expression) (*storage.Node, error) {
	switch op := expr.(type) {
	case *ast.BinaryExpression:
		return comparisonNode(op)
	case *ast.LogicalExpression:
		return logicalNode(op)
	default:
		return nil, fmt.Errorf("Expression type expected to be arithmatic, relational or logical, but is %t", op)
	}
}

func comparisonNode(expr *ast.BinaryExpression) (*storage.Node, error) {
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
		NodeType: storage.NodeTypeBooleanExpression,
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
				NodeType: storage.NodeTypeRef,
				Value: &storage.Node_RefValue{
					RefValue: lit.Value,
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
	case *ast.RegexpLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: lit.Value.String(),
			},
		}, nil
	case *ast.FieldLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeRef,
			Value: &storage.Node_RefValue{
				RefValue: lit.Value,
			},
		}, nil
	case *ast.DurationLiteral, *ast.DateTimeLiteral, *ast.BooleanLiteral:
		return nil, fmt.Errorf("durations, datetimes and booleans not yet support in expressions")
	default:
		return nil, fmt.Errorf("Unknown primary literal type: %t", lit)
	}
}

func logicalNode(expr *ast.LogicalExpression) (*storage.Node, error) {
	lhs, err := binaryOperation(expr.Left)
	if err != nil {
		return nil, err
	}

	rhs, err := binaryOperation(expr.Right)
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
		NodeType: storage.NodeTypeGroupExpression,
		Value:    value(expr.Operator),
		Children: []*storage.Node{lhs, rhs},
	}, nil
}

func windowOperation(args []ast.Expression) (*query.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(`window operation requires arguments "every" and "period" with optional arguments "start" and "round"`)
	}

	params, ok := args[0].(*ast.ObjectExpression)
	if !ok {
		return nil, fmt.Errorf("Arguments not a valid object expression")
	}

	if len(params.Properties) < 2 && len(params.Properties) > 4 {
		return nil, fmt.Errorf(`window operation requires arguments "every" and "period" with optional arguments "start" and "round"`)
	}

	spec := &query.WindowOpSpec{}
	everySet := false
	periodSet := false
	for _, farg := range params.Properties {
		name := farg.Key.Name
		arg := farg.Value
		switch name {
		case "every":
			everySet = true
			dur, ok := arg.(*ast.DurationLiteral)
			if !ok {
				return nil, fmt.Errorf("every argument must be a duration")
			}
			spec.Every = query.Duration(dur.Value)
		case "period":
			periodSet = true
			dur, ok := arg.(*ast.DurationLiteral)
			if !ok {
				return nil, fmt.Errorf("period argument must be a duration")
			}
			spec.Period = query.Duration(dur.Value)
		case "round":
			dur, ok := arg.(*ast.DurationLiteral)
			if !ok {
				return nil, fmt.Errorf("round argument must be a duration")
			}
			spec.Round = query.Duration(dur.Value)
		case "start":
			t, err := queryTime(arg.(ast.Literal))
			if err != nil {
				return nil, err
			}
			spec.Start = t
		}
	}
	// Apply defaults
	if !everySet {
		spec.Every = spec.Period
	}
	if !periodSet {
		spec.Period = spec.Every
	}
	return &query.Operation{
		ID:   "window", // TODO: Change this to a UUID
		Spec: spec,
	}, nil
}
