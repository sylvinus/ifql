package execute

import "github.com/influxdata/arrow/array"

type ArrowBlock struct {
	bounds  Bounds
	tags    Tags
	colMeta []ColMeta
	cols    []interface{}
	nrows   int
}

func (b *ArrowBlock) Bounds() Bounds  { return b.bounds }
func (b *ArrowBlock) Tags() Tags      { return b.tags }
func (b *ArrowBlock) Cols() []ColMeta { return b.colMeta }

func (b *ArrowBlock) Col(c int) ValueIterator {
	return &arrayValueIterator{
		col:     c,
		colMeta: b.colMeta,
		cols:    b.cols,
		nrows:   b.nrows,
	}
}

func (b *ArrowBlock) Times() ValueIterator {
	j := TimeIdx(b.colMeta)
	if j >= 0 {
		return &arrayValueIterator{
			col:     j,
			colMeta: b.colMeta,
			cols:    b.cols,
			nrows:   b.nrows,
		}
	}
	return nil
}

func (b *ArrowBlock) Values() (ValueIterator, error) {
	j := ValueIdx(b.colMeta)
	if j >= 0 {
		return &arrayValueIterator{
			col:     j,
			colMeta: b.colMeta,
			cols:    b.cols,
			nrows:   b.nrows,
		}, nil
	}
	return nil, NoDefaultValueColumn
}

func (b *ArrowBlock) Retain()  { panic("implement me") }
func (b *ArrowBlock) Release() { panic("implement me") }

type arrayValueIterator struct {
	col     int
	colMeta []ColMeta
	cols    []interface{}
	nrows   int
}

// ValueIterator funcs

func (itr *arrayValueIterator) DoBool(f func([]bool, RowReader)) {
	checkColType(itr.colMeta[itr.col], TBool)

	// TODO(sgc): callback signature will change to pass Arrow array
	a := itr.cols[itr.col].(*array.Boolean)
	c := a.Len()
	res := make([]bool, c)
	for i := 0; i < c; i++ {
		res[i] = a.Value(i)
	}
	f(res, itr)
}

func (itr *arrayValueIterator) DoInt(f func([]int64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TInt)
	f(itr.cols[itr.col].(*array.Int64).Int64Values(), itr)
}

func (itr *arrayValueIterator) DoUInt(f func([]uint64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TUInt)
	f(itr.cols[itr.col].(*array.Uint64).Uint64Values(), itr)
}

func (itr *arrayValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TFloat)
	f(itr.cols[itr.col].(*array.Float64).Float64Values(), itr)
}

func (itr *arrayValueIterator) DoString(f func([]string, RowReader)) {
	checkColType(itr.colMeta[itr.col], TString)

	// TODO(sgc): callback signature will change to pass Arrow array
	a := itr.cols[itr.col].(*array.Binary)
	c := a.Len()
	res := make([]string, c)
	for i := 0; i < c; i++ {
		res[i] = a.ValueString(i)
	}
	f(res, itr)
}

func (itr *arrayValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(itr.colMeta[itr.col], TTime)
	a := itr.cols[itr.col].(*array.Timestamp)
	f(timestampToTimeSlice(a.TimestampValues()), itr)
}

// RowReader funcs

func (itr *arrayValueIterator) Cols() []ColMeta { return itr.colMeta }

func (itr *arrayValueIterator) AtBool(i, j int) bool {
	checkColType(itr.colMeta[j], TBool)
	return itr.cols[j].(*array.Boolean).Value(i)
}

func (itr *arrayValueIterator) AtInt(i, j int) int64 {
	checkColType(itr.colMeta[j], TInt)
	return itr.cols[j].(*array.Int64).Value(i)
}

func (itr *arrayValueIterator) AtUInt(i, j int) uint64 {
	checkColType(itr.colMeta[j], TUInt)
	return itr.cols[j].(*array.Uint64).Value(i)
}

func (itr *arrayValueIterator) AtFloat(i, j int) float64 {
	checkColType(itr.colMeta[j], TFloat)
	return itr.cols[j].(*array.Float64).Value(i)
}

func (itr *arrayValueIterator) AtString(i, j int) string {
	checkColType(itr.colMeta[j], TString)
	return itr.cols[j].(*array.Binary).ValueString(i)
}

func (itr *arrayValueIterator) AtTime(i, j int) Time {
	checkColType(itr.colMeta[j], TTime)
	return Time(itr.cols[j].(*array.Timestamp).Value(i))
}
