package execute

type aggregateTransformation struct {
	d     Dataset
	cache BlockBuilderCache

	parents []DatasetID
}

func NewAggregateTransformation(id DatasetID, mode AccumulationMode, bounds Bounds, agg AggFunc) (*aggregateTransformation, Dataset) {
	bbCache := NewBlockBuilderCache()
	cache := aggBlockCache{
		blockBuilderCache: bbCache,
		agg:               agg,
		bounds:            bounds,
	}
	d := NewDataset(id, mode, cache)
	return &aggregateTransformation{
		d:     d,
		cache: cache,
	}, d
}

func (t *aggregateTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	//TODO(nathanielc): Store intermediate state for retractions
	key := ToBlockKey(meta)
	t.d.RetractBlock(key)
}

func (t *aggregateTransformation) Process(id DatasetID, b Block) {
	builder, new := t.cache.BlockBuilder(b)
	if new {
		builder.AddCol(TimeCol)
		builder.AddCol(ValueCol)
	}

	// Append block to builder
	times := b.Times()
	times.DoTime(func(vs []Time) {
		builder.AppendTimes(0, vs)
	})
	values := b.Values()
	values.DoFloat(func(vs []float64) {
		builder.AppendFloats(1, vs)
	})

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

type aggBlockCache struct {
	*blockBuilderCache
	agg    AggFunc
	bounds Bounds
}

func (c aggBlockCache) Block(key BlockKey) Block {
	//TODO(nathanielc): Add types to the aggFuncs

	b := c.blockBuilderCache.Block(key)

	values := b.Values()
	values.DoFloat(c.agg.Do)

	builder := NewColListBlockBuilder()
	builder.SetBounds(c.bounds)
	builder.SetTags(b.Tags())
	builder.AddCol(TimeCol)
	builder.AddCol(ValueCol)
	builder.AppendTime(0, b.Bounds().Stop)
	builder.AppendFloat(1, c.agg.Value())

	c.agg.Reset()
	return builder.Block()
}
