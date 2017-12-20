package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
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

func (s *FilterProcedureSpec) PushDownRules() []plan.PushDownRule {
	return []plan.PushDownRule{
		{
			Root:    FromKind,
			Through: []plan.ProcedureKind{GroupKind, LimitKind, RangeKind},
			Match: func(spec plan.ProcedureSpec) bool {
				// TODO(nathanielc): Remove once row functions support calling functions
				if _, ok := s.Fn.Body.(ast.Expression); !ok {
					return false
				}
				fs := spec.(*FromProcedureSpec)
				if fs.Filter != nil {
					if _, ok := fs.Filter.Body.(ast.Expression); !ok {
						return false
					}
				}
				return true
			},
		},
		{
			Root:    FilterKind,
			Through: []plan.ProcedureKind{GroupKind, LimitKind, RangeKind},
			Match: func(spec plan.ProcedureSpec) bool {
				// TODO(nathanielc): Remove once row functions support calling functions
				if _, ok := s.Fn.Body.(ast.Expression); !ok {
					return false
				}
				fs := spec.(*FilterProcedureSpec)
				if _, ok := fs.Fn.Body.(ast.Expression); !ok {
					return false
				}
				return true
			},
		},
	}
}

func (s *FilterProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	switch spec := root.Spec.(type) {
	case *FromProcedureSpec:
		if spec.FilterSet {
			spec.Filter = mergeArrowFunction(spec.Filter, s.Fn)
			return
		}
		spec.FilterSet = true
		spec.Filter = s.Fn
	case *FilterProcedureSpec:
		spec.Fn = mergeArrowFunction(spec.Fn, s.Fn)
	}
}

func mergeArrowFunction(a, b *ast.ArrowFunctionExpression) *ast.ArrowFunctionExpression {
	fn := a.Copy().(*ast.ArrowFunctionExpression)

	aExp, aOK := a.Body.(ast.Expression)
	bExp, bOK := b.Body.(ast.Expression)

	if aOK && bOK {
		fn.Body = &ast.LogicalExpression{
			Operator: ast.AndOperator,
			Left:     aExp,
			Right:    bExp,
		}
		return fn
	}

	// TODO(nathanielc): This code is unreachable while the current PushDownRule Match function is inplace.

	and := &ast.LogicalExpression{
		Operator: ast.AndOperator,
		Left:     aExp,
		Right:    bExp,
	}

	// Create pass through arguments expression
	passThroughArgs := &ast.ObjectExpression{
		Properties: make([]*ast.Property, len(a.Params)),
	}
	for i, p := range a.Params {
		passThroughArgs.Properties[i] = &ast.Property{
			Key:   p.Key,
			Value: p.Key,
		}
	}

	if !aOK {
		// Rewrite left expression as a function call.
		and.Left = &ast.CallExpression{
			Callee:    a.Copy().(*ast.ArrowFunctionExpression),
			Arguments: []ast.Expression{passThroughArgs.Copy().(*ast.ObjectExpression)},
		}
	}
	if !bOK {
		// Rewrite right expression as a function call.
		and.Right = &ast.CallExpression{
			Callee:    b.Copy().(*ast.ArrowFunctionExpression),
			Arguments: []ast.Expression{passThroughArgs.Copy().(*ast.ObjectExpression)},
		}
	}
	return fn
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

	fn *execute.RowPredicateFn
}

func NewFilterTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *FilterProcedureSpec) (*filterTransformation, error) {
	fn, err := execute.NewRowPredicateFn(spec.Fn)
	if err != nil {
		return nil, err
	}

	return &filterTransformation{
		d:     d,
		cache: cache,
		fn:    fn,
	}, nil
}

func (t *filterTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *filterTransformation) Process(id execute.DatasetID, b execute.Block) error {
	log.Println("process")
	builder, new := t.cache.BlockBuilder(b)
	if new {
		execute.AddBlockCols(b, builder)
	}

	// Prepare the function for the column types.
	cols := b.Cols()
	if err := t.fn.Prepare(cols); err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// Append only matching rows to block
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			if pass, err := t.fn.Eval(i, rr); err != nil {
				log.Printf("failed to evaluate filter expression: %v", err)
				continue
			} else if !pass {
				// No match, skipping
				continue
			}
			for j, c := range cols {
				if c.Common {
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
