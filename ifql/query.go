package ifql

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/query"
)

func NewQuery(ifql string) (*query.QuerySpec, error) {
	function, err := NewAST(ifql)
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
