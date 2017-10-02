package execute

type aggregateTransformation struct {
	d     Dataset
	cache BlockBuilderCache
	agg   aggFunc

	trigger Trigger

	parents []DatasetID
}

func NewAggregateTransformation(d Dataset, cache *blockBuilderCache, agg aggFunc) *aggregateTransformation {
	return &aggregateTransformation{
		d:     d,
		cache: cache,
		agg:   agg,
	}
}

func (t *aggregateTransformation) setTrigger(trigger Trigger) {
	t.trigger = trigger
}

func (t *aggregateTransformation) IsPerfect() bool {
	return false
}

func (t *aggregateTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	key := ToBlockKey(meta)
	t.d.RetractBlock(key)
}

func (t *aggregateTransformation) Process(id DatasetID, b Block) {
	builder := t.cache.BlockBuilder(b)

	values := b.Values()
	values.Do(t.agg.Do)

	builder.AddCol(TimeCol)
	//TODO(nathanielc): Add types to the aggFuncs
	builder.AddCol(ValueCol)
	builder.AddRow()
	builder.SetTime(0, 0, b.Bounds().Stop)
	builder.SetFloat(0, 1, t.agg.Value())
	t.agg.Reset()
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

type aggFunc interface {
	Do([]float64)
	Value() float64
	Reset()
}

type sumAgg struct {
	sum float64
}

func (a *sumAgg) Do(vs []float64) {
	for _, v := range vs {
		a.sum += v
	}
}
func (a *sumAgg) Value() float64 {
	return a.sum
}
func (a *sumAgg) Reset() {
	a.sum = 0
}

type countAgg struct {
	count float64
}

func (a *countAgg) Do(vs []float64) {
	a.count += float64(len(vs))
}
func (a *countAgg) Value() float64 {
	return a.count
}
func (a *countAgg) Reset() {
	a.count = 0
}
