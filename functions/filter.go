package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

const FilterKind = "filter"

type FilterOpSpec struct {
	Expression expression.Expression `json:"expression"`
}

func init() {
	ifql.RegisterMethod(FilterKind, createFilterOpSpec)
	query.RegisterOpSpec(FilterKind, newFilterOp)
	plan.RegisterProcedureSpec(FilterKind, newFilterProcedure, FilterKind)
	execute.RegisterTransformation(FilterKind, createFilterTransformation)
}

func createFilterOpSpec(args ifql.Arguments, ctx ifql.Context) (query.OperationSpec, error) {
	expr, err := args.GetRequiredExpression("f")
	if err != nil {
		return nil, err
	}

	return &FilterOpSpec{
		Expression: expr,
	}, nil
}
func newFilterOp() query.OperationSpec {
	return new(FilterOpSpec)
}

func (s *FilterOpSpec) Kind() query.OperationKind {
	return FilterKind
}

type FilterProcedureSpec struct {
	Expression expression.Expression
}

func newFilterProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FilterOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FilterProcedureSpec{
		Expression: spec.Expression,
	}, nil
}

func (s *FilterProcedureSpec) Kind() plan.ProcedureKind {
	return FilterKind
}
func (s *FilterProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FilterProcedureSpec)
	//TODO copy expression
	ns.Expression = s.Expression
	return ns
}

func (s *FilterProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    FromKind,
		Through: []plan.ProcedureKind{GroupKind, LimitKind, RangeKind},
	}
}
func (s *FilterProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*FromProcedureSpec)
	if selectSpec.FilterSet {
		root = dup()
		selectSpec = root.Spec.(*FromProcedureSpec)
		selectSpec.FilterSet = false
		selectSpec.Filter = expression.Expression{}
		return
	}
	selectSpec.FilterSet = true
	selectSpec.Filter = s.Expression
}

func createFilterTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FilterProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache(ctx.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewFilterTransformation(d, cache, s)
	return t, d, nil
}

type filterTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	names     []string
	scope     execute.Scope
	scopeCols map[string]int
	ces       map[execute.DataType]expressionOrError

	colMap []int
}

type expressionOrError struct {
	Err  error
	Expr execute.CompiledExpression
}

func NewFilterTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *FilterProcedureSpec) *filterTransformation {
	names := execute.ExpressionNames(spec.Expression.Root)
	types := make(map[string]execute.DataType, len(names))
	ces := make(map[execute.DataType]expressionOrError, len(execute.ValueDataTypes))
	for _, n := range names {
		if n != "$" {
			types[n] = execute.TString
		}
	}
	for _, typ := range execute.ValueDataTypes {
		types["$"] = typ
		ce, err := execute.CompileExpression(spec.Expression, types)
		ces[typ] = expressionOrError{
			Err:  err,
			Expr: ce,
		}
		if err == nil && ce.Type() != execute.TBool {
			ces[typ] = expressionOrError{
				Err:  errors.New("expression does not evaluate to boolean"),
				Expr: nil,
			}
		}
	}

	return &filterTransformation{
		d:         d,
		cache:     cache,
		names:     names,
		scope:     make(execute.Scope),
		scopeCols: make(map[string]int),
		ces:       ces,
	}
}

func (t *filterTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *filterTransformation) Process(id execute.DatasetID, b execute.Block) error {
	builder, new := t.cache.BlockBuilder(b)
	if new {
		execute.AddBlockCols(b, builder)
	}

	ncols := builder.NCols()
	if cap(t.colMap) < ncols {
		t.colMap = make([]int, ncols)
		for j := range t.colMap {
			t.colMap[j] = j
		}
	} else {
		t.colMap = t.colMap[:ncols]
	}

	// Prepare scope
	cols := b.Cols()
	valueIdx := execute.ValueIdx(cols)
	for j, c := range cols {
		if c.Label == execute.ValueColLabel {
			t.scopeCols["$"] = valueIdx
		} else {
			for _, k := range t.names {
				if k == c.Label {
					t.scopeCols[c.Label] = j
					break
				}
			}
		}
	}

	valueCol := cols[valueIdx]
	exprErr := t.ces[valueCol.Type]
	if exprErr.Err != nil {
		return errors.Wrapf(exprErr.Err, "expression does not support type %v", valueCol.Type)
	}
	ce := exprErr.Expr

	// Append only matching rows to block
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			for _, k := range t.names {
				t.scope[k] = execute.ValueForRow(i, t.scopeCols[k], rr)
			}
			if pass, err := ce.EvalBool(t.scope); !pass {
				if err != nil {
					log.Printf("failed to evaluate expression: %v", err)
				}
				continue
			}
			for j, c := range cols {
				if c.IsCommon {
					continue
				}
				switch c.Type {
				case execute.TBool:
					builder.AppendBool(j, rr.AtBool(i, t.colMap[j]))
				case execute.TInt:
					builder.AppendInt(j, rr.AtInt(i, t.colMap[j]))
				case execute.TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, t.colMap[j]))
				case execute.TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, t.colMap[j]))
				case execute.TString:
					builder.AppendString(j, rr.AtString(i, t.colMap[j]))
				case execute.TTime:
					builder.AppendTime(j, rr.AtTime(i, t.colMap[j]))
				default:
					execute.PanicUnknownType(c.Type)
				}
			}
		}
	})
	return nil
}

func (t *filterTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *filterTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *filterTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
func (t *filterTransformation) SetParents(ids []execute.DatasetID) {
}
