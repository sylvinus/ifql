package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

const FilterKind = "filter"

type FilterOpSpec struct {
	Fn *ast.ArrowFunctionExpression `json:"fn"`
}

func init() {
	query.RegisterMethod(FilterKind, createFilterOpSpec)
	query.RegisterOpSpec(FilterKind, newFilterOp)
	plan.RegisterProcedureSpec(FilterKind, newFilterProcedure, FilterKind)
	execute.RegisterTransformation(FilterKind, createFilterTransformation)
}

func createFilterOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	resolved, err := f.Resolve()
	if err != nil {
		return nil, err
	}

	return &FilterOpSpec{
		Fn: resolved,
	}, nil
}
func newFilterOp() query.OperationSpec {
	return new(FilterOpSpec)
}

func (s *FilterOpSpec) Kind() query.OperationKind {
	return FilterKind
}

type FilterProcedureSpec struct {
	Fn *ast.ArrowFunctionExpression
}

func newFilterProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FilterOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FilterProcedureSpec{
		Fn: spec.Fn,
	}, nil
}

func (s *FilterProcedureSpec) Kind() plan.ProcedureKind {
	return FilterKind
}
func (s *FilterProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FilterProcedureSpec)
	ns.Fn = s.Fn.Copy().(*ast.ArrowFunctionExpression)
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
		selectSpec.Filter = nil
		return
	}
	selectSpec.FilterSet = true
	selectSpec.Filter = s.Fn
}

func createFilterTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FilterProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewFilterTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type filterTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	references []execute.Reference
	scope      execute.Scope
	scopeCols  map[string]int
	ces        map[execute.DataType]expressionOrError
}

type expressionOrError struct {
	Err  error
	Expr execute.CompiledExpression
}

func NewFilterTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *FilterProcedureSpec) (*filterTransformation, error) {
	if len(spec.Fn.Params) != 1 {
		return nil, fmt.Errorf("filter functions should only have a single parameter, got %v", spec.Fn.Params)
	}
	objectName := spec.Fn.Params[0].Name
	references, err := execute.FindReferences(spec.Fn)
	if err != nil {
		return nil, err
	}

	valueRP := execute.Reference{objectName, "_value"}.Path()
	types := make(map[execute.ReferencePath]execute.DataType, len(references))
	for _, r := range references {
		if len(r) != 2 {
			return nil, fmt.Errorf("found invalid reference in the filter function %q", r)
		}
		rp := r.Path()
		if rp != valueRP {
			types[rp] = execute.TString
		}
	}

	ces := make(map[execute.DataType]expressionOrError, len(execute.ValueDataTypes))
	for _, typ := range execute.ValueDataTypes {
		types[valueRP] = typ
		ce, err := execute.CompileExpression(spec.Fn, types)
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
		d:          d,
		cache:      cache,
		references: references,
		scope:      make(execute.Scope),
		scopeCols:  make(map[string]int),
		ces:        ces,
	}, nil
}

func (t *filterTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *filterTransformation) Process(id execute.DatasetID, b execute.Block) error {
	builder, new := t.cache.BlockBuilder(b)
	if new {
		execute.AddBlockCols(b, builder)
	}

	// Prepare scope
	cols := b.Cols()
	valueIdx := execute.ValueIdx(cols)
	for j, c := range cols {
		if c.Label == execute.ValueColLabel {
			t.scopeCols["_value"] = valueIdx
		} else {
			for _, r := range t.references {
				if r[1] == c.Label {
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
			for _, r := range t.references {
				t.scope[r.Path()] = execute.ValueForRow(i, t.scopeCols[r[1]], rr)
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
					builder.AppendBool(j, rr.AtBool(i, j))
				case execute.TInt:
					builder.AppendInt(j, rr.AtInt(i, j))
				case execute.TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, j))
				case execute.TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, j))
				case execute.TString:
					builder.AppendString(j, rr.AtString(i, j))
				case execute.TTime:
					builder.AppendTime(j, rr.AtTime(i, j))
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
