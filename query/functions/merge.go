package functions

import (
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const MergeKind = "merge"

type MergeOpSpec struct {
	Keys []string `json:"keys"`
}

func init() {
	query.RegisterOpSpec(MergeKind, newMergeOp)
	plan.RegisterProcedureSpec(MergeKind, newMergeProcedure, MergeKind)
	execute.RegisterTransformation(MergeKind, createMergeTransformation)
}

func newMergeOp() query.OperationSpec {
	return new(MergeOpSpec)
}

func (s *MergeOpSpec) Kind() query.OperationKind {
	return MergeKind
}

type MergeProcedureSpec struct {
	Keys []string `json:"keys"`
}

func newMergeProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MergeOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &MergeProcedureSpec{
		Keys: spec.Keys,
	}
	sort.Strings(p.Keys)
	return p, nil
}

func (s *MergeProcedureSpec) Kind() plan.ProcedureKind {
	return MergeKind
}

func createMergeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, now time.Time) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := newMergeTransformation(d, cache, s)
	return t, d, nil
}

type mergeTransformation struct {
	d       execute.Dataset
	cache   execute.BlockBuilderCache
	keys    []string
	parents []execute.DatasetID
}

func newMergeTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *MergeProcedureSpec) *mergeTransformation {
	sort.Strings(spec.Keys)
	return &mergeTransformation{
		d:     d,
		cache: cache,
		keys:  spec.Keys,
	}
}

func (t *mergeTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.cache.ForEachBuilder(func(bk execute.BlockKey, builder execute.BlockBuilder) {
		if meta.Bounds().Equal(builder.Bounds()) {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *mergeTransformation) Process(id execute.DatasetID, b execute.Block) {
	builder := t.cache.BlockBuilder(blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	})
	cells := b.Cells()
	cells.Do(func(cs []execute.Cell) {
		for _, c := range cs {
			builder.AddCell(c)
		}
	})
}

func (t *mergeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.d.UpdateWatermark(mark)
}
func (t *mergeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *mergeTransformation) Finish(id execute.DatasetID) {
	t.d.Finish()
}
func (t *mergeTransformation) SetParents(ids []execute.DatasetID) {
	t.parents = ids
}

type blockMetadata struct {
	tags   execute.Tags
	bounds execute.Bounds
}

func (m blockMetadata) Tags() execute.Tags {
	return m.tags
}
func (m blockMetadata) Bounds() execute.Bounds {
	return m.bounds
}
