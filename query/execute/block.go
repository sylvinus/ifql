package execute

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/influxdata/ifql/query"
)

type BlockMetadata interface {
	Bounds() Bounds
	Tags() Tags
}

type BlockKey string

func ToBlockKey(meta BlockMetadata) BlockKey {
	// TODO: Make this not a hack
	return BlockKey(fmt.Sprintf("%s:%d-%d", meta.Tags().Key(), meta.Bounds().Start, meta.Bounds().Stop))
}

type Block interface {
	BlockMetadata

	Cols() []ColMeta
	// Col returns an iterator to consume the values for a given column.
	Col(c int) ValueIterator

	// Times returns an iterator to consume the values for the "time" column.
	Times() ValueIterator
	// Values returns an iterator to consume the values for the "value" column.
	Values() ValueIterator
}

// OneTimeBlock is a Block that permits reading data only once.
// Specifically the ValueIterator may only be consumed once from any of the columns.
type OneTimeBlock interface {
	Block
	onetime()
}

// CacheOneTimeBlock returns a block that can be read multiple times.
// If the block is not a OneTimeBlock it is returned directly.
// Otherwise its contents are read into a new block.
func CacheOneTimeBlock(b Block) Block {
	_, ok := b.(OneTimeBlock)
	if !ok {
		return b
	}
	return CopyBlock(b)
}

// CopyBlock returns a copy of the block and is OneTimeBlock safe.
func CopyBlock(b Block) Block {
	builder := NewColListBlockBuilder()
	builder.SetBounds(b.Bounds())

	cols := b.Cols()
	colMap := make([]int, len(cols))
	for j, c := range cols {
		colMap[j] = j
		builder.AddCol(c)
		if c.IsTag && c.IsCommon {
			builder.SetCommonString(j, b.Tags()[c.Label])
		}
	}

	AppendBlock(b, builder, colMap)
	return builder.Block()
}

// AddBlockCols adds the columns of b onto builder.
func AddBlockCols(b Block, builder BlockBuilder) {
	cols := b.Cols()
	for j, c := range cols {
		builder.AddCol(c)
		if c.IsTag && c.IsCommon {
			builder.SetCommonString(j, b.Tags()[c.Label])
		}
	}
}

// AppendBlock append data from block b onto builder.
// The colMap is a map of builder columnm index to block column index.
// AppendBlock is OneTimeBlock safe.
func AppendBlock(b Block, builder BlockBuilder, colMap []int) {
	times := b.Times()

	cols := builder.Cols()
	timeIdx := TimeIdx(cols)
	times.DoTime(func(ts []Time, rr RowReader) {
		builder.AppendTimes(timeIdx, ts)
		for j, c := range cols {
			if j == timeIdx || c.IsCommon {
				continue
			}
			for i := range ts {
				switch c.Type {
				case TString:
					builder.AppendString(j, rr.AtString(i, colMap[j]))
				case TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, colMap[j]))
				case TTime:
					builder.AppendTime(j, rr.AtTime(i, colMap[j]))
				}
			}
		}
	})
}

// AddTags add columns to the builder for the given tags.
// It is assumed that all tags are common to all rows of this block.
func AddTags(t Tags, b BlockBuilder) {
	keys := t.Keys()
	for _, k := range keys {
		j := b.AddCol(ColMeta{
			Label:    k,
			Type:     TString,
			IsTag:    true,
			IsCommon: true,
		})
		b.SetCommonString(j, t[k])
	}
}

func ValueIdx(cols []ColMeta) int {
	for j, c := range cols {
		if c.Label == valueColLabel {
			return j
		}
	}
	return -1
}
func TimeIdx(cols []ColMeta) int {
	for j, c := range cols {
		if c.Label == timeColLabel {
			return j
		}
	}
	return -1
}

// BlockBuilder builds blocks that can be used multiple times
type BlockBuilder interface {
	SetBounds(Bounds)

	BlockMetadata

	NRows() int
	NCols() int
	Cols() []ColMeta

	// AddCol increases the size of the block by one column.
	// The index of the column is returned.
	AddCol(ColMeta) int

	// Set sets the value at the specified coordinates
	// The rows and columns must exist before calling set, otherwise Set panics.
	SetFloat(i, j int, value float64)
	SetString(i, j int, value string)
	SetTime(i, j int, value Time)

	// SetCommonString sets a single value for the entire column.
	SetCommonString(j int, value string)

	AppendFloat(j int, value float64)
	AppendString(j int, value string)
	AppendTime(j int, value Time)

	AppendFloats(j int, values []float64)
	AppendStrings(j int, values []string)
	AppendTimes(j int, values []Time)

	// Sort the rows of the by the values of the columns in the order listed.
	Sort(cols []string, desc bool)

	// Clear removes all rows, while preserving the column meta data.
	ClearData()

	// Block returns the block that has been built.
	// Further modifications of the builder will not effect the returned block.
	Block() Block
}

type DataType int

const (
	TInvalid DataType = iota
	TTime
	TString
	TFloat
	TInt
)

func (t DataType) String() string {
	switch t {
	case TInvalid:
		return "invalid"
	case TTime:
		return "time"
	case TString:
		return "string"
	case TFloat:
		return "float"
	case TInt:
		return "int"
	default:
		return "unknown"
	}
}

type ColMeta struct {
	Label string
	Type  DataType
	IsTag bool
	// IsCommon indicates that the value for the column is shared by all rows.
	IsCommon bool
}

const (
	valueColLabel = "value"
	timeColLabel  = "time"
)

var (
	TimeCol = ColMeta{
		Label: timeColLabel,
		Type:  TTime,
	}
	ValueCol = ColMeta{
		Label: valueColLabel,
		Type:  TFloat,
	}
)

type BlockIterator interface {
	Do(f func(Block)) error
}

type ValueIterator interface {
	DoFloat(f func([]float64, RowReader))
	DoString(f func([]string, RowReader))
	DoTime(f func([]Time, RowReader))
}

type RowReader interface {
	// AtFloat returns the float value of another column and given index.
	AtFloat(i, j int) float64
	// AtString returns the string value of another column and given index.
	AtString(i, j int) string
	// AtTime returns the time value of another column and given index.
	AtTime(i, j int) Time
}

func TagsForRow(cols []ColMeta, rr RowReader, i int) Tags {
	tags := make(Tags, len(cols)-2)
	for j, c := range cols {
		if c.IsTag {
			tags[c.Label] = rr.AtString(i, j)
		}
	}
	return tags
}

type Tags map[string]string

func (t Tags) Copy() Tags {
	nt := make(Tags, len(t))
	for k, v := range t {
		nt[k] = v
	}
	return nt
}

func (t Tags) Equal(o Tags) bool {
	if len(t) != len(o) {
		return false
	}
	for k, v := range t {
		if o[k] != v {
			return false
		}
	}
	return true
}

func (t Tags) Keys() []string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type TagsKey string

func (t Tags) Key() TagsKey {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return TagsToKey(keys, t)
}

func (t Tags) Subset(keys []string) Tags {
	subset := make(Tags, len(keys))
	for _, k := range keys {
		subset[k] = t[k]
	}
	return subset
}

func TagsToKey(order []string, t Tags) TagsKey {
	var buf bytes.Buffer
	for i, k := range order {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteString(t[k])
	}
	return TagsKey(buf.String())
}

type blockMetadata struct {
	tags   Tags
	bounds Bounds
}

func (m blockMetadata) Tags() Tags {
	return m.tags
}
func (m blockMetadata) Bounds() Bounds {
	return m.bounds
}

type colListBlockBuilder struct {
	blk *colListBlock
	key BlockKey
}

func NewColListBlockBuilder() BlockBuilder {
	return &colListBlockBuilder{
		blk: new(colListBlock),
	}
}

func (b colListBlockBuilder) SetBounds(bounds Bounds) {
	b.blk.bounds = bounds
}
func (b colListBlockBuilder) Bounds() Bounds {
	return b.blk.bounds
}

func (b colListBlockBuilder) Tags() Tags {
	return b.blk.tags
}
func (b colListBlockBuilder) NRows() int {
	return b.blk.nrows
}
func (b colListBlockBuilder) NCols() int {
	return len(b.blk.cols)
}
func (b colListBlockBuilder) Cols() []ColMeta {
	return b.blk.colMeta
}

func (b colListBlockBuilder) AddCol(c ColMeta) int {
	var col column
	switch c.Type {
	case TFloat:
		col = &floatColumn{
			ColMeta: c,
		}
	case TString:
		if c.IsCommon {
			col = &commonStrColumn{
				ColMeta: c,
			}
		} else {
			col = &stringColumn{
				ColMeta: c,
			}
		}
	case TTime:
		col = &timeColumn{
			ColMeta: c,
		}
	}
	b.blk.colMeta = append(b.blk.colMeta, c)
	b.blk.cols = append(b.blk.cols, col)
	return len(b.blk.cols) - 1
}

func (b colListBlockBuilder) SetFloat(i int, j int, value float64) {
	b.checkColType(j, TFloat)
	b.blk.cols[j].(*floatColumn).data[i] = value
}
func (b colListBlockBuilder) AppendFloat(j int, value float64) {
	b.checkColType(j, TFloat)
	col := b.blk.cols[j].(*floatColumn)
	col.data = append(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b colListBlockBuilder) AppendFloats(j int, values []float64) {
	b.checkColType(j, TFloat)
	col := b.blk.cols[j].(*floatColumn)
	col.data = append(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b colListBlockBuilder) SetString(i int, j int, value string) {
	b.checkColType(j, TString)
	b.blk.cols[j].(*stringColumn).data[i] = value
}
func (b colListBlockBuilder) AppendString(j int, value string) {
	meta := b.blk.cols[j].Meta()
	checkColType(meta, TString)
	if meta.IsCommon {
		v := b.blk.cols[j].(*commonStrColumn).value
		if value != v {
			panic(fmt.Errorf("attempting to append a different value to the column %s, which has all common values", meta.Label))
		}
		return
	}
	col := b.blk.cols[j].(*stringColumn)
	col.data = append(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b colListBlockBuilder) AppendStrings(j int, values []string) {
	b.checkColType(j, TString)
	col := b.blk.cols[j].(*stringColumn)
	col.data = append(col.data, values...)
	b.blk.nrows = len(col.data)
}
func (b colListBlockBuilder) SetCommonString(j int, value string) {
	meta := b.blk.cols[j].Meta()
	checkColType(meta, TString)
	if !meta.IsCommon {
		panic(fmt.Errorf("cannot set common value for column %s, column is not marked as common", meta.Label))
	}
	b.blk.cols[j].(*commonStrColumn).value = value
	if meta.IsTag {
		if b.blk.tags == nil {
			b.blk.tags = make(Tags)
		}
		b.blk.tags[meta.Label] = value
	}
}

func (b colListBlockBuilder) SetTime(i int, j int, value Time) {
	b.checkColType(j, TTime)
	b.blk.cols[j].(*timeColumn).data[i] = value
}
func (b colListBlockBuilder) AppendTime(j int, value Time) {
	b.checkColType(j, TTime)
	col := b.blk.cols[j].(*timeColumn)
	col.data = append(col.data, value)
	b.blk.nrows = len(col.data)
}
func (b colListBlockBuilder) AppendTimes(j int, values []Time) {
	b.checkColType(j, TTime)
	col := b.blk.cols[j].(*timeColumn)
	col.data = append(col.data, values...)
	b.blk.nrows = len(col.data)
}

func (b colListBlockBuilder) checkColType(j int, typ DataType) {
	checkColType(b.blk.cols[j].Meta(), typ)
}

func checkColType(col ColMeta, typ DataType) {
	if col.Type != typ {
		panic(fmt.Errorf("column %s is not of type %v", col.Label, typ))
	}
}

func (b colListBlockBuilder) Block() Block {
	// Create copy in mutable state
	blk := b.blk.Copy()
	return blk
}

func (b colListBlockBuilder) ClearData() {
	for _, c := range b.blk.cols {
		c.Clear()
	}
	b.blk.nrows = 0
}

func (b colListBlockBuilder) Sort(cols []string, desc bool) {
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

// Block implements Block using list of columns.
type colListBlock struct {
	bounds Bounds
	tags   Tags

	colMeta []ColMeta
	cols    []column
	nrows   int
}

func (b *colListBlock) Bounds() Bounds {
	return b.bounds
}

func (b *colListBlock) Tags() Tags {
	return b.tags
}

func (b *colListBlock) Cols() []ColMeta {
	return b.colMeta
}

func (b *colListBlock) Col(c int) ValueIterator {
	return colListValueIterator{col: c, cols: b.cols, nrows: b.nrows}
}

func (b *colListBlock) Values() ValueIterator {
	j := ValueIdx(b.colMeta)
	if j >= 0 {
		return colListValueIterator{col: j, cols: b.cols, nrows: b.nrows}
	}
	return nil
}

func (b *colListBlock) Times() ValueIterator {
	j := TimeIdx(b.colMeta)
	if j >= 0 {
		return colListValueIterator{col: j, cols: b.cols, nrows: b.nrows}
	}
	return nil
}

func (b *colListBlock) Copy() *colListBlock {
	cpy := new(colListBlock)
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
	col   int
	cols  []column
	nrows int
}

func (itr colListValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(itr.cols[itr.col].Meta(), TFloat)
	f(itr.cols[itr.col].(*floatColumn).data, itr)
}
func (itr colListValueIterator) DoString(f func([]string, RowReader)) {
	meta := itr.cols[itr.col].Meta()
	checkColType(meta, TString)
	if meta.IsTag && meta.IsCommon {
		value := itr.cols[itr.col].(*commonStrColumn).value
		strs := make([]string, itr.nrows)
		for i := range strs {
			strs[i] = value
		}
		f(strs, itr)
		return
	}
	f(itr.cols[itr.col].(*stringColumn).data, itr)
}
func (itr colListValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(itr.cols[itr.col].Meta(), TTime)
	f(itr.cols[itr.col].(*timeColumn).data, itr)
}
func (itr colListValueIterator) AtFloat(i, j int) float64 {
	checkColType(itr.cols[j].Meta(), TFloat)
	return itr.cols[j].(*floatColumn).data[i]
}
func (itr colListValueIterator) AtString(i, j int) string {
	meta := itr.cols[j].Meta()
	checkColType(meta, TString)
	if meta.IsTag && meta.IsCommon {
		return itr.cols[j].(*commonStrColumn).value
	}
	return itr.cols[j].(*stringColumn).data[i]
}
func (itr colListValueIterator) AtTime(i, j int) Time {
	checkColType(itr.cols[j].Meta(), TTime)
	return itr.cols[j].(*timeColumn).data[i]
}

type colListBlockSorter struct {
	cols []int
	desc bool
	b    *colListBlock
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

type floatColumn struct {
	ColMeta
	data []float64
}

func (c *floatColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *floatColumn) Clear() {
	c.data = c.data[0:0]
}
func (c *floatColumn) Copy() column {
	cpy := &floatColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]float64, len(c.data))
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
	data []string
}

func (c *stringColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *stringColumn) Clear() {
	c.data = c.data[0:0]
}
func (c *stringColumn) Copy() column {
	cpy := &stringColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]string, len(c.data))
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
	data []Time
}

func (c *timeColumn) Meta() ColMeta {
	return c.ColMeta
}

func (c *timeColumn) Clear() {
	c.data = c.data[0:0]
}
func (c *timeColumn) Copy() column {
	cpy := &timeColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]Time, len(c.data))
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

type BlockBuilderCache interface {
	// BlockBuilder returns an existing or new BlockBuilder for the given meta data.
	// The boolean return value indicates if BlockBuilder is new.
	BlockBuilder(meta BlockMetadata) (BlockBuilder, bool)
	ForEachBuilder(f func(BlockKey, BlockBuilder))
}

type blockBuilderCache struct {
	blocks map[BlockKey]blockState

	triggerSpec query.TriggerSpec
}

func NewBlockBuilderCache() *blockBuilderCache {
	return &blockBuilderCache{
		blocks: make(map[BlockKey]blockState),
	}
}

type blockState struct {
	builder BlockBuilder
	trigger Trigger
}

func (d *blockBuilderCache) SetTriggerSpec(ts query.TriggerSpec) {
	d.triggerSpec = ts
}

func (d *blockBuilderCache) Block(key BlockKey) Block {
	return d.blocks[key].builder.Block()
}
func (d *blockBuilderCache) BlockMetadata(key BlockKey) BlockMetadata {
	return d.blocks[key].builder
}

// BlockBuilder will return the builder for the specified block.
// If no builder exists, one will be created.
func (d *blockBuilderCache) BlockBuilder(meta BlockMetadata) (BlockBuilder, bool) {
	key := ToBlockKey(meta)
	b, ok := d.blocks[key]
	if !ok {
		builder := NewColListBlockBuilder()
		builder.SetBounds(meta.Bounds())
		t := NewTriggerFromSpec(d.triggerSpec)
		b = blockState{
			builder: builder,
			trigger: t,
		}
		d.blocks[key] = b
	}
	return b.builder, !ok
}

func (d *blockBuilderCache) ForEachBuilder(f func(BlockKey, BlockBuilder)) {
	for k, b := range d.blocks {
		f(k, b.builder)
	}
}

func (d *blockBuilderCache) DiscardBlock(key BlockKey) {
	d.blocks[key].builder.ClearData()
}
func (d *blockBuilderCache) ExpireBlock(key BlockKey) {
	delete(d.blocks, key)
}

func (d *blockBuilderCache) ForEach(f func(BlockKey)) {
	for bk := range d.blocks {
		f(bk)
	}
}

func (d *blockBuilderCache) ForEachWithContext(f func(BlockKey, Trigger, BlockContext)) {
	for bk, b := range d.blocks {
		f(bk, b.trigger, BlockContext{
			Bounds: b.builder.Bounds(),
			Count:  b.builder.NRows(),
		})
	}
}
