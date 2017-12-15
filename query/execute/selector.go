package execute

type selectorTransformation struct {
	d          Dataset
	cache      BlockBuilderCache
	bounds     Bounds
	useRowTime bool
}

type rowSelectorTransformation struct {
	selectorTransformation
	selector RowSelector
}
type indexSelectorTransformation struct {
	selectorTransformation
	selector IndexSelector
}

func NewRowSelectorTransformationAndDataset(id DatasetID, mode AccumulationMode, bounds Bounds, selector RowSelector, useRowTime bool, a *Allocator) (*rowSelectorTransformation, Dataset) {
	cache := NewBlockBuilderCache(a)
	d := NewDataset(id, mode, cache)
	return NewRowSelectorTransformation(d, cache, bounds, selector, useRowTime), d
}
func NewRowSelectorTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, selector RowSelector, useRowTime bool) *rowSelectorTransformation {
	return &rowSelectorTransformation{
		selectorTransformation: newSelectorTransformation(d, c, bounds, useRowTime),
		selector:               selector,
	}
}

func NewIndexSelectorTransformationAndDataset(id DatasetID, mode AccumulationMode, bounds Bounds, selector IndexSelector, useRowTime bool, a *Allocator) (*indexSelectorTransformation, Dataset) {
	cache := NewBlockBuilderCache(a)
	d := NewDataset(id, mode, cache)
	return NewIndexSelectorTransformation(d, cache, bounds, selector, useRowTime), d
}
func NewIndexSelectorTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, selector IndexSelector, useRowTime bool) *indexSelectorTransformation {
	return &indexSelectorTransformation{
		selectorTransformation: newSelectorTransformation(d, c, bounds, useRowTime),
		selector:               selector,
	}
}

func newSelectorTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, useRowTime bool) selectorTransformation {
	return selectorTransformation{
		d:          d,
		cache:      c,
		bounds:     bounds,
		useRowTime: useRowTime,
	}
}

func (t *selectorTransformation) RetractBlock(id DatasetID, meta BlockMetadata) error {
	//TODO(nathanielc): Store intermediate state for retractions
	key := ToBlockKey(meta)
	return t.d.RetractBlock(key)
}
func (t *selectorTransformation) UpdateWatermark(id DatasetID, mark Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *selectorTransformation) UpdateProcessingTime(id DatasetID, pt Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *selectorTransformation) Finish(id DatasetID, err error) {
	t.d.Finish(err)
}

func (t *selectorTransformation) setupBuilder(b Block) (BlockBuilder, []int, ColMeta) {
	cols := b.Cols()
	valueIdx := ValueIdx(cols)
	valueCol := cols[valueIdx]

	builder, new := t.cache.BlockBuilder(blockMetadata{
		bounds: t.bounds,
		tags:   b.Tags(),
	})
	if new {
		builder.AddCol(TimeCol)
		builder.AddCol(valueCol)
		AddTags(b.Tags(), builder)
	}

	colMap := AddNewCols(b, builder)

	return builder, colMap, valueCol
}

func (t *indexSelectorTransformation) Process(id DatasetID, b Block) error {
	builder, colMap, valueCol := t.setupBuilder(b)

	values := b.Values()
	switch valueCol.Type {
	case TBool:
		s := t.selector.NewBoolSelector()
		values.DoBool(func(vs []bool, rr RowReader) {
			selected := s.DoBool(vs)
			t.appendSelected(selected, colMap, builder, rr, b.Bounds().Stop)
		})
	case TInt:
		s := t.selector.NewIntSelector()
		values.DoInt(func(vs []int64, rr RowReader) {
			selected := s.DoInt(vs)
			t.appendSelected(selected, colMap, builder, rr, b.Bounds().Stop)
		})
	case TUInt:
		s := t.selector.NewUIntSelector()
		values.DoUInt(func(vs []uint64, rr RowReader) {
			selected := s.DoUInt(vs)
			t.appendSelected(selected, colMap, builder, rr, b.Bounds().Stop)
		})
	case TFloat:
		s := t.selector.NewFloatSelector()
		values.DoFloat(func(vs []float64, rr RowReader) {
			selected := s.DoFloat(vs)
			t.appendSelected(selected, colMap, builder, rr, b.Bounds().Stop)
		})
	case TString:
		s := t.selector.NewStringSelector()
		values.DoString(func(vs []string, rr RowReader) {
			selected := s.DoString(vs)
			t.appendSelected(selected, colMap, builder, rr, b.Bounds().Stop)
		})
	}
	return nil
}

func (t *rowSelectorTransformation) Process(id DatasetID, b Block) error {
	builder, colMap, valueCol := t.setupBuilder(b)

	values := b.Values()
	var rower Rower
	switch valueCol.Type {
	case TBool:
		s := t.selector.NewBoolSelector()
		values.DoBool(s.DoBool)
		rower = s
	case TInt:
		s := t.selector.NewIntSelector()
		values.DoInt(s.DoInt)
		rower = s
	case TUInt:
		s := t.selector.NewUIntSelector()
		values.DoUInt(s.DoUInt)
		rower = s
	case TFloat:
		s := t.selector.NewFloatSelector()
		values.DoFloat(s.DoFloat)
		rower = s
	case TString:
		s := t.selector.NewStringSelector()
		values.DoString(s.DoString)
		rower = s
	}

	rows := rower.Rows()
	t.appendRows(builder, rows, colMap, b.Bounds().Stop)
	return nil
}

func (t *indexSelectorTransformation) appendSelected(selected, colMap []int, builder BlockBuilder, rr RowReader, stop Time) {
	if len(selected) == 0 {
		return
	}
	cols := builder.Cols()
	for j, c := range cols {
		for _, i := range selected {
			switch c.Type {
			case TBool:
				builder.AppendBool(j, rr.AtBool(i, colMap[j]))
			case TInt:
				builder.AppendInt(j, rr.AtInt(i, colMap[j]))
			case TUInt:
				builder.AppendUInt(j, rr.AtUInt(i, colMap[j]))
			case TFloat:
				builder.AppendFloat(j, rr.AtFloat(i, colMap[j]))
			case TString:
				builder.AppendString(j, rr.AtString(i, colMap[j]))
			case TTime:
				time := stop
				if t.useRowTime {
					time = rr.AtTime(i, colMap[j])
				}
				builder.AppendTime(j, time)
			default:
				PanicUnknownType(c.Type)
			}
		}
	}
}

func (t *rowSelectorTransformation) appendRows(builder BlockBuilder, rows []Row, colMap []int, stop Time) {
	cols := builder.Cols()
	for j, c := range cols {
		for _, row := range rows {
			v := row.Values[colMap[j]]
			switch c.Type {
			case TBool:
				builder.AppendBool(j, v.(bool))
			case TInt:
				builder.AppendInt(j, v.(int64))
			case TUInt:
				builder.AppendUInt(j, v.(uint64))
			case TFloat:
				builder.AppendFloat(j, v.(float64))
			case TString:
				builder.AppendString(j, v.(string))
			case TTime:
				if t.useRowTime {
					builder.AppendTime(j, v.(Time))
				} else {
					builder.AppendTime(j, stop)
				}
			default:
				PanicUnknownType(c.Type)
			}
		}
	}
}

type IndexSelector interface {
	NewBoolSelector() DoBoolIndexSelector
	NewIntSelector() DoIntIndexSelector
	NewUIntSelector() DoUIntIndexSelector
	NewFloatSelector() DoFloatIndexSelector
	NewStringSelector() DoStringIndexSelector
}
type DoBoolIndexSelector interface {
	DoBool([]bool) []int
}
type DoIntIndexSelector interface {
	DoInt([]int64) []int
}
type DoUIntIndexSelector interface {
	DoUInt([]uint64) []int
}
type DoFloatIndexSelector interface {
	DoFloat([]float64) []int
}
type DoStringIndexSelector interface {
	DoString([]string) []int
}

type RowSelector interface {
	NewBoolSelector() DoBoolRowSelector
	NewIntSelector() DoIntRowSelector
	NewUIntSelector() DoUIntRowSelector
	NewFloatSelector() DoFloatRowSelector
	NewStringSelector() DoStringRowSelector
}

type Rower interface {
	Rows() []Row
}

type DoBoolRowSelector interface {
	Rower
	// What if the selector doesn't know yet and needs to wait all is finalized?
	DoBool(vs []bool, rr RowReader)
}
type DoIntRowSelector interface {
	Rower
	DoInt(vs []int64, rr RowReader)
}
type DoUIntRowSelector interface {
	Rower
	DoUInt(vs []uint64, rr RowReader)
}
type DoFloatRowSelector interface {
	Rower
	DoFloat(vs []float64, rr RowReader)
}
type DoStringRowSelector interface {
	Rower
	DoString(vs []string, rr RowReader)
}

type Row struct {
	Values []interface{}
}

func ReadRow(i int, rr RowReader) (row Row) {
	cols := rr.Cols()
	row.Values = make([]interface{}, len(cols))
	for j, c := range cols {
		switch c.Type {
		case TString:
			row.Values[j] = rr.AtString(i, j)
		case TFloat:
			row.Values[j] = rr.AtFloat(i, j)
		case TTime:
			row.Values[j] = rr.AtTime(i, j)
		}
	}
	return
}
