package execute

type aggregateTransformation struct {
	d      Dataset
	cache  BlockBuilderCache
	bounds Bounds
	aggF   AggFunc

	parents []DatasetID
}

func NewAggregateTransformation(id DatasetID, mode AccumulationMode, bounds Bounds, aggF AggFunc) (*aggregateTransformation, Dataset) {
	cache := NewBlockBuilderCache()
	d := NewDataset(id, mode, cache)
	return &aggregateTransformation{
		d:      d,
		cache:  cache,
		bounds: bounds,
		aggF:   aggF,
	}, d
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
	}

	values := b.Values()
	values.DoFloat(func(vs []float64, _ RowReader) {
		t.aggF.Do(vs)
	})

	timeIdx := TimeIdx(builder.Cols())
	valueIdx := ValueIdx(builder.Cols())

	builder.AppendTime(timeIdx, b.Bounds().Stop)
	builder.AppendFloat(valueIdx, t.aggF.Value())
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
