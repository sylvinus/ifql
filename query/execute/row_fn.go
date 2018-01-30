package execute

import (
	"fmt"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/compiler"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/ifql/semantic"
	"github.com/pkg/errors"
)

type rowFn struct {
	fn               *semantic.FunctionExpression
	compilationCache *compiler.CompilationCache
	scope            compiler.Scope

	preparedFn compiler.Func

	recordName string
	record     *compiler.Object

	recordCols map[string]int
	references []string
}

func newRowFn(fn *semantic.FunctionExpression) (rowFn, error) {
	if len(fn.Params) != 1 {
		return rowFn{}, fmt.Errorf("function should only have a single parameter, got %d", len(fn.Params))
	}
	return rowFn{
		compilationCache: compiler.NewCompilationCache(fn),
		scope:            make(compiler.Scope, 1),
		recordName:       fn.Params[0].Key.Name,
		references:       findColReferences(fn),
		recordCols:       make(map[string]int),
		record:           compiler.NewObject(),
	}, nil
}

func (f *rowFn) prepare(cols []ColMeta) error {
	// Prepare types and recordCols
	propertyTypes := make(map[string]semantic.Type, len(f.references))
	for _, r := range f.references {
		found := false
		for j, c := range cols {
			if r == c.Label {
				f.recordCols[r] = j
				found = true
				propertyTypes[r] = ConvertToKind(c.Type)
				break
			}
		}
		if !found {
			return fmt.Errorf("function references unknown column %q", r)
		}
	}
	// Compile fn for given types
	fn, err := f.compilationCache.Compile(map[string]semantic.Type{
		f.recordName: semantic.NewObjectType(propertyTypes),
	})
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

func ConvertToKind(t DataType) semantic.Kind {
	// TODO make this an array lookup.
	switch t {
	case TInvalid:
		return semantic.Invalid
	case TBool:
		return semantic.Bool
	case TInt:
		return semantic.Int
	case TUInt:
		return semantic.UInt
	case TFloat:
		return semantic.Float
	case TString:
		return semantic.String
	case TTime:
		return semantic.Time
	default:
		return semantic.Invalid
	}
}

func ConvertFromKind(k semantic.Kind) DataType {
	// TODO make this an array lookup.
	switch k {
	case semantic.Invalid:
		return TInvalid
	case semantic.Bool:
		return TBool
	case semantic.Int:
		return TInt
	case semantic.UInt:
		return TUInt
	case semantic.Float:
		return TFloat
	case semantic.String:
		return TString
	case semantic.Time:
		return TTime
	default:
		return TInvalid
	}
}

func (f *rowFn) eval(row int, rr RowReader) (compiler.Value, error) {
	for _, r := range f.references {
		f.record.Set(r, ValueForRow(row, f.recordCols[r], rr))
	}
	f.scope[f.recordName] = f.record
	return f.preparedFn.Eval(f.scope)
}

type RowPredicateFn struct {
	rowFn
}

func NewRowPredicateFn(fn *semantic.FunctionExpression) (*RowPredicateFn, error) {
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
	if f.preparedFn.Type() != semantic.Bool {
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
	wrapObj *compiler.Object
}

func NewRowMapFn(fn *semantic.FunctionExpression) (*RowMapFn, error) {
	r, err := newRowFn(fn)
	if err != nil {
		return nil, err
	}
	return &RowMapFn{
		rowFn:   r,
		wrapObj: compiler.NewObject(),
	}, nil
}

func (f *RowMapFn) Prepare(cols []ColMeta) error {
	err := f.rowFn.prepare(cols)
	if err != nil {
		return err
	}
	k := f.preparedFn.Type().Kind()
	f.isWrap = k != semantic.Object
	if f.isWrap {
		f.wrapObj.SetPropertyType(DefaultValueColLabel, f.preparedFn.Type())
	}
	return nil
}

func (f *RowMapFn) Type() semantic.Type {
	if f.isWrap {
		return f.wrapObj.Type()
	}
	return f.preparedFn.Type()
}

func (f *RowMapFn) Eval(row int, rr RowReader) (*compiler.Object, error) {
	v, err := f.rowFn.eval(row, rr)
	if err != nil {
		return nil, err
	}
	if f.isWrap {
		f.wrapObj.Set(DefaultValueColLabel, v)
		return f.wrapObj, nil
	}
	return v.Object(), nil
}

func ValueForRow(i, j int, rr RowReader) compiler.Value {
	t := rr.Cols()[j].Type
	switch t {
	case TBool:
		return compiler.NewBool(rr.AtBool(i, j))
	case TInt:
		return compiler.NewInt(rr.AtInt(i, j))
	case TUInt:
		return compiler.NewUInt(rr.AtUInt(i, j))
	case TFloat:
		return compiler.NewFloat(rr.AtFloat(i, j))
	case TString:
		return compiler.NewString(rr.AtString(i, j))
	case TTime:
		return compiler.NewTime(compiler.Time(rr.AtTime(i, j)))
	default:
		PanicUnknownType(t)
		return nil
	}
}

func AppendValue(builder BlockBuilder, j int, v compiler.Value) {
	switch k := v.Type().Kind(); k {
	case semantic.Bool:
		builder.AppendBool(j, v.Bool())
	case semantic.Int:
		builder.AppendInt(j, v.Int())
	case semantic.UInt:
		builder.AppendUInt(j, v.UInt())
	case semantic.Float:
		builder.AppendFloat(j, v.Float())
	case semantic.String:
		builder.AppendString(j, v.Str())
	case semantic.Time:
		builder.AppendTime(j, Time(v.Time()))
	default:
		PanicUnknownType(ConvertFromKind(k))
	}
}

func ToStoragePredicate(f *semantic.FunctionExpression) (*storage.Predicate, error) {
	if len(f.Params) != 1 {
		return nil, errors.New("storage predicate functions must have exactly one parameter")
	}

	root, err := toStoragePredicate(f.Body.(semantic.Expression), f.Params[0].Key.Name)
	if err != nil {
		return nil, err
	}

	return &storage.Predicate{
		Root: root,
	}, nil
}

func toStoragePredicate(n semantic.Expression, objectName string) (*storage.Node, error) {
	switch n := n.(type) {
	case *semantic.LogicalExpression:
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
	case *semantic.BinaryExpression:
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
	case *semantic.StringLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: n.Value,
			},
		}, nil
	case *semantic.IntegerLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_IntegerValue{
				IntegerValue: n.Value,
			},
		}, nil
	case *semantic.BooleanLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_BooleanValue{
				BooleanValue: n.Value,
			},
		}, nil
	case *semantic.FloatLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: n.Value,
			},
		}, nil
	case *semantic.RegexpLiteral:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: n.Value.String(),
			},
		}, nil
	case *semantic.MemberExpression:
		// Sanity check that the object is the objectName identifier
		if ident, ok := n.Object.(*semantic.IdentifierExpression); !ok || ident.Name != objectName {
			return nil, fmt.Errorf("unknown object %q", n.Object)
		}
		if n.Property == "_value" {
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
				TagRefValue: n.Property,
			},
		}, nil
	case *semantic.DurationLiteral:
		return nil, errors.New("duration literals not supported in storage predicates")
	case *semantic.DateTimeLiteral:
		return nil, errors.New("time literals not supported in storage predicates")
	default:
		return nil, fmt.Errorf("unsupported semantic expression type %T", n)
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

func findColReferences(fn *semantic.FunctionExpression) []string {
	v := &colReferenceVisitor{
		recordName: fn.Params[0].Key.Name,
	}
	semantic.Walk(v, fn)
	return v.refs
}

type colReferenceVisitor struct {
	recordName string
	refs       []string
}

func (c *colReferenceVisitor) Visit(node semantic.Node) semantic.Visitor {
	if me, ok := node.(*semantic.MemberExpression); ok {
		if obj, ok := me.Object.(*semantic.IdentifierExpression); ok && obj.Name == c.recordName {
			c.refs = append(c.refs, me.Property)
		}
	}
	return c
}

func (c *colReferenceVisitor) Done() {}
