package functions

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const WindowKind = "window"

type WindowOpSpec struct {
	Every      query.Duration    `json:"every"`
	Period     query.Duration    `json:"period"`
	Start      query.Time        `json:"start"`
	Round      query.Duration    `json:"round"`
	Triggering query.TriggerSpec `json:"triggering"`
}

func init() {
	query.RegisterOpSpec(WindowKind, newWindowOp)
	plan.RegisterProcedureSpec(WindowKind, newWindowProcedure, WindowKind)
	execute.RegisterTransformation(WindowKind, createWindowTransformation)
}

func newWindowOp() query.OperationSpec {
	return new(WindowOpSpec)
}

func (s *WindowOpSpec) Kind() query.OperationKind {
	return WindowKind
}

type WindowProcedureSpec struct {
	Window     plan.WindowSpec
	Triggering query.TriggerSpec
}

func newWindowProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	s, ok := qs.(*WindowOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	p := &WindowProcedureSpec{
		Window: plan.WindowSpec{
			Every:  s.Every,
			Period: s.Period,
			Round:  s.Round,
			Start:  s.Start,
		},
		Triggering: s.Triggering,
	}
	if p.Triggering == nil {
		p.Triggering = query.DefaultTrigger
	}
	return p, nil
}

func (s *WindowProcedureSpec) Kind() plan.ProcedureKind {
	return WindowKind
}

func (s *WindowProcedureSpec) TriggerSpec() query.TriggerSpec {
	return s.Triggering
}

func createWindowTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, now time.Time) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WindowProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := newFixedWindowTransformation(d, cache, execute.Window{
		Every:  execute.Duration(s.Window.Every),
		Period: execute.Duration(s.Window.Period),
		Round:  execute.Duration(s.Window.Round),
		Start:  execute.Time(s.Window.Start.Time(now).UnixNano()),
	})
	return t, d, nil
}

type fixedWindowTransformation struct {
	d       execute.Dataset
	cache   execute.BlockBuilderCache
	w       execute.Window
	parents []execute.DatasetID
}

func newFixedWindowTransformation(d execute.Dataset, cache execute.BlockBuilderCache, w execute.Window) execute.Transformation {
	return &fixedWindowTransformation{
		d:     d,
		cache: cache,
		w:     w,
	}
}

func (t *fixedWindowTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	tagKey := meta.Tags().Key()
	t.cache.ForEachBuilder(func(bk execute.BlockKey, bld execute.BlockBuilder) {
		if bld.Bounds().Overlaps(meta.Bounds()) && tagKey == bld.Tags().Key() {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *fixedWindowTransformation) Process(id execute.DatasetID, b execute.Block) {
	tagKey := b.Tags().Key()

	cells := b.Cells()
	cells.Do(func(cs []execute.Cell) {
		for _, c := range cs {
			found := false
			t.cache.ForEachBuilder(func(bk execute.BlockKey, bld execute.BlockBuilder) {
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

func (t *fixedWindowTransformation) getWindowBounds(time execute.Time) execute.Bounds {
	stop := time.Truncate(t.w.Every)
	stop += execute.Time(t.w.Every)
	return execute.Bounds{
		Stop:  stop,
		Start: stop - execute.Time(t.w.Period),
	}
}

func (t *fixedWindowTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.d.UpdateWatermark(mark)
}
func (t *fixedWindowTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *fixedWindowTransformation) Finish(id execute.DatasetID) {
	t.d.Finish()
}
func (t *fixedWindowTransformation) SetParents(ids []execute.DatasetID) {
	t.parents = ids
}
