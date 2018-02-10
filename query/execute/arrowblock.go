package execute

import (
	"sync/atomic"

	"github.com/influxdata/arrow/array"
)

type ArrowBlock struct {
	refCount int64
	bounds   Bounds
	tags     Tags
	colMeta  []ColMeta
	cols     []interface{}
	nrows    int
}

func newArrowBlock() *ArrowBlock {
	return &ArrowBlock{refCount: 1}
}

func (b *ArrowBlock) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}

func (b *ArrowBlock) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		b.ClearData()
	}
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

func (b *ArrowBlock) ClearData() {
	for _, arrays := range b.cols {
		switch s := arrays.(type) {
		case []*array.Timestamp:
			for _, c := range s {
				c.Release()
			}
		case []*array.Int64:
			for _, c := range s {
				c.Release()
			}
		case []*array.Float64:
			for _, c := range s {
				c.Release()
			}
		case []*array.Uint64:
			for _, c := range s {
				c.Release()
			}
		case []*array.Boolean:
			for _, c := range s {
				c.Release()
			}
		case []*array.Binary:
			for _, c := range s {
				c.Release()
			}
		}
	}
	b.cols = nil
}

type arrayValueIterator struct {
	col     int
	row     int
	colMeta []ColMeta
	cols    []interface{} // weak reference; owned by ArrowBlock
	nrows   int
}

// ValueIterator funcs

func (itr *arrayValueIterator) DoBool(f func([]bool, RowReader)) {
	checkColType(itr.colMeta[itr.col], TBool)

	// TODO(sgc): callback signature will change to pass Arrow array
	var res []bool
	for row, a := range itr.cols[itr.col].([]*array.Boolean) {
		c := a.Len()
		if cap(res) < a.Len() {
			res = make([]bool, c)
		}
		res = res[:0]
		for i := 0; i < c; i++ {
			res[i] = a.Value(i)
		}
		itr.row = row
		f(res, itr)
	}
}

func (itr *arrayValueIterator) DoInt(f func([]int64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TInt)
	for row, c := range itr.cols[itr.col].([]*array.Int64) {
		itr.row = row
		f(c.Int64Values(), itr)
	}
}

func (itr *arrayValueIterator) DoUInt(f func([]uint64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TUInt)
	for row, c := range itr.cols[itr.col].([]*array.Uint64) {
		itr.row = row
		f(c.Uint64Values(), itr)
	}
}

func (itr *arrayValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TFloat)
	for row, c := range itr.cols[itr.col].([]*array.Float64) {
		itr.row = row
		f(c.Float64Values(), itr)
	}

}

func (itr *arrayValueIterator) DoString(f func([]string, RowReader)) {
	meta := itr.colMeta[itr.col]
	checkColType(itr.colMeta[itr.col], TString)
	if meta.IsTag() && meta.Common {
		value := itr.cols[itr.col].([]*commonStrColumn)[0].value
		strs := make([]string, itr.nrows)
		for i := range strs {
			strs[i] = value
		}
		f(strs, itr)
		return
	}

	// TODO(sgc): callback signature will change to pass Arrow array
	var res []string
	for row, a := range itr.cols[itr.col].([]*array.Binary) {
		c := a.Len()
		if cap(res) < a.Len() {
			res = make([]string, c)
		}
		res = res[:0]
		for i := 0; i < c; i++ {
			res[i] = a.ValueString(i)
		}
		itr.row = row
		f(res, itr)
	}
}

func (itr *arrayValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(itr.colMeta[itr.col], TTime)
	for row, c := range itr.cols[itr.col].([]*array.Timestamp) {
		itr.row = row
		f(timestampToTimeSlice(c.TimestampValues()), itr)
	}
}

// RowReader funcs

func (itr *arrayValueIterator) Cols() []ColMeta { return itr.colMeta }

func (itr *arrayValueIterator) AtBool(i, j int) bool {
	checkColType(itr.colMeta[j], TBool)
	return itr.cols[j].([]*array.Boolean)[itr.row].Value(i)
}

func (itr *arrayValueIterator) AtInt(i, j int) int64 {
	checkColType(itr.colMeta[j], TInt)
	return itr.cols[j].([]*array.Int64)[itr.row].Value(i)
}

func (itr *arrayValueIterator) AtUInt(i, j int) uint64 {
	checkColType(itr.colMeta[j], TUInt)
	return itr.cols[j].([]*array.Uint64)[itr.row].Value(i)
}

func (itr *arrayValueIterator) AtFloat(i, j int) float64 {
	checkColType(itr.colMeta[j], TFloat)
	return itr.cols[j].([]*array.Float64)[itr.row].Value(i)
}

func (itr *arrayValueIterator) AtString(i, j int) string {
	meta := itr.colMeta[j]
	checkColType(meta, TString)
	if meta.IsTag() && meta.Common {
		return itr.cols[j].([]*commonStrColumn)[0].value
	}
	return itr.cols[j].([]*array.Binary)[itr.row].ValueString(i)
}

func (itr *arrayValueIterator) AtTime(i, j int) Time {
	checkColType(itr.colMeta[j], TTime)
	return Time(itr.cols[j].([]*array.Timestamp)[itr.row].Value(i))
}
