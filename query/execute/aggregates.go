package execute

type aggregateTransformation struct {
	d      Dataset
	cache  BlockBuilderCache
	bounds Bounds
	aggF   AggFunc

	parents []DatasetID
}

func NewAggregateTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, aggF AggFunc) *aggregateTransformation {
	return &aggregateTransformation{
		d:      d,
		cache:  c,
		bounds: bounds,
		aggF:   aggF,
	}
}

func NewAggregateTransformationAndDataset(id DatasetID, mode AccumulationMode, bounds Bounds, aggF AggFunc) (*aggregateTransformation, Dataset) {
	cache := NewBlockBuilderCache()
	d := NewDataset(id, mode, cache)
	return NewAggregateTransformation(d, cache, bounds, aggF), d
}

func (t *aggregateTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	//TODO(nathanielc): Store intermediate state for retractions
	key := ToBlockKey(meta)
	t.d.RetractBlock(key)
}

func (t *aggregateTransformation) Process(id DatasetID, b Block) {
	builder, new := t.cache.BlockBuilder(blockMetadata{
		bounds: t.bounds,
		tags:   b.Tags(),
	})
	if new {
		builder.AddCol(TimeCol)
		builder.AddCol(ValueCol)
		AddTags(b.Tags(), builder)
	}

	values := b.Values()
	values.DoFloat(func(vs []float64, _ RowReader) {
		t.aggF.Do(vs)
	})

	builder.AppendTime(0, b.Bounds().Stop)
	builder.AppendFloat(1, t.aggF.Value())
	t.aggF.Reset()
}

func (t *aggregateTransformation) UpdateWatermark(id DatasetID, mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *aggregateTransformation) UpdateProcessingTime(id DatasetID, pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *aggregateTransformation) Finish(id DatasetID) {
	t.d.Finish()
}
func (t *aggregateTransformation) SetParents(ids []DatasetID) {
	t.parents = ids
}

type AggFunc interface {
	Do([]float64)
	Value() float64
	Reset()
}
