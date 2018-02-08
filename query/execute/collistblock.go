package execute

import (
	"fmt"
	"sort"
	"sync/atomic"
)

type ColListBlockBuilder struct {
	blk   *ColListBlock
	key   BlockKey
	alloc *ColListAllocator
}

func NewColListBlockBuilder(a *ColListAllocator) *ColListBlockBuilder {
	return &ColListBlockBuilder{
		blk:   new(ColListBlock),
		alloc: a,
	}
}

func (b ColListBlockBuilder) SetBounds(bounds Bounds) {
	b.blk.bounds = bounds
}
func (b ColListBlockBuilder) Bounds() Bounds {
	return b.blk.bounds
}

func (b ColListBlockBuilder) Tags() Tags {
	return b.blk.tags
}
func (b ColListBlockBuilder) NRows() int {
	return b.blk.nrows
}
func (b ColListBlockBuilder) NCols() int {
	return len(b.blk.cols)
}
func (b ColListBlockBuilder) Cols() []ColMeta {
	return b.blk.colMeta
}

func (b ColListBlockBuilder) AddCol(c ColMeta) int {
	var col column
	switch c.Type {
	case TBool:
		col = &boolColumn{
			ColMeta: c,
			alloc:   b.alloc,
		}
	case TInt:
		col = &intColumn{
			ColMeta: c,
			alloc:   b.alloc,
		}
	case TUInt:
		col = &uintColumn{
			ColMeta: c,
			alloc:   b.alloc,
		}
	case TFloat:
		col = &floatColumn{
			ColMeta: c,
			alloc:   b.alloc,
		}
	case TString:
		if c.Common {
			col = &commonStrColumn{
				ColMeta: c,
			}
		} else {
			col = &stringColumn{
				ColMeta: c,
				alloc:   b.alloc,
			}
		}
	case TTime:
		col = &timeColumn{
			ColMeta: c,
			alloc:   b.alloc,
		}
	default:
		PanicUnknownType(c.Type)
	}
	b.blk.colMeta = append(b.blk.colMeta, c)
	b.blk.cols = append(b.blk.cols, col)
	return len(b.blk.cols) - 1
}

func (b ColListBlockBuilder) AppendBool(j int, value bool) {
	b.checkColType(j, TBool)
	col := b.blk.cols[j].(*boolColumn)
	col.data = b.alloc.AppendBools(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendBools(j int, values []bool) {
	b.checkColType(j, TBool)
	col := b.blk.cols[j].(*boolColumn)
	col.data = b.alloc.AppendBools(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b ColListBlockBuilder) AppendInt(j int, value int64) {
	b.checkColType(j, TInt)
	col := b.blk.cols[j].(*intColumn)
	col.data = b.alloc.AppendInts(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendInts(j int, values []int64) {
	b.checkColType(j, TInt)
	col := b.blk.cols[j].(*intColumn)
	col.data = b.alloc.AppendInts(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b ColListBlockBuilder) AppendUInt(j int, value uint64) {
	b.checkColType(j, TUInt)
	col := b.blk.cols[j].(*uintColumn)
	col.data = b.alloc.AppendUInts(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendUInts(j int, values []uint64) {
	b.checkColType(j, TUInt)
	col := b.blk.cols[j].(*uintColumn)
	col.data = b.alloc.AppendUInts(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b ColListBlockBuilder) AppendFloat(j int, value float64) {
	b.checkColType(j, TFloat)
	col := b.blk.cols[j].(*floatColumn)
	col.data = b.alloc.AppendFloats(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendFloats(j int, values []float64) {
	b.checkColType(j, TFloat)
	col := b.blk.cols[j].(*floatColumn)
	col.data = b.alloc.AppendFloats(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b ColListBlockBuilder) AppendString(j int, value string) {
	meta := b.blk.cols[j].Meta()
	checkColType(meta, TString)
	if meta.Common {
		v := b.blk.cols[j].(*commonStrColumn).value
		if value != v {
			panic(fmt.Errorf("attempting to append a different value to the column %s, which has all common values", meta.Label))
		}
		return
	}
	col := b.blk.cols[j].(*stringColumn)
	col.data = b.alloc.AppendStrings(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendStrings(j int, values []string) {
	b.checkColType(j, TString)
	col := b.blk.cols[j].(*stringColumn)
	col.data = b.alloc.AppendStrings(col.data, values...)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) SetCommonString(j int, value string) {
	meta := b.blk.cols[j].Meta()
	checkColType(meta, TString)
	if !meta.Common {
		panic(fmt.Errorf("cannot set common value for column %s, column is not marked as common", meta.Label))
	}
	b.blk.cols[j].(*commonStrColumn).value = value
	if meta.IsTag() {
		if b.blk.tags == nil {
			b.blk.tags = make(Tags)
		}
		b.blk.tags[meta.Label] = value
	}
}

func (b ColListBlockBuilder) AppendTime(j int, value Time) {
	b.checkColType(j, TTime)
	col := b.blk.cols[j].(*timeColumn)
	col.data = b.alloc.AppendTimes(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b ColListBlockBuilder) AppendTimes(j int, values []Time) {
	b.checkColType(j, TTime)
	col := b.blk.cols[j].(*timeColumn)
	col.data = b.alloc.AppendTimes(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b ColListBlockBuilder) checkColType(j int, typ DataType) {
	checkColType(b.blk.colMeta[j], typ)
}

func checkColType(col ColMeta, typ DataType) {
	if col.Type != typ {
		panic(fmt.Errorf("column %s is not of type %v", col.Label, typ))
	}
}

func PanicUnknownType(typ DataType) {
	panic(fmt.Errorf("unknown type %v", typ))
}

func (b ColListBlockBuilder) Block() (Block, error) {
	// Create copy in mutable state
	return b.blk.Copy(), nil
}

// RawBlock returns the underlying block being constructed.
// The Block returned will be modified by future calls to any BlockBuilder methods.
func (b ColListBlockBuilder) RawBlock() *ColListBlock {
	// Create copy in mutable state
	return b.blk
}

func (b ColListBlockBuilder) ClearData() {
	for _, c := range b.blk.cols {
		c.Clear()
	}
	b.blk.nrows = 0
}

func (b ColListBlockBuilder) Sort(cols []string, desc bool) {
	colIdxs := make([]int, len(cols))
	for i, label := range cols {
		for j, c := range b.blk.colMeta {
			if c.Label == label {
				colIdxs[i] = j
				break
			}
		}
	}
	s := colListBlockSorter{cols: colIdxs, desc: desc, b: b.blk}
	sort.Sort(s)
}

// ColListBlock implements Block using list of columns.
// All data for the block is stored in RAM.
// As a result At* methods are provided directly on the block for easy access.
type ColListBlock struct {
	bounds Bounds
	tags   Tags

	colMeta []ColMeta
	cols    []column
	nrows   int

	refCount int32
}

func (b *ColListBlock) Retain() {
	atomic.AddInt32(&b.refCount, 1)
}

func (b *ColListBlock) Release() {
	if atomic.AddInt32(&b.refCount, -1) == 0 {
		for _, c := range b.cols {
			c.Clear()
		}
	}
}

func (b *ColListBlock) Bounds() Bounds {
	return b.bounds
}

func (b *ColListBlock) Tags() Tags {
	return b.tags
}

func (b *ColListBlock) Cols() []ColMeta {
	return b.colMeta
}
func (b ColListBlock) NRows() int {
	return b.nrows
}

func (b *ColListBlock) Col(c int) ValueIterator {
	return colListValueIterator{
		col:     c,
		colMeta: b.colMeta,
		cols:    b.cols,
		nrows:   b.nrows,
	}
}

func (b *ColListBlock) Values() (ValueIterator, error) {
	j := ValueIdx(b.colMeta)
	if j >= 0 {
		return colListValueIterator{
			col:     j,
			colMeta: b.colMeta,
			cols:    b.cols,
			nrows:   b.nrows,
		}, nil
	}
	return nil, NoDefaultValueColumn
}

func (b *ColListBlock) Times() ValueIterator {
	j := TimeIdx(b.colMeta)
	if j >= 0 {
		return colListValueIterator{
			col:     j,
			colMeta: b.colMeta,
			cols:    b.cols,
			nrows:   b.nrows,
		}
	}
	return nil
}
func (b *ColListBlock) AtBool(i, j int) bool {
	checkColType(b.colMeta[j], TBool)
	return b.cols[j].(*boolColumn).data[i]
}
func (b *ColListBlock) AtInt(i, j int) int64 {
	checkColType(b.colMeta[j], TInt)
	return b.cols[j].(*intColumn).data[i]
}
func (b *ColListBlock) AtUInt(i, j int) uint64 {
	checkColType(b.colMeta[j], TUInt)
	return b.cols[j].(*uintColumn).data[i]
}
func (b *ColListBlock) AtFloat(i, j int) float64 {
	checkColType(b.colMeta[j], TFloat)
	return b.cols[j].(*floatColumn).data[i]
}
func (b *ColListBlock) AtString(i, j int) string {
	meta := b.colMeta[j]
	checkColType(meta, TString)
	if meta.IsTag() && meta.Common {
		return b.cols[j].(*commonStrColumn).value
	}
	return b.cols[j].(*stringColumn).data[i]
}
func (b *ColListBlock) AtTime(i, j int) Time {
	checkColType(b.colMeta[j], TTime)
	return b.cols[j].(*timeColumn).data[i]
}

func (b *ColListBlock) Copy() *ColListBlock {
	cpy := new(ColListBlock)
	cpy.bounds = b.bounds
	cpy.tags = b.tags.Copy()
	cpy.nrows = b.nrows

	cpy.colMeta = make([]ColMeta, len(b.colMeta))
	copy(cpy.colMeta, b.colMeta)

	cpy.cols = make([]column, len(b.cols))
	for i, c := range b.cols {
		cpy.cols[i] = c.Copy()
	}

	return cpy
}

type colListValueIterator struct {
	col     int
	cols    []column
	colMeta []ColMeta
	nrows   int
}

func (itr colListValueIterator) Cols() []ColMeta {
	return itr.colMeta
}
func (itr colListValueIterator) DoBool(f func([]bool, RowReader)) {
	checkColType(itr.colMeta[itr.col], TBool)
	f(itr.cols[itr.col].(*boolColumn).data, itr)
}
func (itr colListValueIterator) DoInt(f func([]int64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TInt)
	f(itr.cols[itr.col].(*intColumn).data, itr)
}
func (itr colListValueIterator) DoUInt(f func([]uint64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TUInt)
	f(itr.cols[itr.col].(*uintColumn).data, itr)
}
func (itr colListValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(itr.colMeta[itr.col], TFloat)
	f(itr.cols[itr.col].(*floatColumn).data, itr)
}
func (itr colListValueIterator) DoString(f func([]string, RowReader)) {
	meta := itr.colMeta[itr.col]
	checkColType(meta, TString)
	if meta.IsTag() && meta.Common {
		value := itr.cols[itr.col].(*commonStrColumn).value
		strs := make([]string, itr.nrows)
		for i := range strs {
			strs[i] = value
		}
		f(strs, itr)
	} else {
		f(itr.cols[itr.col].(*stringColumn).data, itr)
	}
}
func (itr colListValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(itr.colMeta[itr.col], TTime)
	f(itr.cols[itr.col].(*timeColumn).data, itr)
}
func (itr colListValueIterator) AtBool(i, j int) bool {
	checkColType(itr.colMeta[j], TBool)
	return itr.cols[j].(*boolColumn).data[i]
}
func (itr colListValueIterator) AtInt(i, j int) int64 {
	checkColType(itr.colMeta[j], TInt)
	return itr.cols[j].(*intColumn).data[i]
}
func (itr colListValueIterator) AtUInt(i, j int) uint64 {
	checkColType(itr.colMeta[j], TUInt)
	return itr.cols[j].(*uintColumn).data[i]
}
func (itr colListValueIterator) AtFloat(i, j int) float64 {
	checkColType(itr.colMeta[j], TFloat)
	return itr.cols[j].(*floatColumn).data[i]
}
func (itr colListValueIterator) AtString(i, j int) string {
	meta := itr.colMeta[j]
	checkColType(meta, TString)
	if meta.IsTag() && meta.Common {
		return itr.cols[j].(*commonStrColumn).value
	}
	return itr.cols[j].(*stringColumn).data[i]
}
func (itr colListValueIterator) AtTime(i, j int) Time {
	checkColType(itr.colMeta[j], TTime)
	return itr.cols[j].(*timeColumn).data[i]
}

type colListBlockSorter struct {
	cols []int
	desc bool
	b    *ColListBlock
}

func (c colListBlockSorter) Len() int {
	return c.b.nrows
}

func (c colListBlockSorter) Less(x int, y int) (less bool) {
	for _, j := range c.cols {
		if !c.b.cols[j].Equal(x, y) {
			less = c.b.cols[j].Less(x, y)
			break
		}
	}
	if c.desc {
		less = !less
	}
	return
}

func (c colListBlockSorter) Swap(x int, y int) {
	for _, col := range c.b.cols {
		col.Swap(x, y)
	}
}

type column interface {
	Meta() ColMeta
	Clear()
	Copy() column
	Equal(i, j int) bool
	Less(i, j int) bool
	Swap(i, j int)
}

type boolColumn struct {
	ColMeta
	data  []bool
	alloc *ColListAllocator
}

func (c *boolColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *boolColumn) Clear() {
	c.alloc.FreeBools(c.data)
	c.data = nil
}
func (c *boolColumn) Copy() column {
	cpy := &boolColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}
	l := len(c.data)
	cpy.data = c.alloc.Bools(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *boolColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *boolColumn) Less(i, j int) bool {
	if c.data[i] == c.data[j] {
		return false
	}
	return c.data[i]
}
func (c *boolColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

type intColumn struct {
	ColMeta
	data  []int64
	alloc *ColListAllocator
}

func (c *intColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *intColumn) Clear() {
	c.alloc.FreeInts(c.data)
	c.data = nil
}
func (c *intColumn) Copy() column {
	cpy := &intColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}
	l := len(c.data)
	cpy.data = c.alloc.Ints(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *intColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *intColumn) Less(i, j int) bool {
	return c.data[i] < c.data[j]
}
func (c *intColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

type uintColumn struct {
	ColMeta
	data  []uint64
	alloc *ColListAllocator
}

func (c *uintColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *uintColumn) Clear() {
	c.alloc.FreeUInts(c.data)
	c.data = nil
}
func (c *uintColumn) Copy() column {
	cpy := &uintColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}
	l := len(c.data)
	cpy.data = c.alloc.UInts(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *uintColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *uintColumn) Less(i, j int) bool {
	return c.data[i] < c.data[j]
}
func (c *uintColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

type floatColumn struct {
	ColMeta
	data  []float64
	alloc *ColListAllocator
}

func (c *floatColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *floatColumn) Clear() {
	c.alloc.FreeFloats(c.data)
	c.data = nil
}
func (c *floatColumn) Copy() column {
	cpy := &floatColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}
	l := len(c.data)
	cpy.data = c.alloc.Floats(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *floatColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *floatColumn) Less(i, j int) bool {
	return c.data[i] < c.data[j]
}
func (c *floatColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

type stringColumn struct {
	ColMeta
	data  []string
	alloc *ColListAllocator
}

func (c *stringColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *stringColumn) Clear() {
	c.alloc.FreeStrings(c.data)
	c.data = nil
}
func (c *stringColumn) Copy() column {
	cpy := &stringColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}

	l := len(c.data)
	cpy.data = c.alloc.Strings(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *stringColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *stringColumn) Less(i, j int) bool {
	return c.data[i] < c.data[j]
}
func (c *stringColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

type timeColumn struct {
	ColMeta
	data  []Time
	alloc *ColListAllocator
}

func (c *timeColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *timeColumn) Clear() {
	c.alloc.FreeTimes(c.data)
	c.data = nil
}
func (c *timeColumn) Copy() column {
	cpy := &timeColumn{
		ColMeta: c.ColMeta,
		alloc:   c.alloc,
	}
	l := len(c.data)
	cpy.data = c.alloc.Times(l, l)
	copy(cpy.data, c.data)
	return cpy
}
func (c *timeColumn) Equal(i, j int) bool {
	return c.data[i] == c.data[j]
}
func (c *timeColumn) Less(i, j int) bool {
	return c.data[i] < c.data[j]
}
func (c *timeColumn) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

//commonStrColumn has the same string value for all rows
type commonStrColumn struct {
	ColMeta
	value string
}

func (c *commonStrColumn) Meta() ColMeta {
	return c.ColMeta
}
func (c *commonStrColumn) Clear() {
}
func (c *commonStrColumn) Copy() column {
	cpy := new(commonStrColumn)
	*cpy = *c
	return cpy
}
func (c *commonStrColumn) Equal(i, j int) bool {
	return true
}
func (c *commonStrColumn) Less(i, j int) bool {
	return false
}
func (c *commonStrColumn) Swap(i, j int) {}
