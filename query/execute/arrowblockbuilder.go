package execute

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/influxdata/arrow"
	"github.com/influxdata/arrow/array"
	"github.com/influxdata/arrow/memory"
)

type ArrowBlockBuilder struct {
	refCount int64
	alloc    memory.Allocator
	cols     []interface{}
	blk      *ArrowBlock
	nrows    int
}

func NewArrowBlockBuilder(alloc memory.Allocator) *ArrowBlockBuilder {
	return &ArrowBlockBuilder{refCount: 1, alloc: alloc, blk: newArrowBlock()}
}

func (b *ArrowBlockBuilder) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}

func (b *ArrowBlockBuilder) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		b.releaseData()
	}
}

func (b *ArrowBlockBuilder) SetBounds(bounds Bounds) { b.blk.bounds = bounds }
func (b *ArrowBlockBuilder) Bounds() Bounds          { return b.blk.bounds }
func (b *ArrowBlockBuilder) Tags() Tags              { return b.blk.tags }
func (b *ArrowBlockBuilder) NRows() int              { return b.nrows }
func (b *ArrowBlockBuilder) NCols() int              { return len(b.blk.colMeta) }
func (b *ArrowBlockBuilder) Cols() []ColMeta         { return b.blk.colMeta }

var (
	timestampType = &arrow.TimestampType{Unit: arrow.Nanosecond}
)

func (b *ArrowBlockBuilder) AddCol(c ColMeta) int {
	var col interface{}
	switch c.Type {
	case TBool:
		col = array.NewBooleanBuilder(b.alloc)
		b.blk.cols = append(b.blk.cols, []*array.Boolean{})
	case TInt:
		col = array.NewInt64Builder(b.alloc)
		b.blk.cols = append(b.blk.cols, []*array.Int64{})
	case TUInt:
		col = array.NewUint64Builder(b.alloc)
		b.blk.cols = append(b.blk.cols, []*array.Uint64{})
	case TFloat:
		col = array.NewFloat64Builder(b.alloc)
		b.blk.cols = append(b.blk.cols, []*array.Float64{})
	case TString:
		if c.Common {
			col = &commonStrColumn{
				ColMeta: c,
			}
			b.blk.cols = append(b.blk.cols, []*commonStrColumn{})
		} else {
			col = array.NewBinaryBuilder(b.alloc, arrow.BinaryTypes.String)
			b.blk.cols = append(b.blk.cols, []*array.Binary{})
		}
	case TTime:
		col = array.NewTimestampBuilder(b.alloc, timestampType)
		b.blk.cols = append(b.blk.cols, []*array.Timestamp{})
	default:
		PanicUnknownType(c.Type)
	}
	b.blk.colMeta = append(b.blk.colMeta, c)
	b.cols = append(b.cols, col)
	return len(b.cols) - 1
}

func (b *ArrowBlockBuilder) SetCommonString(j int, value string) {
	meta := b.blk.colMeta[j]
	checkColType(meta, TString)
	if !meta.Common {
		panic(fmt.Errorf("cannot set common value for column %s, column is not marked as common", meta.Label))
	}
	b.cols[j].(*commonStrColumn).value = value
	if meta.IsTag() {
		if b.blk.tags == nil {
			b.blk.tags = make(Tags)
		}
		b.blk.tags[meta.Label] = value
	}
}

func (b *ArrowBlockBuilder) AppendBool(j int, value bool) {
	checkColType(b.blk.colMeta[j], TBool)
	b.cols[j].(*array.BooleanBuilder).Append(value)
}

func (b *ArrowBlockBuilder) AppendInt(j int, value int64) {
	checkColType(b.blk.colMeta[j], TInt)
	b.cols[j].(*array.Int64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendUInt(j int, value uint64) {
	checkColType(b.blk.colMeta[j], TUInt)
	b.cols[j].(*array.Uint64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendFloat(j int, value float64) {
	checkColType(b.blk.colMeta[j], TFloat)
	b.cols[j].(*array.Float64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendString(j int, value string) {
	checkColType(b.blk.colMeta[j], TString)
	b.cols[j].(*array.BinaryBuilder).AppendString(value)
}

func (b *ArrowBlockBuilder) AppendTime(j int, value Time) {
	checkColType(b.blk.colMeta[j], TTime)
	b.cols[j].(*array.TimestampBuilder).Append(arrow.Timestamp(value))
	b.nrows++
}

func (b *ArrowBlockBuilder) AppendFloats(j int, values []float64) {
	checkColType(b.blk.colMeta[j], TFloat)
	b.cols[j].(*array.Float64Builder).AppendValues(values, nil)
}

func (b *ArrowBlockBuilder) AppendStrings(j int, values []string) {
	checkColType(b.blk.colMeta[j], TString)
	b.cols[j].(*array.BinaryBuilder).AppendStringValues(values, nil)
}

func timeToTimestampSlice(in []Time) []arrow.Timestamp {
	return *(*[]arrow.Timestamp)(unsafe.Pointer(&in))
}

func timestampToTimeSlice(in []arrow.Timestamp) []Time {
	return *(*[]Time)(unsafe.Pointer(&in))
}

func (b *ArrowBlockBuilder) AppendTimes(j int, values []Time) {
	checkColType(b.blk.colMeta[j], TTime)
	b.cols[j].(*array.TimestampBuilder).AppendValues(timeToTimestampSlice(values), nil)
	b.nrows += len(values)
}

func (b *ArrowBlockBuilder) Sort(cols []string, desc bool) {
	panic("implement me")
}

func (b *ArrowBlockBuilder) releaseData() {
	for i, c := range b.cols {
		if r, ok := c.(interface{ Release() }); ok {
			r.Release()
		}
		b.cols[i] = nil
	}
	b.cols = b.cols[:0]
	b.blk.Release()
	b.blk = nil
}

func (b *ArrowBlockBuilder) ClearData() {
	meta := b.blk.colMeta
	b.releaseData()

	b.blk = newArrowBlock()
	for i := range meta {
		b.AddCol(meta[i])
	}
	b.nrows = 0
}

func (b *ArrowBlockBuilder) Block() (Block, error) {
	b.blk.nrows = b.nrows
	for i, c := range b.cols {
		switch bb := c.(type) {
		case *array.TimestampBuilder:
			rows := b.blk.cols[i].([]*array.Timestamp)
			b.blk.cols[i] = append(rows, bb.NewTimestampArray())
		case *array.Float64Builder:
			rows := b.blk.cols[i].([]*array.Float64)
			b.blk.cols[i] = append(rows, bb.NewFloat64Array())
		case *array.Int64Builder:
			rows := b.blk.cols[i].([]*array.Int64)
			b.blk.cols[i] = append(rows, bb.NewInt64Array())
		case *array.Uint64Builder:
			rows := b.blk.cols[i].([]*array.Uint64)
			b.blk.cols[i] = append(rows, bb.NewUint64Array())
		case *array.BooleanBuilder:
			rows := b.blk.cols[i].([]*array.Boolean)
			b.blk.cols[i] = append(rows, bb.NewBooleanArray())
		case *array.BinaryBuilder:
			rows := b.blk.cols[i].([]*array.Binary)
			b.blk.cols[i] = append(rows, bb.NewBinaryArray())
		case *commonStrColumn:
			rows := b.blk.cols[i].([]*commonStrColumn)
			b.blk.cols[i] = append(rows, bb)
		default:
			panic(fmt.Sprintf("unexpected column type: %t", c))
		}
	}
	return b.blk, nil
}
