package execute

import (
	"bytes"
	"fmt"
	"math"
	"runtime/debug"
	"sort"
	"strconv"

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
	Col(c int) ValueIterator

	Values() ValueIterator
	Times() ValueIterator

	AtFloat(i, j int) float64
	AtString(i, j int) string
	AtTime(i, j int) Time
}

func ValueIdx(b Block) int {
	for j, c := range b.Cols() {
		if c.Label == valueColLabel {
			return j
		}
	}
	return -1
}
func TimeIdx(b Block) int {
	for j, c := range b.Cols() {
		if c.Label == timeColLabel {
			return j
		}
	}
	return -1
}

// BlockBuilder builds blocks that can be used multiple times
type BlockBuilder interface {
	SetBounds(Bounds)

	//SetTags sets tags that are common to all records of this block
	SetTags(Tags)

	BlockMetadata

	NRows() int
	NCols() int

	// AddCol increases the size of the block by one column
	// Columns need not be added for tags that are common to the block
	AddCol(ColMeta)

	// Set sets the value at the specified coordinates
	// The rows and columns must exist before calling set, otherwise Set panics.
	SetFloat(i, j int, value float64)
	SetString(i, j int, value string)
	SetTime(i, j int, value Time)

	AppendFloat(j int, value float64)
	AppendString(j int, value string)
	AppendTime(j int, value Time)

	AppendFloats(j int, values []float64)
	AppendStrings(j int, values []string)
	AppendTimes(j int, values []Time)

	// Clear removes all rows and columns from the block
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
	default:
		return "unknown"
	}
}

type ColMeta struct {
	Label string
	Type  DataType
	IsTag bool
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
	Do(f func(Block))
}

type ValueIterator interface {
	DoFloat(f func([]float64))
	DoString(f func([]string))
	DoTime(f func([]Time))
}

func TagsForRow(b Block, i int) Tags {
	cols := b.Cols()
	tags := make(Tags, len(cols)-2)
	for j, c := range cols {
		if c.IsTag {
			tags[c.Label] = b.AtString(i, j)
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

func (b colListBlockBuilder) SetTags(tags Tags) {
	b.blk.tags = tags
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

func (b colListBlockBuilder) AddCol(c ColMeta) {
	if len(b.blk.cols) > 4 {
		panic("asdf")
	}
	var col column
	switch c.Type {
	case TFloat:
		col = &floatColumn{
			ColMeta: c,
		}
	case TString:
		col = &stringColumn{
			ColMeta: c,
		}
	case TTime:
		col = &timeColumn{
			ColMeta: c,
		}
	}
	b.blk.cols = append(b.blk.cols, col)
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
	b.checkColType(j, TString)
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

	// Add tagColums
	keys := blk.tags.Keys()
	for _, k := range keys {
		blk.cols = append(blk.cols, &tagColumn{
			ColMeta: ColMeta{
				Label: k,
				Type:  TString,
				IsTag: true,
			},
			value: blk.tags[k],
			size:  b.blk.nrows,
		})
	}

	// Build meta list
	blk.colMeta = make([]ColMeta, len(blk.cols))
	for i, c := range blk.cols {
		blk.colMeta[i] = c.Meta()
	}
	return blk
}

func (b colListBlockBuilder) ClearData() {
	for _, c := range b.blk.cols {
		c.Clear()
	}
	b.blk.nrows = 0
}

// Block implements Block using list of rows.
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
	meta := b.colMeta[c]
	col := b.cols[c]
	if meta.IsTag {
		return &tagColValueIterator{col: col.(*tagColumn)}
	}
	return colListValueIterator{col: col}
}

func (b *colListBlock) Values() ValueIterator {
	j := ValueIdx(b)
	if j >= 0 {
		return colListValueIterator{col: b.cols[j]}
	}
	return nil
}

func (b *colListBlock) Times() ValueIterator {
	j := TimeIdx(b)
	if j >= 0 {
		return colListValueIterator{col: b.cols[j]}
	}
	return nil
}

func (b *colListBlock) AtFloat(i, j int) float64 {
	checkColType(b.colMeta[j], TFloat)
	return b.cols[j].(*floatColumn).data[i]
}
func (b *colListBlock) AtString(i, j int) string {
	if b.colMeta[j].IsTag {
		return b.cols[j].(*tagColumn).value
	}
	checkColType(b.colMeta[j], TString)
	return b.cols[j].(*stringColumn).data[i]
}
func (b *colListBlock) AtTime(i, j int) Time {
	checkColType(b.colMeta[j], TTime)
	return b.cols[j].(*timeColumn).data[i]
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
	col column
}

func (itr colListValueIterator) DoFloat(f func([]float64)) {
	if itr.col.Meta().Type != TFloat {
		panic("column is not of type float")
	}
	f(itr.col.(*floatColumn).data)
}
func (itr colListValueIterator) DoString(f func([]string)) {
	if itr.col.Meta().Type != TString {
		panic("column is not of type string")
	}
	f(itr.col.(*stringColumn).data)
}
func (itr colListValueIterator) DoTime(f func([]Time)) {
	if itr.col.Meta().Type != TTime {
		panic("column is not of type time")
	}
	f(itr.col.(*timeColumn).data)
}

type column interface {
	Meta() ColMeta
	Clear()
	Len() int
	Copy() column
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
func (c *floatColumn) Len() int {
	return len(c.data)
}
func (c *floatColumn) Copy() column {
	cpy := &floatColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]float64, len(c.data))
	copy(cpy.data, c.data)
	return cpy
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
func (c *stringColumn) Len() int {
	return len(c.data)
}
func (c *stringColumn) Copy() column {
	cpy := &stringColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]string, len(c.data))
	copy(cpy.data, c.data)
	return cpy
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
func (c *timeColumn) Len() int {
	return len(c.data)
}
func (c *timeColumn) Copy() column {
	cpy := &timeColumn{
		ColMeta: c.ColMeta,
	}
	cpy.data = make([]Time, len(c.data))
	copy(cpy.data, c.data)
	return cpy
}

//tagColumn has the same value for all rows
type tagColumn struct {
	ColMeta
	value string
	size  int
}

func (c *tagColumn) Meta() ColMeta {
	return c.ColMeta
}
func (c *tagColumn) Clear() {
	c.size = 0
}
func (c *tagColumn) Len() int {
	return c.size
}
func (c *tagColumn) Copy() column {
	cpy := new(tagColumn)
	*cpy = *c
	return cpy
}

type tagColValueIterator struct {
	col    *tagColumn
	values []string
}

func (*tagColValueIterator) DoFloat(f func([]float64)) {}
func (itr *tagColValueIterator) DoString(f func([]string)) {
	strs := make([]string, itr.col.size)
	for i := range strs {
		strs[i] = itr.col.value
	}
	f(strs)
}

func (*tagColValueIterator) DoTime(f func([]Time)) {}

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
		builder.SetTags(meta.Tags())
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

type FormatOption func(*formatter)

func Formatted(b Block, opts ...FormatOption) fmt.Formatter {
	f := formatter{
		b: b,
	}
	for _, o := range opts {
		o(&f)
	}
	return f
}

func Head(m int) FormatOption {
	return func(f *formatter) { f.head = m }
}
func Squeeze() FormatOption {
	return func(f *formatter) { f.squeeze = true }
}

type formatter struct {
	b       Block
	head    int
	squeeze bool
}

func (f formatter) Format(fs fmt.State, c rune) {
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "%#v", f.b)
		return
	}
	defer func() {
		r := recover()
		if r != nil {
			panic(fmt.Sprintf("%v\n%s", r, debug.Stack()))
		}
	}()
	f.format(fs, c)
}

func (f formatter) format(fs fmt.State, c rune) {
	tags := f.b.Tags()
	keys := tags.Keys()
	cols := f.b.Cols()
	fmt.Fprintf(fs, "Block: keys: %v bounds: %v\n", keys, f.b.Bounds())
	nCols := len(cols) + len(keys)

	// Determine number of rows to print
	nrows := math.MaxInt64
	if f.head > 0 {
		nrows = f.head
	}

	// Determine precision of floating point values
	prec, pOk := fs.Precision()
	if !pOk {
		prec = -1
	}

	var widths widther
	if f.squeeze {
		widths = make(columnWidth, nCols)
	} else {
		widths = new(uniformWidth)
	}

	fmtC := byte(c)
	if fmtC == 'v' {
		fmtC = 'g'
	}
	floatBuf := make([]byte, 0, 64)
	maxWidth := computeWidths(f.b, fmtC, nrows, prec, widths, floatBuf)

	width, _ := fs.Width()
	if width < maxWidth {
		width = maxWidth
	}
	if width < 2 {
		width = 2
	}
	pad := make([]byte, width)
	for i := range pad {
		pad[i] = ' '
	}
	dash := make([]byte, width)
	for i := range dash {
		dash[i] = '-'
	}
	eol := []byte{'\n'}

	// Print column headers
	for j, c := range cols {
		buf := []byte(c.Label)
		// Check justification
		if fs.Flag('-') {
			fs.Write(buf)
			fs.Write(pad[:widths.width(j)-len(buf)])
		} else {
			fs.Write(pad[:widths.width(j)-len(buf)])
			fs.Write(buf)
		}
		fs.Write(pad[:2])
	}
	fs.Write(eol)
	// Print header separator
	for j := range cols {
		fs.Write(dash[:widths.width(j)])
		fs.Write(pad[:2])
	}
	fs.Write(eol)

	n := nrows
	times := f.b.Times()
	i := 0
	times.DoTime(func(ts []Time) {
		l := len(ts)
		if n < l {
			l = n
			n = 0
		} else {
			n -= l
		}
		for range ts[:l] {
			for j, c := range cols {
				var buf []byte
				switch c.Type {
				case TFloat:
					buf = strconv.AppendFloat(floatBuf, f.b.AtFloat(i, j), fmtC, prec, 64)
				case TTime:
					buf = []byte(f.b.AtTime(i, j).String())
				case TString:
					buf = []byte(f.b.AtString(i, j))
				}
				// Check justification
				if fs.Flag('-') {
					fs.Write(buf)
					fs.Write(pad[:widths.width(j)-len(buf)])
				} else {
					fs.Write(pad[:widths.width(j)-len(buf)])
					fs.Write(buf)
				}
				fs.Write(pad[:2])
			}
			fs.Write(eol)
			i++
		}
	})
}

func computeWidths(b Block, fmtC byte, rows, prec int, widths widther, buf []byte) int {
	maxWidth := 0
	for j, c := range b.Cols() {
		n := rows
		values := b.Col(j)
		width := len(c.Label)
		switch c.Type {
		case TFloat:
			values.DoFloat(func(vs []float64) {
				l := len(vs)
				if n < l {
					l = n
					n = 0
				} else {
					n -= l
				}
				for _, v := range vs[:l] {
					buf = strconv.AppendFloat(buf[0:0], v, fmtC, prec, 64)
					if w := len(buf); w > width {
						width = w
					}
				}
			})
		case TString:
			values.DoString(func(vs []string) {
				l := len(vs)
				if n < l {
					l = n
					n = 0
				} else {
					n -= l
				}
				for _, v := range vs[:l] {
					if w := len(v); w > width {
						width = w
					}
				}
			})
		case TTime:
			width = len(fixedWidthTimeFmt)
		}
		widths.setWidth(j, width)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

type widther interface {
	width(i int) int
	setWidth(i, w int)
}

type uniformWidth int

func (u *uniformWidth) width(_ int) int { return int(*u) }
func (u *uniformWidth) setWidth(_, w int) {
	if uniformWidth(w) > *u {
		*u = uniformWidth(w)
	}
}

type columnWidth []int

func (c columnWidth) width(i int) int   { return c[i] }
func (c columnWidth) setWidth(i, w int) { c[i] = w }
