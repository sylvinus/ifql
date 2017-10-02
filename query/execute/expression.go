package execute

import (
	"fmt"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/pkg/errors"
)

func ExpressionToStoragePredicate(e expression.Node) (*storage.Predicate, error) {
	root, err := doExpression(e)
	if err != nil {
		return nil, err
	}
	return &storage.Predicate{
		Root: root,
	}, nil
}

func doExpression(e expression.Node) (*storage.Node, error) {
	switch expr := e.(type) {
	case *expression.BinaryNode:
		left, err := doExpression(expr.Left)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := doExpression(expr.Right)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		children := []*storage.Node{left, right}
		switch expr.Operator {
		case expression.AndOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
				Children: children,
			}, nil
		case expression.OrOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalOr},
				Children: children,
			}, nil
		}
		op, err := toComparisonOperator(expr.Operator)
		if err != nil {
			return nil, err
		}
		return &storage.Node{
			NodeType: storage.NodeTypeComparisonExpression,
			Value:    &storage.Node_Comparison_{Comparison: op},
			Children: children,
		}, nil
	case *expression.StringLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: expr.Value,
			},
		}, nil
	case *expression.IntegerLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_IntegerValue{
				IntegerValue: expr.Value,
			},
		}, nil
	case *expression.BooleanLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_BooleanValue{
				BooleanValue: expr.Value,
			},
		}, nil
	case *expression.FloatLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: expr.Value,
			},
		}, nil
	case *expression.RegexpLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: expr.Value,
			},
		}, nil
	case *expression.ReferenceNode:
		switch expr.Kind {
		case "tag":
			return &storage.Node{
				NodeType: storage.NodeTypeTagRef,
				Value: &storage.Node_TagRefValue{
					TagRefValue: expr.Name,
				},
			}, nil
		case "field":
			return &storage.Node{
				NodeType: storage.NodeTypeFieldRef,
				Value: &storage.Node_FieldRefValue{
					FieldRefValue: "_field",
				},
			}, nil
		default:
			return nil, fmt.Errorf("unsupported reference kind %q", expr.Kind)
		}
	case *expression.DurationLiteralNode:
		return nil, errors.New("duration literals not supported in storage predicates")
	case *expression.TimeLiteralNode:
		return nil, errors.New("time literals not supported in storage predicates")
	default:
		return nil, fmt.Errorf("unsupported expression type %v", e.Type())
	}
}

func toComparisonOperator(o expression.Operator) (storage.Node_Comparison, error) {
	switch o {
	case expression.EqualOperator:
		return storage.ComparisonEqual, nil
	case expression.NotOperator:
		return storage.ComparisonNotEqual, nil
	case expression.StartsWithOperator:
		return storage.ComparisonStartsWith, nil
	case expression.RegexpMatchOperator:
		return storage.ComparisonRegex, nil
	case expression.RegexpNotMatchOperator:
		return storage.ComparisonNotRegex, nil
	case expression.LessThanOperator:
		return storage.ComparisonLess, nil
	case expression.LessThanEqualOperator:
		return storage.ComparisonLessEqual, nil
	case expression.GreaterThanOperator:
		return storage.ComparisonGreater, nil
	case expression.GreaterThanEqualOperator:
		return storage.ComparisonGreaterEqual, nil
	default:
		return 0, fmt.Errorf("unknown expression operator %v", o)
	}
}

func ExpressionNames(n expression.Node) []string {
	var names []string
	expression.Walk(n, func(n expression.Node) error {
		if rn, ok := n.(*expression.ReferenceNode); ok {
			found := false
			for _, n := range names {
				if n == rn.Name {
					found = true
					break
				}
			}
			if !found {
				names = append(names, rn.Name)
			}
		}
		return nil
	})
	return names
}

type Scope map[string]float64

func (s Scope) Names() []string {
	names := make([]string, 0, len(s))
	for k := range s {
		names = append(names, k)
	}
	return names
}

type floatStack struct {
	data []float64
}

func (s *floatStack) push(v float64) {
	s.data = append(s.data, v)
}
func (s *floatStack) pop() float64 {
	l := len(s.data)
	v := s.data[l-1]
	s.data = s.data[:l-1]
	return v
}
func (s *floatStack) len() int {
	return len(s.data)
}

func EvalExpression(n expression.Node, s Scope) (float64, error) {
	stack := &floatStack{data: make([]float64, 0, 64)}
	err := eval(n, s, stack)
	if err != nil {
		return 0, err
	}
	if stack.len() != 1 {
		return 0, errors.New("expression didn't return a value")
	}
	return stack.pop(), nil
}

func eval(n expression.Node, s Scope, stack *floatStack) error {
	switch node := n.(type) {
	case *expression.IntegerLiteralNode:
		stack.push(float64(node.Value))
	case *expression.FloatLiteralNode:
		stack.push(node.Value)
	case *expression.ReferenceNode:
		v, ok := s[node.Name]
		if !ok {
			return fmt.Errorf("name %q not in scope: %v", node.Name, s.Names())
		}
		stack.push(v)
	case *expression.BinaryNode:
		if err := eval(node.Left, s, stack); err != nil {
			return err
		}
		if err := eval(node.Right, s, stack); err != nil {
			return err
		}
		r := stack.pop()
		l := stack.pop()
		switch node.Operator {
		case expression.AdditionOperator:
			stack.push(l + r)
		case expression.SubtractionOperator:
			stack.push(l - r)
		case expression.MultiplicationOperator:
			stack.push(l * r)
		case expression.DivisionOperator:
			stack.push(l / r)
		default:
			return fmt.Errorf("unsupported binary operator %v", node.Operator)
		}
	case *expression.UnaryNode:
		err := eval(node.Node, s, stack)
		if err != nil {
			return err
		}
		v := stack.pop()
		switch node.Operator {
		case expression.SubtractionOperator:
			stack.push(0 - v)
		default:
			return fmt.Errorf("unsupported unary operator %v", node.Operator)
		}
	default:
		return fmt.Errorf("unsupported expression %T", n)
	}
	return nil
}
