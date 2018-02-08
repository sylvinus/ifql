package execute

import (
	"fmt"
	"unsafe"

	"github.com/influxdata/arrow"
	"github.com/influxdata/arrow/array"
	"github.com/influxdata/arrow/memory"
)

type ArrowBlockBuilder struct {
	alloc memory.Allocator

	bounds  Bounds
	tags    Tags
	nrows   int
	colMeta []ColMeta
	cols    []interface{}
}

func NewArrowBlockBuilder(alloc memory.Allocator) *ArrowBlockBuilder {
	return &ArrowBlockBuilder{alloc: alloc}
}

func (b *ArrowBlockBuilder) SetBounds(bounds Bounds) { b.bounds = bounds }
func (b *ArrowBlockBuilder) Bounds() Bounds          { return b.bounds }
func (b *ArrowBlockBuilder) Tags() Tags              { return b.tags }
func (b *ArrowBlockBuilder) NRows() int              { return b.nrows }
func (b *ArrowBlockBuilder) NCols() int              { return len(b.colMeta) }
func (b *ArrowBlockBuilder) Cols() []ColMeta         { return b.colMeta }

var (
	timestampType = &arrow.TimestampType{Unit: arrow.Nanosecond}
)

func (b *ArrowBlockBuilder) AddCol(c ColMeta) int {
	var col interface{}
	switch c.Type {
	case TBool:
		col = array.NewBooleanBuilder(b.alloc)
	case TInt:
		col = array.NewInt64Builder(b.alloc)
	case TUInt:
		col = array.NewUint64Builder(b.alloc)
	case TFloat:
		col = array.NewFloat64Builder(b.alloc)
	case TString:
		if c.Common {
			col = &commonStrColumn{
				ColMeta: c,
			}
		} else {
			col = array.NewBinaryBuilder(b.alloc, arrow.BinaryTypes.String)
		}
	case TTime:
		col = array.NewTimestampBuilder(b.alloc, timestampType)
	default:
		PanicUnknownType(c.Type)
	}
	b.colMeta = append(b.colMeta, c)
	b.cols = append(b.cols, col)
	return len(b.cols) - 1
}

func (b *ArrowBlockBuilder) SetCommonString(j int, value string) {
	meta := b.colMeta[j]
	checkColType(meta, TString)
	if !meta.Common {
		panic(fmt.Errorf("cannot set common value for column %s, column is not marked as common", meta.Label))
	}
	b.cols[j].(*commonStrColumn).value = value
	if meta.IsTag() {
		if b.tags == nil {
			b.tags = make(Tags)
		}
		b.tags[meta.Label] = value
	}
}

func (b *ArrowBlockBuilder) AppendBool(j int, value bool) {
	checkColType(b.colMeta[j], TBool)
	b.cols[j].(*array.BooleanBuilder).Append(value)
}

func (b *ArrowBlockBuilder) AppendInt(j int, value int64) {
	checkColType(b.colMeta[j], TInt)
	b.cols[j].(*array.Int64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendUInt(j int, value uint64) {
	checkColType(b.colMeta[j], TUInt)
	b.cols[j].(*array.Uint64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendFloat(j int, value float64) {
	checkColType(b.colMeta[j], TFloat)
	b.cols[j].(*array.Float64Builder).Append(value)
}

func (b *ArrowBlockBuilder) AppendString(j int, value string) {
	checkColType(b.colMeta[j], TString)
	b.cols[j].(*array.BinaryBuilder).AppendString(value)
}

func (b *ArrowBlockBuilder) AppendTime(j int, value Time) {
	checkColType(b.colMeta[j], TTime)
	b.cols[j].(*array.TimestampBuilder).Append(arrow.Timestamp(value))
}

func (b *ArrowBlockBuilder) AppendFloats(j int, values []float64) {
	checkColType(b.colMeta[j], TFloat)
	b.cols[j].(*array.Float64Builder).AppendValues(values, nil)
}

func (b *ArrowBlockBuilder) AppendStrings(j int, values []string) {
	checkColType(b.colMeta[j], TString)
	b.cols[j].(*array.BinaryBuilder).AppendStringValues(values, nil)
}

func timeToTimestampSlice(in []Time) []arrow.Timestamp {
	return *(*[]arrow.Timestamp)(unsafe.Pointer(&in))
}

func timestampToTimeSlice(in []arrow.Timestamp) []Time {
	return *(*[]Time)(unsafe.Pointer(&in))
}

func (b *ArrowBlockBuilder) AppendTimes(j int, values []Time) {
	checkColType(b.colMeta[j], TTime)
	b.cols[j].(*array.TimestampBuilder).AppendValues(timeToTimestampSlice(values), nil)
}

func (b *ArrowBlockBuilder) Sort(cols []string, desc bool) {
	panic("implement me")
}

func (b *ArrowBlockBuilder) ClearData() {
	panic("implement me")
}

func (b *ArrowBlockBuilder) Block() (Block, error) {
	panic("implement me")
}
