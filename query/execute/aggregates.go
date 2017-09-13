package execute

type aggregateTransformation struct {
	d   Dataset
	agg aggFunc

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

func (t *aggregateTransformation) Process(id DatasetID, b Block) {
	builder := t.d.BlockBuilder(b)

	values := b.Values()
	values.Do(t.agg.Do)

	builder.SetTags(b.Tags())
	builder.SetBounds(b.Bounds())
	builder.AddCol(b.Bounds().Stop)
	builder.AddRow(b.Tags())
	builder.Set(0, 0, t.agg.Value())
	t.agg.Reset()
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
