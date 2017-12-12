package execute

import (
	"fmt"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/pkg/errors"
)

func ToStoragePredicate(f *ast.ArrowFunctionExpression) (*storage.Predicate, error) {
	if len(f.Params) != 1 {
		return nil, errors.New("storage predicate functions must have exactly one parameter")
	}

	root, err := toStoragePredicate(f.Body.(ast.Expression), f.Params[0].Name)
	if err != nil {
		return nil, err
	}

	return &storage.Predicate{
		Root: root,
	}, nil
}

func toStoragePredicate(n ast.Expression, objectName string) (*storage.Node, error) {
	switch n := n.(type) {
	case *ast.LogicalExpression:
		left, err := toStoragePredicate(n.Left, objectName)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := toStoragePredicate(n.Right, objectName)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		children := []*storage.Node{left, right}
		switch n.Operator {
		case ast.AndOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
				Children: children,
			}, nil
		case ast.OrOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalOr},
				Children: children,
			}, nil
		default:
			return nil, fmt.Errorf("unknown logical operator %v", n.Operator)
		}
	case *ast.BinaryExpression:
		left, err := toStoragePredicate(n.Left, objectName)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := toStoragePredicate(n.Right, objectName)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		children := []*storage.Node{left, right}
		op, err := toComparisonOperator(n.Operator)
		if err != nil {
			return nil, err
		}
		return &storage.Node{
			NodeType: storage.NodeTypeComparisonExpression,
			Value:    &storage.Node_Comparison_{Comparison: op},
			Children: children,
		}, nil
	case *ast.StringLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: n.Value,
			},
		}, nil
	case *ast.IntegerLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_IntegerValue{
				IntegerValue: n.Value,
			},
		}, nil
	case *ast.BooleanLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_BooleanValue{
				BooleanValue: n.Value,
			},
		}, nil
	case *ast.NumberLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: n.Value,
			},
		}, nil
	case *ast.RegexpLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: n.Value.String(),
			},
		}, nil
	case *ast.MemberExpression:
		// Sanity check that the object is the objectName identifier
		if ident, ok := n.Object.(*ast.Identifier); !ok || ident.Name != objectName {
			return nil, fmt.Errorf("unknown object %q", n.Object)
		}
		property, err := propertyName(n)
		if err != nil {
			return nil, err
		}
		if property == "_value" {
			return &storage.Node{
				NodeType: storage.NodeTypeFieldRef,
				Value: &storage.Node_FieldRefValue{
					FieldRefValue: "_value",
				},
			}, nil
		}
		return &storage.Node{
			NodeType: storage.NodeTypeTagRef,
			Value: &storage.Node_TagRefValue{
				TagRefValue: property,
			},
		}, nil
	case *ast.DurationLiteral:
		return nil, errors.New("duration literals not supported in storage predicates")
	case *ast.DateTimeLiteral:
		return nil, errors.New("time literals not supported in storage predicates")
	default:
		return nil, fmt.Errorf("unsupported ast expression type %T", n)
	}
}

func toComparisonOperator(o ast.OperatorKind) (storage.Node_Comparison, error) {
	switch o {
	case ast.EqualOperator:
		return storage.ComparisonEqual, nil
	case ast.NotOperator:
		return storage.ComparisonNotEqual, nil
	case ast.StartsWithOperator:
		return storage.ComparisonStartsWith, nil
	case ast.LessThanOperator:
		return storage.ComparisonLess, nil
	case ast.LessThanEqualOperator:
		return storage.ComparisonLessEqual, nil
	case ast.GreaterThanOperator:
		return storage.ComparisonGreater, nil
	case ast.GreaterThanEqualOperator:
		return storage.ComparisonGreaterEqual, nil
	default:
		return 0, fmt.Errorf("unknown operator %v", o)
	}
}
