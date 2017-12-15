package execute

type aggregateTransformation struct {
	d      Dataset
	cache  BlockBuilderCache
	bounds Bounds
	agg    Aggregate
}

func NewAggregateTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, agg Aggregate) *aggregateTransformation {
	return &aggregateTransformation{
		d:      d,
		cache:  c,
		bounds: bounds,
		agg:    agg,
	}
}

func NewAggregateTransformationAndDataset(id DatasetID, mode AccumulationMode, bounds Bounds, agg Aggregate, a *Allocator) (*aggregateTransformation, Dataset) {
	cache := NewBlockBuilderCache(a)
	d := NewDataset(id, mode, cache)
	return NewAggregateTransformation(d, cache, bounds, agg), d
}

func (t *aggregateTransformation) RetractBlock(id DatasetID, meta BlockMetadata) error {
	//TODO(nathanielc): Store intermediate state for retractions
	key := ToBlockKey(meta)
	return t.d.RetractBlock(key)
}

func (t *aggregateTransformation) Process(id DatasetID, b Block) error {
	cols := b.Cols()
	valueCol := ValueCol(cols)

	values := b.Values()
	var vf ValueFunc
	switch valueCol.Type {
	case TBool:
		f := t.agg.NewBoolAgg()
		values.DoBool(func(vs []bool, _ RowReader) {
			f.DoBool(vs)
		})
		vf = f
	case TInt:
		f := t.agg.NewIntAgg()
		values.DoInt(func(vs []int64, _ RowReader) {
			f.DoInt(vs)
		})
		vf = f
	case TUInt:
		f := t.agg.NewUIntAgg()
		values.DoUInt(func(vs []uint64, _ RowReader) {
			f.DoUInt(vs)
		})
		vf = f
	case TFloat:
		f := t.agg.NewFloatAgg()
		values.DoFloat(func(vs []float64, _ RowReader) {
			f.DoFloat(vs)
		})
		vf = f
	case TString:
		f := t.agg.NewStringAgg()
		values.DoString(func(vs []string, _ RowReader) {
			f.DoString(vs)
		})
		vf = f
	}

	builder, new := t.cache.BlockBuilder(blockMetadata{
		bounds: t.bounds,
		tags:   b.Tags(),
	})
	if new {
		builder.AddCol(TimeCol)
		builder.AddCol(ColMeta{
			Label: ValueColLabel,
			Type:  vf.Type(),
		})
		AddTags(b.Tags(), builder)
	}

	builder.AppendTime(0, b.Bounds().Stop)
	switch vf.Type() {
	case TBool:
		v := vf.(BoolValueFunc)
		builder.AppendBool(1, v.ValueBool())
	case TInt:
		v := vf.(IntValueFunc)
		builder.AppendInt(1, v.ValueInt())
	case TUInt:
		v := vf.(UIntValueFunc)
		builder.AppendUInt(1, v.ValueUInt())
	case TFloat:
		v := vf.(FloatValueFunc)
		builder.AppendFloat(1, v.ValueFloat())
	case TString:
		v := vf.(StringValueFunc)
		builder.AppendString(1, v.ValueString())
	}
	return nil
}

func (t *aggregateTransformation) UpdateWatermark(id DatasetID, mark Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *aggregateTransformation) UpdateProcessingTime(id DatasetID, pt Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *aggregateTransformation) Finish(id DatasetID, err error) {
	t.d.Finish(err)
}

type Aggregate interface {
	NewBoolAgg() DoBoolAgg
	NewIntAgg() DoIntAgg
	NewUIntAgg() DoUIntAgg
	NewFloatAgg() DoFloatAgg
	NewStringAgg() DoStringAgg
}

type ValueFunc interface {
	Type() DataType
}
type DoBoolAgg interface {
	ValueFunc
	DoBool([]bool)
}
type DoFloatAgg interface {
	ValueFunc
	DoFloat([]float64)
}
type DoIntAgg interface {
	ValueFunc
	DoInt([]int64)
}
type DoUIntAgg interface {
	ValueFunc
	DoUInt([]uint64)
}
type DoStringAgg interface {
	ValueFunc
	DoString([]string)
}

type BoolValueFunc interface {
	ValueBool() bool
}
type FloatValueFunc interface {
	ValueFloat() float64
}
type IntValueFunc interface {
	ValueInt() int64
}
type UIntValueFunc interface {
	ValueUInt() uint64
}
type StringValueFunc interface {
	ValueString() string
}
