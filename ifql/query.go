package ifql

import (
	"fmt"
	"strings"
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
	node := arg.Value().(*storage.Node)
	return &query.Operation{
		ID: "where", // TODO: Change this to a UUID
		Spec: &query.WhereOpSpec{
			Exp: &query.WhereExpressionSpec{
				Predicate: &storage.Predicate{
					Root: node,
				},
			},
		},
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
	}
	return nil
}

func NewComparisonOperator(text []byte) (storage.Node_Comparison, error) {
	op := strings.ToLower(string(text))
	// "<=" / "<" / ">=" / ">" / "=" / "!=" / "startsWith"i / "in"i / "not empty"i / "empty"i
	switch op {
	case "=":
		return storage.ComparisonEqual, nil
	case "!=":
		return storage.ComparisonNotEqual, nil
	case "startswith":
		return storage.ComparisonStartsWith, nil
	case "<=", "<", ">=", ">", "in", "not empty", "empty":
		return 0, fmt.Errorf("Unimplemented comparison operator %s", op)
	default:
		return 0, fmt.Errorf("Unknown comparison operator %s", op)
	}
}

func NewLogicalOperator(text []byte) (storage.Node_Logical, error) {
	op := strings.ToLower(string(text))
	if op == "and" {
		return storage.LogicalAnd, nil
	} else if op == "or" {
		return storage.LogicalOr, nil
	}
	return 0, fmt.Errorf("Unknown logical operator %s", op)
}

func NewRHS(op, rhs interface{}) (*storage.Node, error) {
	r, ok := rhs.(*storage.Node)
	if !ok {
		if r = NewNodeLiteral(rhs); r == nil {
			return nil, fmt.Errorf("Invalid RHS: not a storage.Node")
		}
	}
	switch o := op.(type) {
	case storage.Node_Comparison:
		return &storage.Node{
			NodeType: storage.NodeTypeBooleanExpression,
			Value:    &storage.Node_Comparison_{Comparison: o},
			Children: []*storage.Node{r},
		}, nil
	case storage.Node_Logical:
		return &storage.Node{
			NodeType: storage.NodeTypeGroupExpression,
			Value:    &storage.Node_Logical_{Logical: o},
			Children: []*storage.Node{r},
		}, nil
	default:
		return nil, fmt.Errorf("Unknown operator type %t %#+v", op, op)
	}
}

func NewExpr(lhs, rhs interface{}) (*storage.Node, error) {
	right := toIfaceSlice(rhs)
	top := right[0].(*storage.Node)
	switch l := lhs.(type) {
	case *storage.Node:
		top.Children = append([]*storage.Node{l}, top.Children...)
	case *StringLiteral, *Field:
		// TODO: we are assuming LHS is always a tag or field
		left := NewNodeRef(l)
		top.Children = append([]*storage.Node{left}, top.Children...)
	case *Duration, *DateTime, *Number:
		return nil, fmt.Errorf("We don't support LHS durations, date times or numbers yet. sorry")
	default:
		return nil, fmt.Errorf("Unknown LHS %t", lhs)
	}

	for _, next := range right[1:] {
		// if op is the same as the top, then this is a child.  Otherwise,
		// we create a new top node where the current top will be left and
		// right will be the top.
		r := next.(*storage.Node)
		if top.NodeType != r.NodeType {
			left := *top
			top = r
			top.Children = append([]*storage.Node{&left}, top.Children...)
			continue
		}
		top.Children = append(top.Children, r)
	}
	return top, nil
}
