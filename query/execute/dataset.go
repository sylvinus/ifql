package execute

import (
	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

// Dataset represents the set of data produced by a transformation.
type Dataset interface {
	Node

	RetractBlock(key BlockKey)
	UpdateProcessingTime(t Time)
	UpdateWatermark(mark Time)
	Finish()

	SetTriggerSpec(t query.TriggerSpec)
}

// DataCache holds all working data for a transformation.
type DataCache interface {
	BlockMetadata(BlockKey) BlockMetadata
	Block(BlockKey) Block

	ForEach(func(BlockKey))
	ForEachWithContext(func(BlockKey, Trigger, BlockContext))

	DiscardBlock(BlockKey)
	ExpireBlock(BlockKey)

	SetTriggerSpec(t query.TriggerSpec)
}

type AccumulationMode int

const (
	DiscardingMode AccumulationMode = iota
	AccumulatingMode
	AccumulatingRetractingMode
)

type DatasetID uuid.UUID

func (id DatasetID) String() string {
	return uuid.UUID(id).String()
}

var ZeroDatasetID DatasetID

func (id DatasetID) IsZero() bool {
	return id == ZeroDatasetID
}

type dataset struct {
	id DatasetID

	ts      []Transformation
	accMode AccumulationMode

	watermark      Time
	processingTime Time

	cache DataCache
}

func NewDataset(id DatasetID, accMode AccumulationMode, cache DataCache) *dataset {
	return &dataset{
		id:      id,
		accMode: accMode,
		cache:   cache,
	}
}

func (d *dataset) AddTransformation(t Transformation) {
	d.ts = append(d.ts, t)
}

func (d *dataset) SetTriggerSpec(spec query.TriggerSpec) {
	d.cache.SetTriggerSpec(spec)
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
			d.triggerBlock(bk)
		}
		if trigger.Finished() {
			d.expireBlock(bk)
		}
	})
}

func (d *dataset) triggerBlock(key BlockKey) {
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

func (d *dataset) expireBlock(key BlockKey) {
	d.cache.ExpireBlock(key)
}

func (d *dataset) RetractBlock(key BlockKey) {
	d.cache.DiscardBlock(key)
	for _, t := range d.ts {
		t.RetractBlock(d.id, d.cache.BlockMetadata(key))
	}
}

func (d *dataset) Finish() {
	d.cache.ForEach(func(bk BlockKey) {
		d.triggerBlock(bk)
		d.cache.ExpireBlock(bk)
	})
	for _, t := range d.ts {
		t.Finish(d.id)
	}
}
