package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SortKind = "sort"

type SortOpSpec struct {
	Cols []string `json:"cols"`
	Desc bool     `json:"desc"`
}

func init() {
	ifql.RegisterFunction(SortKind, createSortOpSpec)
	query.RegisterOpSpec(SortKind, newSortOp)
	plan.RegisterProcedureSpec(SortKind, newSortProcedure, SortKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	execute.RegisterTransformation(SortKind, createSortTransformation)
}

func createSortOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(SortOpSpec)

	if value, ok := args["cols"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf("cols argument must be a list of strings got %v", value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf("cols argument must be a list of strings, got list of %v", list.Type)
		}
		spec.Cols = list.Elements.([]string)
	} else {
		//Default behavior to sort by value
		spec.Cols = []string{execute.ValueColLabel}
	}

	if value, ok := args["desc"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf("desc argument must be a boolean, got %v", value.Type)
		}
		spec.Desc = value.Value.(bool)
	}

	return spec, nil
}

func newSortOp() query.OperationSpec {
	return new(SortOpSpec)
}

func (s *SortOpSpec) Kind() query.OperationKind {
	return SortKind
}

type SortProcedureSpec struct {
	Cols []string
	Desc bool
}

func newSortProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SortOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &SortProcedureSpec{
		Cols: spec.Cols,
		Desc: spec.Desc,
	}, nil
}

func (s *SortProcedureSpec) Kind() plan.ProcedureKind {
	return SortKind
}
func (s *SortProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SortProcedureSpec)

	ns.Cols = make([]string, len(s.Cols))
	copy(ns.Cols, s.Cols)

	ns.Desc = s.Desc
	return ns
}

func createSortTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SortProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := NewSortTransformation(d, cache, s)
	return t, d, nil
}

type sortTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	cols []string
	desc bool

	colMap []int
}

func NewSortTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *SortProcedureSpec) *sortTransformation {
	return &sortTransformation{
		d:     d,
		cache: cache,
		cols:  spec.Cols,
		desc:  spec.Desc,
	}
}

func (t *sortTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *sortTransformation) Process(id execute.DatasetID, b execute.Block) {
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

	execute.AppendBlock(b, builder, t.colMap)

	builder.Sort(t.cols, t.desc)
}

func (t *sortTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.d.UpdateWatermark(mark)
}
func (t *sortTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *sortTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
func (t *sortTransformation) SetParents(ids []execute.DatasetID) {
}
