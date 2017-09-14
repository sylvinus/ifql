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

type DatasetID uuid.UUID

var ZeroDatasetID DatasetID

func (id DatasetID) IsZero() bool {
	return id == ZeroDatasetID
}

type Dataset interface {
	Node

	RetractBlock(key BlockKey)
	UpdateProcessingTime(t Time)
	UpdateWatermark(mark Time)
	Finish()

	TriggerBlock(BlockKey)

	setTriggerSpec(t query.TriggerSpec)
}

type DataCache interface {
	BlockMetadata(BlockKey) BlockMetadata
	Block(BlockKey) Block

	ForEach(func(BlockKey))
	ForEachWithContext(func(BlockKey, Trigger, BlockContext))

	DiscardBlock(BlockKey)
	ExpireBlock(BlockKey)

	setTriggerSpec(t query.TriggerSpec)
}

type dataset struct {
	id DatasetID

	ts      []Transformation
	accMode AccumulationMode

	watermark      Time
	processingTime Time

	cache DataCache
}

func newDataset(id DatasetID, accMode AccumulationMode, cache DataCache) *dataset {
	return &dataset{
		id:      id,
		accMode: accMode,
		cache:   cache,
	}
}

func (d *dataset) addTransformation(t Transformation) {
	d.ts = append(d.ts, t)
}

func (d *dataset) setTriggerSpec(spec query.TriggerSpec) {
	d.cache.setTriggerSpec(spec)
}

func (d *dataset) UpdateWatermark(mark Time) {
	d.watermark = mark
	d.evalTriggers()
	for _, t := range d.ts {
		t.UpdateWatermark(d.id, mark)
	}
}

func (d *dataset) UpdateProcessingTime(time Time) {
	d.processingTime = time
	d.evalTriggers()
	for _, t := range d.ts {
		t.UpdateProcessingTime(d.id, time)
	}
}

func (d *dataset) evalTriggers() {
	d.cache.ForEachWithContext(func(bk BlockKey, trigger Trigger, bc BlockContext) {
		c := TriggerContext{
			Block:                 bc,
			Watermark:             d.watermark,
			CurrentProcessingTime: d.processingTime,
		}

		if trigger.Triggered(c) {
			d.TriggerBlock(bk)
		}
		if trigger.Finished() {
			d.cache.ExpireBlock(bk)
		}
	})
}

func (d *dataset) TriggerBlock(key BlockKey) {
	b := d.cache.Block(key)
	switch d.accMode {
	case DiscardingMode:
		for _, t := range d.ts {
			t.Process(d.id, b)
		}
		d.cache.DiscardBlock(key)
	case AccumulatingRetractingMode:
		for _, t := range d.ts {
			t.RetractBlock(d.id, b)
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
		t.RetractBlock(d.id, d.cache.BlockMetadata(key))
	}
}

func (d *dataset) Finish() {
	d.cache.ForEach(func(bk BlockKey) {
		d.TriggerBlock(bk)
		d.cache.ExpireBlock(bk)
	})
	for _, t := range d.ts {
		t.Finish(d.id)
	}
}

type blockBuilderDataset struct {
	// Stateful triggers per stop time bound?
	blocks map[BlockKey]blockState

	triggerSpec query.TriggerSpec
}

func newBlockBuilderDataset() *blockBuilderDataset {
	return &blockBuilderDataset{
		blocks: make(map[BlockKey]blockState),
	}
}

type blockState struct {
	builder BlockBuilder
	trigger Trigger
}

func (d *blockBuilderDataset) setTriggerSpec(ts query.TriggerSpec) {
	d.triggerSpec = ts
}

func (d *blockBuilderDataset) Block(key BlockKey) Block {
	return d.blocks[key].builder.Block()
}
func (d *blockBuilderDataset) BlockMetadata(key BlockKey) BlockMetadata {
	return d.blocks[key].builder
}

// BlockBuilder will return the builder for the specified block.
// If no builder exists, one will be created.
func (d *blockBuilderDataset) BlockBuilder(meta BlockMetadata) BlockBuilder {
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

func (d *blockBuilderDataset) ForEachBuilder(f func(BlockKey, BlockBuilder)) {
	for k, b := range d.blocks {
		f(k, b.builder)
	}
}

func (d *blockBuilderDataset) DiscardBlock(key BlockKey) {
	d.blocks[key].builder.ClearData()
}
func (d *blockBuilderDataset) ExpireBlock(key BlockKey) {
	delete(d.blocks, key)
}

func (d *blockBuilderDataset) ForEach(f func(BlockKey)) {
	for bk := range d.blocks {
		f(bk)
	}
}

func (d *blockBuilderDataset) ForEachWithContext(f func(BlockKey, Trigger, BlockContext)) {
	for bk, b := range d.blocks {
		f(bk, b.trigger, BlockContext{
			Bounds: b.builder.Bounds(),
			Count:  b.builder.NRows(),
		})
	}
}
