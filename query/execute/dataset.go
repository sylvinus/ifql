package execute

import (
	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

type AccumulationMode int

const (
	DiscardingMode AccumulationMode = iota
	AccumulatingMode
	AccumulatingRetractingMode
)

type BlockKey string

type DatasetID uuid.UUID

var ZeroDatasetID DatasetID

func (id DatasetID) IsZero() bool {
	return id == ZeroDatasetID
}

type Dataset interface {
	Node

	// BlockBuilder will return the builder for the specified block.
	// If no builder exists, one will be created.
	BlockBuilder(meta BlockMetadata) BlockBuilder
	ForEachBuilder(func(BlockKey, BlockBuilder))

	TriggerBlock(key BlockKey)
	ExpireBlock(key BlockKey)
	RetractBlock(key BlockKey)

	UpdateProcessingTime(t Time)
	UpdateWatermark(mark Time)
	Finish()

	setTriggerSpec(t query.TriggerSpec)
}

func newDataset(id DatasetID, accMode AccumulationMode) Dataset {
	return &dataset{
		id:      id,
		accMode: accMode,
		blocks:  make(map[BlockKey]blockState),
	}
}

type dataset struct {
	id DatasetID
	// Stateful triggers per stop time bound?
	blocks  map[BlockKey]blockState
	ts      []Transformation
	accMode AccumulationMode

	triggerSpec    query.TriggerSpec
	watermark      Time
	processingTime Time
}

type blockState struct {
	builder BlockBuilder
	trigger Trigger
}

func (d *dataset) addTransformation(t Transformation) {
	d.ts = append(d.ts, t)
}
func (d *dataset) setTriggerSpec(ts query.TriggerSpec) {
	d.triggerSpec = ts
}

func (d *dataset) UpdateWatermark(mark Time) {
	d.watermark = mark
	d.evalTriggers()
	for _, t := range d.ts {
		t.UpdateWatermark(mark)
	}
}

func (d *dataset) UpdateProcessingTime(time Time) {
	d.processingTime = time
	d.evalTriggers()
	for _, t := range d.ts {
		t.UpdateProcessingTime(time)
	}
}

func (d *dataset) evalTriggers() {
	c := TriggerContext{
		Watermark:             d.watermark,
		CurrentProcessingTime: d.processingTime,
	}
	for bk, b := range d.blocks {
		c.Builder = b.builder

		if b.trigger.Triggered(c) {
			d.TriggerBlock(bk)
		}
		if b.trigger.Finished() {
			d.ExpireBlock(bk)
		}
	}
}

func (d *dataset) BlockBuilder(meta BlockMetadata) BlockBuilder {
	key := ToBlockKey(meta)
	b, ok := d.blocks[key]
	if !ok {
		builder := newRowListBlockBuilder()
		builder.SetTags(meta.Tags())
		builder.SetBounds(meta.Bounds())
		t := newTriggerFromSpec(d.triggerSpec)
		b = blockState{
			builder: builder,
			trigger: t,
		}
		d.blocks[key] = b
	}
	return b.builder
}

func (d *dataset) ForEachBuilder(f func(BlockKey, BlockBuilder)) {
	for k, b := range d.blocks {
		f(k, b.builder)
	}
}

func (d *dataset) ExpireBlock(key BlockKey) {
	delete(d.blocks, key)
}

func (d *dataset) TriggerBlock(key BlockKey) {
	b := d.blocks[key].builder.Block()
	switch d.accMode {
	case DiscardingMode:
		for _, t := range d.ts {
			t.Process(d.id, b)
		}
		d.blocks[key].builder.ClearData()
	case AccumulatingRetractingMode:
		for _, t := range d.ts {
			t.RetractBlock(b)
		}
		fallthrough
	case AccumulatingMode:
		for _, t := range d.ts {
			t.Process(d.id, b)
		}
	}
}

func (d *dataset) RetractBlock(key BlockKey) {
	for _, t := range d.ts {
		t.RetractBlock(d.blocks[key].builder)
	}
}

func (d *dataset) Finish() {
	for k := range d.blocks {
		d.TriggerBlock(k)
	}
	for _, t := range d.ts {
		t.Finish()
	}
}
