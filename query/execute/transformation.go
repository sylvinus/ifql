package execute

import (
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Transformation interface {
	RetractBlock(id DatasetID, meta BlockMetadata)
	Process(id DatasetID, b Block)
	UpdateWatermark(id DatasetID, t Time)
	UpdateProcessingTime(id DatasetID, t Time)
	Finish(id DatasetID)
	setParents(ids []DatasetID)
}

func createTransformationDatasetPair(id DatasetID, mode AccumulationMode, spec plan.ProcedureSpec, now time.Time) (Transformation, Dataset) {
	switch s := spec.(type) {
	case *plan.SumProcedureSpec:
		cache := newBlockBuilderCache()
		d := newDataset(id, mode, cache)
		t := newAggregateTransformation(d, cache, new(sumAgg))
		return t, d
	case *plan.CountProcedureSpec:
		cache := newBlockBuilderCache()
		d := newDataset(id, mode, cache)
		t := newAggregateTransformation(d, cache, new(countAgg))
		return t, d
	case *plan.MergeProcedureSpec:
		cache := newBlockBuilderCache()
		d := newDataset(id, mode, cache)
		t := newMergeTransformation(d, cache, s)
		return t, d
	case *plan.MergeJoinProcedureSpec:
		cache := newMergeJoinCache()
		d := newDataset(id, mode, cache)
		t := newMergeJoinTransformation(d, cache, s)
		return t, d
	case *plan.WindowProcedureSpec:
		cache := newBlockBuilderCache()
		d := newDataset(id, mode, cache)
		t := newFixedWindowTransformation(d, cache, Window{
			Every:  Duration(s.Window.Every),
			Period: Duration(s.Window.Period),
			Round:  Duration(s.Window.Round),
			Start:  Time(s.Window.Start.Time(now).UnixNano()),
		})
		return t, d
	default:
		//TODO add proper error handling
		panic(fmt.Sprintf("unsupported procedure %v", spec.Kind()))
	}
}

type fixedWindowTransformation struct {
	d       Dataset
	cache   BlockBuilderCache
	w       Window
	parents []DatasetID
}

func newFixedWindowTransformation(d Dataset, cache BlockBuilderCache, w Window) Transformation {
	return &fixedWindowTransformation{
		d:     d,
		cache: cache,
		w:     w,
	}
}

func (t *fixedWindowTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	tagKey := meta.Tags().Key()
	t.cache.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
		if bld.Bounds().Overlaps(meta.Bounds()) && tagKey == bld.Tags().Key() {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *fixedWindowTransformation) Process(id DatasetID, b Block) {
	tagKey := b.Tags().Key()

	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, c := range cs {
			found := false
			t.cache.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
				if bld.Bounds().Contains(c.Time) && tagKey == bld.Tags().Key() {
					bld.AddCell(c)
					found = true
				}
			})
			if !found {
				builder := t.cache.BlockBuilder(blockMetadata{
					tags:   b.Tags(),
					bounds: t.getWindowBounds(c.Time),
				})
				builder.AddCell(c)
			}
		}
	})
}

func (t *fixedWindowTransformation) getWindowBounds(time Time) Bounds {
	stop := time.Truncate(t.w.Every)
	stop += Time(t.w.Every)
	return Bounds{
		Stop:  stop,
		Start: stop - Time(t.w.Period),
	}
}

func (t *fixedWindowTransformation) UpdateWatermark(id DatasetID, mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *fixedWindowTransformation) UpdateProcessingTime(id DatasetID, pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *fixedWindowTransformation) Finish(id DatasetID) {
	t.d.Finish()
}
func (t *fixedWindowTransformation) setParents(ids []DatasetID) {
	t.parents = ids
}

type mergeTransformation struct {
	d       Dataset
	cache   BlockBuilderCache
	keys    []string
	parents []DatasetID
}

func newMergeTransformation(d Dataset, cache BlockBuilderCache, spec *plan.MergeProcedureSpec) *mergeTransformation {
	sort.Strings(spec.Keys)
	return &mergeTransformation{
		d:     d,
		cache: cache,
		keys:  spec.Keys,
	}
}

func (t *mergeTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.cache.ForEachBuilder(func(bk BlockKey, builder BlockBuilder) {
		if meta.Bounds().Equal(builder.Bounds()) {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *mergeTransformation) Process(id DatasetID, b Block) {
	builder := t.cache.BlockBuilder(blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	})
	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, c := range cs {
			builder.AddCell(c)
		}
	})
}

func (t *mergeTransformation) UpdateWatermark(id DatasetID, mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *mergeTransformation) UpdateProcessingTime(id DatasetID, pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *mergeTransformation) Finish(id DatasetID) {
	t.d.Finish()
}
func (t *mergeTransformation) setParents(ids []DatasetID) {
	t.parents = ids
}
