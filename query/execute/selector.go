package execute

type selectorTransformation struct {
	d          Dataset
	cache      BlockBuilderCache
	bounds     Bounds
	selectorF  SelectorFunc
	useRowTime bool

	parents []DatasetID
}

func NewSelectorTransformation(d Dataset, c BlockBuilderCache, bounds Bounds, selectorF SelectorFunc, useRowTime bool) *selectorTransformation {
	return &selectorTransformation{
		d:          d,
		cache:      c,
		bounds:     bounds,
		selectorF:  selectorF,
		useRowTime: useRowTime,
	}
}

func NewSelectorTransformationAndDataset(id DatasetID, mode AccumulationMode, bounds Bounds, selectorF SelectorFunc, useRowTime bool) (*selectorTransformation, Dataset) {
	cache := NewBlockBuilderCache()
	d := NewDataset(id, mode, cache)
	return NewSelectorTransformation(d, cache, bounds, selectorF, useRowTime), d
}

func (t *selectorTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	//TODO(nathanielc): Store intermediate state for retractions
	key := ToBlockKey(meta)
	t.d.RetractBlock(key)
}

func (t *selectorTransformation) Process(id DatasetID, b Block) {
	builder, new := t.cache.BlockBuilder(blockMetadata{
		bounds: t.bounds,
		tags:   b.Tags(),
	})
	if new {
		builder.AddCol(TimeCol)
		builder.AddCol(ValueCol)
		AddTags(b.Tags(), builder)
	}

	colMap := AddNewCols(b, builder)

	values := b.Values()
	values.DoFloat(t.selectorF.Do)

	offset := builder.NRows()

	rows := t.selectorF.Rows()
	AppendRows(builder, rows, colMap)
	if !t.useRowTime {
		for i := range rows {
			builder.SetTime(offset+i, 0, b.Bounds().Stop)
		}
	}

	t.selectorF.Reset()
}

func (t *selectorTransformation) UpdateWatermark(id DatasetID, mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *selectorTransformation) UpdateProcessingTime(id DatasetID, pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *selectorTransformation) Finish(id DatasetID, err error) {
	t.d.Finish(err)
}
func (t *selectorTransformation) SetParents(ids []DatasetID) {
	t.parents = ids
}

type Row struct {
	Values []interface{}
}

func AppendRows(builder BlockBuilder, rows []Row, colMap []int) {
	cols := builder.Cols()
	for j, c := range cols {
		for _, row := range rows {
			v := row.Values[colMap[j]]
			switch c.Type {
			case TString:
				builder.AppendString(j, v.(string))
			case TFloat:
				builder.AppendFloat(j, v.(float64))
			case TTime:
				builder.AppendTime(j, v.(Time))
			}
		}
	}
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

type SelectorFunc interface {
	Do([]float64, RowReader)
	Rows() []Row
	Reset()
}
