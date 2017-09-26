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
	Values() ValueIterator
	Cells() CellIterator
}

// BlockBuilder builds blocks that can be used multiple times
type BlockBuilder interface {
	SetBounds(Bounds)

	//SetTags sets tags that are global to all records of this block
	SetTags(Tags)

	BlockMetadata

	NRows() int
	NCols() int

	// AddRow increases the size of the block by one row
	AddRow()
	// AddCol increases the size of the block by one column
	// Columns need not be added for tags that are common to the block
	AddCol(ColMeta)

	// AddCell increases the size of the block by one row and
	// sets the values based on the cell.
	AddCell(Cell)

	// Set sets the value at the specified coordinates
	// The rows and columns must exist before calling set, otherwise Set panics.
	SetFloat(i, j int, value float64)
	SetString(i, j int, value string)
	SetTime(i, j int, value Time)

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

type ColMeta struct {
	Label string
	Type  DataType
}

var (
	TimeCol = ColMeta{
		Label: "time",
		Type:  TTime,
	}
	ValueCol = ColMeta{
		Label: "value",
		Type:  TFloat,
	}
)

type Record struct {
	Values []interface{}
}

type BlockIterator interface {
	Do(f func(Block))
}

type ValueIterator interface {
	Do(f func([]float64))
}
type CellIterator interface {
	Do(f func([]Cell))
}

type Cell struct {
	Value float64
	Time  Time
	Tags  Tags
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

func (b colListBlockBuilder) AddRow() {
	for _, c := range b.blk.cols {
		c.Grow()
	}
	b.blk.nrows++
}

func (b colListBlockBuilder) AddCol(c ColMeta) {
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

func (b colListBlockBuilder) AddCell(cell Cell) {
	//TODO(nathanielc): What do we do about new tags not as columns or as tags on the block.
	for _, c := range b.blk.cols {
		switch col := c.(type) {
		case *floatColumn:
			col.data = append(col.data, cell.Value)
		case *stringColumn:
			col.data = append(col.data, cell.Tags[col.Label])
		case *timeColumn:
			col.data = append(col.data, cell.Time)
		}
	}
	b.blk.nrows++
}

func (b colListBlockBuilder) SetTime(i int, j int, value Time) {
	if b.blk.cols[j].Meta().Type != TTime {
		panic(fmt.Errorf("column %d is not of type time", j))
	}
	b.blk.cols[j].(*timeColumn).data[i] = value
}
func (b colListBlockBuilder) SetString(i int, j int, value string) {
	if b.blk.cols[j].Meta().Type != TString {
		panic(fmt.Errorf("column %d is not of type string", j))
	}
	b.blk.cols[j].(*stringColumn).data[i] = value
}
func (b colListBlockBuilder) SetFloat(i int, j int, value float64) {
	if b.blk.cols[j].Meta().Type != TFloat {
		panic(fmt.Errorf("column %d is not of type float", j))
	}
	b.blk.cols[j].(*floatColumn).data[i] = value
}

func (b colListBlockBuilder) Block() Block {
	//TODO(nathanielc): Construct Apache Arrow based block
	return b.blk.Copy()
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

	cols  []column
	nrows int
}

func (b *colListBlock) Bounds() Bounds {
	return b.bounds
}

func (b *colListBlock) Tags() Tags {
	return b.tags
}
func (b *colListBlock) Cols() []ColMeta {
	cols := make([]ColMeta, len(b.cols))
	for i, c := range b.cols {
		cols[i] = c.Meta()
	}
	return cols
}

func (b *colListBlock) Values() ValueIterator {
	return colListValueIterator{blk: b}
}
func (b *colListBlock) Cells() CellIterator {
	return colListCellIterator{blk: b}
}

func (b *colListBlock) Copy() *colListBlock {
	cpy := new(colListBlock)
	cpy.bounds = b.bounds
	cpy.tags = b.tags.Copy()
	cpy.nrows = b.nrows

	cpy.cols = make([]column, len(b.cols))
	for i, c := range b.cols {
		cpy.cols[i] = c.Copy()
	}

	return cpy
}

type colListValueIterator struct {
	blk *colListBlock
}

func (vi colListValueIterator) Do(f func([]float64)) {
	for _, c := range vi.blk.cols {
		meta := c.Meta()
		// TODO(nathanielc): Change api to deal with multiple value columns
		if meta.Label == "value" && meta.Type == TFloat {
			f(c.(*floatColumn).data)
			break
		}
	}
}

type colListCellIterator struct {
	blk *colListBlock
}

func (ci colListCellIterator) Do(f func([]Cell)) {
	cells := make([]Cell, ci.blk.nrows)
	for i := 0; i < ci.blk.nrows; i++ {
		cells[i].Tags = ci.blk.tags.Copy()
		for _, c := range ci.blk.cols {
			// TODO(nathanielc): This assumes all string cols are tags and that
			// there is only one float(value) and one time column.
			switch col := c.(type) {
			case *floatColumn:
				cells[i].Value = col.data[i]
			case *stringColumn:
				cells[i].Tags[col.Label] = col.data[i]
			case *timeColumn:
				cells[i].Time = col.data[i]
			}
		}
	}
	f(cells)
}

type column interface {
	Meta() ColMeta
	Grow()
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

func (c *floatColumn) Grow() {
	c.data = append(c.data, 0)
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

func (c *stringColumn) Grow() {
	c.data = append(c.data, "")
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

func (c *timeColumn) Grow() {
	c.data = append(c.data, 0)
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

type BlockBuilderCache interface {
	BlockBuilder(meta BlockMetadata) BlockBuilder
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
func (d *blockBuilderCache) BlockBuilder(meta BlockMetadata) BlockBuilder {
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
	return b.builder
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

type TableFmt struct {
	Block Block
}

func (t TableFmt) String() string {
	var buf bytes.Buffer
	b := t.Block
	tags := b.Tags()
	keys := tags.Keys()
	fmt.Fprintf(&buf, "Block: keys: %v bounds: %v\n", keys, b.Bounds())
	cols := b.Cols()
	size := 0
	for _, c := range cols {
		if c.Label == "time" {
			fmt.Fprintf(&buf, "%31s", c.Label)
			size += 31
		} else {
			fmt.Fprintf(&buf, "%20s", c.Label)
			size += 20
		}
	}
	for _, k := range keys {
		fmt.Fprintf(&buf, "%20s", k)
		size += 20
	}
	buf.WriteRune('\n')
	for i := 0; i < size; i++ {
		buf.WriteRune('-')
	}
	buf.WriteRune('\n')
	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, cell := range cs {
			for _, c := range cols {
				label := c.Label
				if label == "time" {
					fmt.Fprintf(&buf, "%31v", cell.Time)
				} else if label == "value" {
					fmt.Fprintf(&buf, "%20f", cell.Value)
				} else {
					fmt.Fprintf(&buf, "%20s", cell.Tags[label])
				}
			}
			for _, k := range keys {
				fmt.Fprintf(&buf, "%20s", tags[k])
			}
			buf.WriteRune('\n')
		}
	})
	return buf.String()
}
