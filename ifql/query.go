package ifql

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
)

func NewQuery(ifql string, opts ...Option) (*query.QuerySpec, error) {
	function, err := NewAST(ifql, opts...)
	if err != nil {
		return nil, err
	}
	ops, edges, err := NewOperations(function, "")
	if err != nil {
		return nil, err
	}
	return &query.QuerySpec{
		Operations: ops,
		Edges:      edges,
	}, nil
}

func NewOperations(function *Function, parent string) ([]*query.Operation, []query.Edge, error) {
	ops := []*query.Operation{}
	edges := []query.Edge{}
	op, me, err := NewOperation(function, parent)
	if err != nil {
		return nil, nil, err
	}
	ops = append(ops, op)
	if parent != "" {
		edges = append(edges, me)
	}

	for _, f := range function.Children {
		o, children, err := NewOperations(f, string(me.Child))
		if err != nil {
			return nil, nil, err
		}

		ops = append(ops, o...)
		edges = append(edges, children...)
	}
	return ops, edges, nil
}

func NewOperation(function *Function, parent string) (*query.Operation, query.Edge, error) {
	edge := query.Edge{
		Parent: query.OperationID(parent),
	}

	switch function.Name {
	case "select":
		edge.Child = query.OperationID(function.Name)
		op, err := NewSelectOperation(function.Args)
		return op, edge, err
	case "sum":
		edge.Child = query.OperationID(function.Name)
		op, err := NewSumOperation(function.Args)
		return op, edge, err
	case "range":
		edge.Child = query.OperationID(function.Name)
		op, err := NewRangeOperation(function.Args)
		return op, edge, err
	case "count":
		edge.Child = query.OperationID(function.Name)
		op, err := NewCountOperation(function.Args)
		return op, edge, err
	case "clear":
		edge.Child = query.OperationID(function.Name)
		op, err := NewClearOperation(function.Args)
		return op, edge, err
	case "where":
		edge.Child = query.OperationID(function.Name)
		op, err := NewWhereOperation(function.Args)
		return op, edge, err
	default:
		return nil, query.Edge{}, fmt.Errorf("Unknown function")
	}
}

func NewSumOperation(args []*FunctionArg) (*query.Operation, error) {
	// TODO: Validate
	if len(args) != 0 {
		return nil, fmt.Errorf("Sum takes no args yo!")
	}
	return &query.Operation{
		ID:   "sum",
		Spec: &query.SumOpSpec{},
	}, nil
}

func NewCountOperation(args []*FunctionArg) (*query.Operation, error) {
	// TODO: Validate
	if len(args) != 0 {
		return nil, fmt.Errorf("Count takes no args yo!")
	}
	return &query.Operation{
		ID:   "count",
		Spec: &query.CountOpSpec{},
	}, nil
}

func NewClearOperation(args []*FunctionArg) (*query.Operation, error) {
	// TODO: Validate
	if len(args) != 0 {
		return nil, fmt.Errorf("Clear takes no args yo!")
	}
	return &query.Operation{
		ID:   "clear",
		Spec: &query.ClearOpSpec{},
	}, nil
}

func NewSelectOperation(args []*FunctionArg) (*query.Operation, error) {
	// TODO: Validate
	if len(args) != 1 {
		return nil, fmt.Errorf("Please specify database")
	}

	arg := args[0]
	if arg.Name != "database" {
		return nil, fmt.Errorf("Argument is not database: %s", arg.Name)
	}

	if arg.Arg.Type() != StringKind {
		return nil, fmt.Errorf("You are not a string!")
	}

	database := arg.Arg.Value().(string)
	return &query.Operation{
		ID: "select", // TODO: Change this to a UUID
		Spec: &query.SelectOpSpec{
			Database: database,
		},
	}, nil
}

func NewWhereOperation(args []*FunctionArg) (*query.Operation, error) {
	if len(args) != 1 || args[0].Name != "exp" {
		return nil, fmt.Errorf("Invalid Where clause... also I should make this error better")
	}
	arg := args[0].Arg
	expr := arg.Value().(*BinaryExpression)

	root, err := NewBinaryNode(expr)
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

func NewBinaryNode(expr *BinaryExpression) (node *storage.Node, err error) {
	switch expr.Operator {
	case "==", "!=", "startswith":
		node, err = NewComparisonNode(expr)
	case "and", "or":
		node, err = NewLogicalNode(expr)
	case "<=", "<", ">=", ">", "in", "not empty", "empty":
		err = fmt.Errorf("Operator %s has not been implemented yet", expr.Operator)
	default:
		err = fmt.Errorf("Unsupported operator %s", expr.Operator)
	}
	return
}

func NewLogicalOperator(op string) (storage.Node_Logical, error) {
	if op == "and" {
		return storage.LogicalAnd, nil
	} else if op == "or" {
		return storage.LogicalOr, nil
	}
	return 0, fmt.Errorf("Unknown logical operator %s", op)
}

func NewLogicalNode(expr *BinaryExpression) (*storage.Node, error) {
	lhs, ok := expr.Left.(*BinaryExpression)
	if !ok {
		return nil, fmt.Errorf("Left-hand side of logical expression not binary")
	}

	left, err := NewBinaryNode(lhs)
	if err != nil {
		return nil, err
	}

	rhs, ok := expr.Right.(*BinaryExpression)
	if !ok {
		return nil, fmt.Errorf("Right-hand side of logical expression not binary")
	}

	right, err := NewBinaryNode(rhs)
	if err != nil {
		return nil, err
	}

	op, err := NewLogicalOperator(expr.Operator)
	if err != nil {
		return nil, err
	}

	return &storage.Node{
		NodeType: storage.NodeTypeGroupExpression,
		Value:    &storage.Node_Logical_{Logical: op},
		Children: []*storage.Node{left, right},
	}, nil
}

func NewComparisonOperator(op string, isRegex bool) (storage.Node_Comparison, error) {
	// "<=" / "<" / ">=" / ">" / "=" / "!=" / "startsWith"i / "in"i / "not empty"i / "empty"i
	switch op {
	case "==":
		if isRegex {
			return storage.ComparisonRegex, nil
		}
		return storage.ComparisonEqual, nil
	case "!=":
		if isRegex {
			return storage.ComparisonNotRegex, nil
		}
		return storage.ComparisonNotEqual, nil
	case "startswith":
		return storage.ComparisonStartsWith, nil
	case "<=", "<", ">=", ">", "in", "not empty", "empty":
		return 0, fmt.Errorf("Unimplemented comparison operator %s", op)
	default:
		return 0, fmt.Errorf("Unknown comparison operator %s", op)
	}
}

func NewComparisonNode(expr *BinaryExpression) (*storage.Node, error) {
	if _, ok := expr.Left.(*BinaryExpression); ok {
		return nil, fmt.Errorf("Left-hand side of comparison expression cannot be a binary expression")
	}
	if _, ok := expr.Right.(*BinaryExpression); ok {
		return nil, fmt.Errorf("Right-hand side of comparison expression cannot be a binary expression")
	}

	var isRegex bool
	var left *storage.Node
	switch lhs := expr.Left.(type) {
	case *StringLiteral, *Field:
		// If the string literal or field is on the left side we assume this is a reference
		left = NewNodeRef(lhs)
	case *Regex:
		isRegex = true
		left = NewNodeLiteral(lhs)
	case *Number:
		left = NewNodeLiteral(lhs)
	case *Duration, *DateTime:
		return nil, fmt.Errorf("We don't support left-hand side durations or date times yet. sorry")
	default:
		return nil, fmt.Errorf("Unknown left-hand side expression")
	}

	var right *storage.Node
	switch rhs := expr.Right.(type) {
	case *StringLiteral, *Field, *Number:
		// If the string literal or field is on the left side we assume this is a reference
		right = NewNodeLiteral(rhs)
	case *Regex:
		isRegex = true
		right = NewNodeLiteral(rhs)
	case *Duration, *DateTime:
		return nil, fmt.Errorf("We don't support right-hand side durations or date times yet. sorry")
	default:
		return nil, fmt.Errorf("Unknown righthand side expression")
	}

	op, err := NewComparisonOperator(expr.Operator, isRegex)
	if err != nil {
		return nil, err
	}

	return &storage.Node{
		NodeType: storage.NodeTypeBooleanExpression,
		Value:    &storage.Node_Comparison_{Comparison: op},
		Children: []*storage.Node{left, right},
	}, nil

}

func NewRangeOperation(args []*FunctionArg) (*query.Operation, error) {
	// TODO: Validate
	if len(args) == 1 && args[0].Name == "start" {
		start, err := NewTime(args[0].Arg)
		if err != nil {
			return nil, err
		}
		return &query.Operation{
			ID: "range", // TODO: Change this to a UUID
			Spec: &query.RangeOpSpec{
				Start: start,
				Stop: query.Time{
					Absolute: time.Now(),
				},
			},
		}, nil
	} else if len(args) == 1 && args[0].Name == "stop" {
		return nil, fmt.Errorf("Must specify a start time")
	} else if len(args) == 2 {
		// TODO: fix this logic to prevent 2 start times
		var start query.Time
		var stop query.Time
		var err error
		// TODO: oh boy this is getting bad...
		switch args[0].Name {
		case "start":
			start, err = NewTime(args[0].Arg)
			if err != nil {
				return nil, err
			}
		case "stop":
			stop, err = NewTime(args[0].Arg)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknown range argument: %s", args[0].Name)
		}

		switch args[1].Name {
		case "start":
			start, err = NewTime(args[1].Arg)
			if err != nil {
				return nil, err
			}
		case "stop":
			stop, err = NewTime(args[1].Arg)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknown range argument: %s", args[1].Name)
		}
		return &query.Operation{
			ID: "range", // TODO: Change this to a UUID
			Spec: &query.RangeOpSpec{
				Start: start,
				Stop:  stop,
			},
		}, nil

	}
	return nil, fmt.Errorf("no idea what this error is... chris, what is going on?")
}

func NewTime(arg Arg) (query.Time, error) {
	kind := arg.Type()
	switch kind {
	case DateTimeKind:
		return query.Time{
			Absolute: arg.Value().(time.Time),
		}, nil
	case DurationKind:
		return query.Time{
			Relative: arg.Value().(time.Duration),
		}, nil
		// TODO: convert from float64 to time.time
	case NumberKind:
		return query.Time{
			Absolute: time.Now(),
		}, nil
	default:
		return query.Time{}, fmt.Errorf("Unknown time type")
	}
}

func NewNodeRef(val interface{}) *storage.Node {
	switch v := val.(type) {
	case *StringLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeRef,
			Value: &storage.Node_RefValue{
				RefValue: v.Value().(string),
			},
		}
	case *Field: // TODO: Change this to a field node when available (remove from this area)
		return &storage.Node{
			NodeType: storage.NodeTypeRef,
			Value: &storage.Node_RefValue{
				RefValue: v.Value().(string),
			},
		}
	}
	return nil
}

func NewNodeLiteral(val interface{}) *storage.Node {
	switch v := val.(type) {
	case *StringLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: v.Value().(string),
			},
		}
	case *Number:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: v.Value().(float64),
			},
		}
	case *Regex:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: v.Value().(*regexp.Regexp).String(),
			},
		}
	}
	return nil
}
