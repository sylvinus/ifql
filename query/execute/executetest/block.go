package executetest

import "github.com/influxdata/ifql/query/execute"

type Block struct {
	Bnds    execute.Bounds
	Tgs     execute.Tags
	ColMeta []execute.ColMeta
	// Data is a list of rows, i.e. Data[row][col]
	// Each row must be a list with length equal to len(ColMeta)
	Data [][]interface{}
}

func (b *Block) Bounds() execute.Bounds {
	return b.Bnds
}

func (b *Block) Tags() execute.Tags {
	return b.Tgs
}

func (b *Block) Cols() []execute.ColMeta {
	return b.ColMeta
}

func (b *Block) Col(c int) execute.ValueIterator {
	return &ValueIterator{col: c, b: b}
}

func (b *Block) Times() execute.ValueIterator {
	timeIdx := execute.TimeIdx(b.ColMeta)
	return b.Col(timeIdx)
}

func (b *Block) Values() execute.ValueIterator {
	valueIdx := execute.ValueIdx(b.ColMeta)
	return b.Col(valueIdx)
}

type ValueIterator struct {
	col int
	b   *Block

	row int
}

func (v *ValueIterator) DoFloat(f func([]float64, execute.RowReader)) {
	for v.row = 0; v.row < len(v.b.Data); v.row++ {
		f([]float64{v.b.Data[v.row][v.col].(float64)}, v)
	}
}

func (v *ValueIterator) DoString(f func([]string, execute.RowReader)) {
	for v.row = 0; v.row < len(v.b.Data); v.row++ {
		f([]string{v.b.Data[v.row][v.col].(string)}, v)
	}
}

func (v *ValueIterator) DoTime(f func([]execute.Time, execute.RowReader)) {
	for v.row = 0; v.row < len(v.b.Data); v.row++ {
		f([]execute.Time{v.b.Data[v.row][v.col].(execute.Time)}, v)
	}
}

func (v *ValueIterator) AtFloat(i int, j int) float64 {
	return v.b.Data[v.row][j].(float64)
}

func (v *ValueIterator) AtString(i int, j int) string {
	return v.b.Data[v.row][j].(string)
}

func (v *ValueIterator) AtTime(i int, j int) execute.Time {
	return v.b.Data[v.row][j].(execute.Time)
}

func BlocksFromCache(c execute.BlockBuilderCache) []*Block {
	var blocks []*Block
	c.ForEachBuilder(func(_ execute.BlockKey, builder execute.BlockBuilder) {
		b := builder.Block()
		blocks = append(blocks, ConvertBlock(b))
	})
	return blocks
}

func ConvertBlock(b execute.Block) *Block {
	blk := &Block{
		Bnds:    b.Bounds(),
		Tgs:     b.Tags().Copy(),
		ColMeta: b.Cols(),
	}

	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			row := make([]interface{}, len(blk.ColMeta))
			for j, c := range blk.ColMeta {
				var v interface{}
				switch c.Type {
				case execute.TTime:
					v = rr.AtTime(i, j)
				case execute.TString:
					v = rr.AtString(i, j)
				case execute.TFloat:
					v = rr.AtFloat(i, j)
				}
				row[j] = v
			}
			blk.Data = append(blk.Data, row)
		}
	})
	return blk
}
