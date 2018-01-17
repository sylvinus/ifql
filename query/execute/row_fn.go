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
	references       []compiler.Reference
	referencePaths   []compiler.ReferencePath
	compilationCache *compiler.CompilationCache
	scope            compiler.Scope

	scopeCols map[compiler.ReferencePath]int
	types     map[compiler.ReferencePath]compiler.Type

	preparedFn compiler.Func
}

func newRowFn(fn *semantic.ArrowFunctionExpression) (rowFn, error) {
	if len(fn.Params) != 1 {
		return rowFn{}, fmt.Errorf("function should only have a single parameter, got %v", fn.Params)
	}
	references, err := compiler.FindReferences(fn)
	if err != nil {
		return rowFn{}, err
	}

	referencePaths := make([]compiler.ReferencePath, len(references))
	for i, r := range references {
		referencePaths[i] = r.Path()
	}

	return rowFn{
		references:       references,
		referencePaths:   referencePaths,
		compilationCache: compiler.NewCompilationCache(fn, referencePaths),
		scope:            make(compiler.Scope, len(references)),
		scopeCols:        make(map[compiler.ReferencePath]int, len(references)),
		types:            make(map[compiler.ReferencePath]compiler.Type, len(references)),
	}, nil
}

func (f *rowFn) prepare(cols []ColMeta) error {
	// Prepare types and scopeCols
	for i, r := range f.references {
		rp := f.referencePaths[i]
		found := false
		for j, c := range cols {
			if r[1] == c.Label {
				f.scopeCols[rp] = j
				f.types[rp] = ConvertToCompilerType(c.Type)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("function references unknown value %q", rp)
		}
	}
	// Compile fn for given types
	fn, err := f.compilationCache.Compile(f.types)
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

func ConvertToCompilerType(t DataType) compiler.Type {
	// TODO make this an array lookup.
	switch t {
	case TInvalid:
		return compiler.TInvalid
	case TBool:
		return compiler.TBool
	case TInt:
		return compiler.TInt
	case TUInt:
		return compiler.TUInt
	case TFloat:
		return compiler.TFloat
	case TString:
		return compiler.TString
	case TTime:
		return compiler.TTime
	default:
		return compiler.TInvalid
	}
}
func ConvertFromCompilerType(t compiler.Type) DataType {
	// TODO make this an array lookup.
	switch t {
	case compiler.TInvalid:
		return TInvalid
	case compiler.TBool:
		return TBool
	case compiler.TInt:
		return TInt
	case compiler.TUInt:
		return TUInt
	case compiler.TFloat:
		return TFloat
	case compiler.TString:
		return TString
	case compiler.TTime:
		return TTime
	default:
		return TInvalid
	}
}

func (f *rowFn) eval(row int, rr RowReader) (compiler.Value, error) {
	for _, rp := range f.referencePaths {
		f.scope[rp] = ValueForRow(row, f.scopeCols[rp], rr)
	}
	return f.preparedFn.Eval(f.scope)
}

type RowPredicateFn struct {
	rowFn
}

func NewRowPredicateFn(fn *semantic.ArrowFunctionExpression) (*RowPredicateFn, error) {
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
	if f.preparedFn.Type() != compiler.TBool {
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
	wrapMap compiler.Map
}

func NewRowMapFn(fn *semantic.ArrowFunctionExpression) (*RowMapFn, error) {
	r, err := newRowFn(fn)
	if err != nil {
		return nil, err
	}
	return &RowMapFn{
		rowFn: r,
		wrapMap: compiler.Map{
			Meta: compiler.MapMeta{
				Properties: []compiler.MapPropertyMeta{{Key: DefaultValueColLabel}},
			},
			Values: make(map[string]compiler.Value, 1),
		},
	}, nil
}

func (f *RowMapFn) Prepare(cols []ColMeta) error {
	err := f.rowFn.prepare(cols)
	if err != nil {
		return err
	}
	t := f.preparedFn.Type()
	f.isWrap = t != compiler.TMap
	if f.isWrap {
		f.wrapMap.Meta.Properties[0].Type = t
	}
	return nil
}

func (f *RowMapFn) MapMeta() compiler.MapMeta {
	if f.isWrap {
		return f.wrapMap.Meta
	}
	return f.preparedFn.MapMeta()
}

func (f *RowMapFn) Eval(row int, rr RowReader) (compiler.Map, error) {
	v, err := f.rowFn.eval(row, rr)
	if err != nil {
		return compiler.Map{}, err
	}
	if f.isWrap {
		f.wrapMap.Values[DefaultValueColLabel] = v
		return f.wrapMap, nil
	}
	return v.Map(), nil
}

func ValueForRow(i, j int, rr RowReader) (v compiler.Value) {
	t := rr.Cols()[j].Type
	v.Type = ConvertToCompilerType(t)
	switch t {
	case TBool:
		v.Value = rr.AtBool(i, j)
	case TInt:
		v.Value = rr.AtInt(i, j)
	case TUInt:
		v.Value = rr.AtUInt(i, j)
	case TFloat:
		v.Value = rr.AtFloat(i, j)
	case TString:
		v.Value = rr.AtString(i, j)
	case TTime:
		v.Value = rr.AtTime(i, j)
	default:
		PanicUnknownType(t)
	}
	return
}

func ToStoragePredicate(f *semantic.ArrowFunctionExpression) (*storage.Predicate, error) {
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
		if ident, ok := n.Object.(*semantic.Identifier); !ok || ident.Name != objectName {
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
