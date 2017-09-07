package execute

import (
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Transformation interface {
	RetractBlock(meta BlockMetadata)
	Process(b Block)
	UpdateWatermark(t Time)
	UpdateProcessingTime(t Time)
	Finish()
}

func transformationFromProcedureSpec(d Dataset, spec plan.ProcedureSpec, now time.Time) Transformation {
	switch s := spec.(type) {
	case *plan.SumProcedureSpec:
		return &aggregateTransformation{
			d:  d,
			at: sumAT{},
		}
	case *plan.MergeProcedureSpec:
		return newMergeTransformation(d, s)
	case *plan.WindowProcedureSpec:
		return newFixedWindowTransformation(d, Window{
			Every:  Duration(s.Window.Every),
			Period: Duration(s.Window.Period),
			Round:  Duration(s.Window.Round),
			Start:  Time(s.Window.Start.Time(now).UnixNano()),
		})
	default:
		//TODO add proper error handling
		panic(fmt.Sprintf("unsupported procedure %v", spec.Kind()))
	}
}

type fixedWindowTransformation struct {
	d Dataset
	w Window
}

func newFixedWindowTransformation(d Dataset, w Window) Transformation {
	return &fixedWindowTransformation{
		d: d,
		w: w,
	}
}

func (t *fixedWindowTransformation) RetractBlock(meta BlockMetadata) {
	tagKey := meta.Tags().Key()
	t.d.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
		if bld.Bounds().Overlaps(meta.Bounds()) && tagKey == bld.Tags().Key() {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *fixedWindowTransformation) Process(b Block) {
	tagKey := b.Tags().Key()

	cells := b.Cells()
	for c, ok := cells.NextCell(); ok; c, ok = cells.NextCell() {
		found := false
		t.d.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
			if bld.Bounds().Contains(c.Time) && tagKey == bld.Tags().Key() {
				bld.AddCell(c)
				found = true
			}
		})
		if !found {
			builder := t.d.BlockBuilder(blockMetadata{
				tags:   b.Tags(),
				bounds: t.getWindowBounds(c.Time),
			})
			builder.AddCell(c)
		}
	}
}

func (t *fixedWindowTransformation) getWindowBounds(time Time) Bounds {
	stop := time.Truncate(t.w.Every)
	stop += Time(t.w.Every)
	return Bounds{
		Stop:  stop,
		Start: stop - Time(t.w.Period),
	}
}

func (t *fixedWindowTransformation) UpdateWatermark(mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *fixedWindowTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *fixedWindowTransformation) Finish() {
	t.d.Finish()
}

type aggregateTransformation struct {
	d  Dataset
	at AggregateTransformation

	trigger Trigger
}

func (t *aggregateTransformation) setTrigger(trigger Trigger) {
	t.trigger = trigger
}

func (t *aggregateTransformation) IsPerfect() bool {
	return false
}

func (t *aggregateTransformation) RetractBlock(meta BlockMetadata) {
	key := ToBlockKey(meta)
	t.d.RetractBlock(key)
}

func (t *aggregateTransformation) Process(b Block) {
	builder := t.d.BlockBuilder(b)
	t.at.Do(b, builder)
}

func (t *aggregateTransformation) UpdateWatermark(mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *aggregateTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *aggregateTransformation) Finish() {
	t.d.Finish()
}

type mergeTransformation struct {
	d    Dataset
	keys []string
}

func newMergeTransformation(d Dataset, spec *plan.MergeProcedureSpec) *mergeTransformation {
	sort.Strings(spec.Keys)
	return &mergeTransformation{
		d:    d,
		keys: spec.Keys,
	}
}

func (t *mergeTransformation) RetractBlock(meta BlockMetadata) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.d.ForEachBuilder(func(bk BlockKey, builder BlockBuilder) {
		if meta.Bounds().Equal(builder.Bounds()) {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *mergeTransformation) Process(b Block) {
	builder := t.d.BlockBuilder(blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	})
	cells := b.Cells()
	for c, ok := cells.NextCell(); ok; c, ok = cells.NextCell() {
		builder.AddCell(c)
	}
}

func (t *mergeTransformation) UpdateWatermark(mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *mergeTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *mergeTransformation) Finish() {
	t.d.Finish()
}
