package ifql

import (
	"encoding/json"
	"log"

	"github.com/influxdata/ifql/query"
)

func NewQuery(ifql string, opts ...Option) (*query.QuerySpec, error) {
	function, err := NewAST(ifql, opts...)
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(function, "", "    ")
	log.Printf("%s", string(b))
	return nil, err
	/*
		ops, edges, err := NewOperations(function, "")
		if err != nil {
			return nil, err
		}
		return &query.QuerySpec{
			Operations: ops,
			Edges:      edges,
		}, nil
	*/
}

/*
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
*/

/*
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
	case "window":
		edge.Child = query.OperationID(function.Name)
		op, err := NewWindowOperation(function.Args)
		return op, edge, err
	default:
		return nil, query.Edge{}, fmt.Errorf("Unknown function")
	}
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

/*
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
*/
