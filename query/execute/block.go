package execute

import (
	"bytes"
	"fmt"
	"sort"
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

	Values() ValueIterator
	Cells() CellIterator
}

// BlockBuilder builds blocks that can be used multiple times
type BlockBuilder interface {
	SetBounds(Bounds)
	SetTags(Tags)

	BlockMetadata
	NRows() int

	// AddRow increases the size of the block by one row
	AddRow(Tags)
	// AddCol increases the size of the block by one column
	AddCol(Time)

	// AddCell adds a new row for the cell and adds a new columns if necessary.
	AddCell(Cell)

	// Set sets the value at the specified coordinates
	// The rows and columns must exist before calling set, otherwise Set panics.
	Set(i, j int, value float64)

	// Clear removes all rows and columns from the block
	ClearData()

	// Block returns the block that has been built.
	// Further modifications of the builder will not effect the returned block.
	Block() Block
}

type BlockIterator interface {
	Do(f func(Block))
}

type CellIterator interface {
	Do(f func([]Cell))
}

type T int

const (
	Float T = iota
	Int
	String
	Bool
)

type ValueIterator interface {
	Do(f func([]float64))
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

type rowListBlockBuilder struct {
	blk *rowListBlock
	key BlockKey
}

func newRowListBlockBuilder() BlockBuilder {
	return &rowListBlockBuilder{
		blk: new(rowListBlock),
	}
}

func (b rowListBlockBuilder) SetBounds(bounds Bounds) {
	b.blk.bounds = bounds
}
func (b rowListBlockBuilder) Bounds() Bounds {
	return b.blk.bounds
}

func (b rowListBlockBuilder) SetTags(tags Tags) {
	b.blk.tags = tags
}
func (b rowListBlockBuilder) Tags() Tags {
	return b.blk.tags
}
func (b rowListBlockBuilder) NRows() int {
	return len(b.blk.rows)
}

func (b rowListBlockBuilder) AddRow(tags Tags) {
	b.blk.rowTags = append(b.blk.rowTags, tags)
	b.blk.rows = append(b.blk.rows, make([]float64, len(b.blk.colTimes)))
}

func (b rowListBlockBuilder) AddCol(t Time) {
	//TODO(nathanielc): ensure bounds are updated when times are added
	b.blk.colTimes = append(b.blk.colTimes, t)
	for i, r := range b.blk.rows {
		b.blk.rows[i] = append(r, 0)
	}
}

func (b rowListBlockBuilder) AddCell(c Cell) {
	// TODO(nathanielc): ensure columns are sorted in increasing time order
	b.AddRow(c.Tags)
	for i, t := range b.blk.colTimes {
		if t == c.Time {
			b.Set(len(b.blk.rows)-1, i, c.Value)
			return
		}
	}
	b.AddCol(c.Time)
	b.Set(len(b.blk.rows)-1, len(b.blk.colTimes)-1, c.Value)
}

func (b rowListBlockBuilder) Set(i int, j int, value float64) {
	b.blk.rows[i][j] = value
}

func (b rowListBlockBuilder) Block() Block {
	return b.blk.asDenseBlock()
}

func (b rowListBlockBuilder) ClearData() {
	b.blk.rows = nil
	b.blk.rowTags = nil
	b.blk.colTimes = nil
}

// Block implements Block using list of rows.
type rowListBlock struct {
	bounds Bounds
	tags   Tags

	rowTags  []Tags
	colTimes []Time

	rows [][]float64
}

func (b *rowListBlock) Bounds() Bounds {
	return b.bounds
}

func (b *rowListBlock) Tags() Tags {
	return b.tags
}

func (b *rowListBlock) Values() ValueIterator {
	return &rowListValueIterator{blk: b}
}

func (b *rowListBlock) Cells() CellIterator {
	return &rowListCellIterator{blk: b}
}

func (b *rowListBlock) asDenseBlock() *denseBlock {
	db := &denseBlock{
		bounds:   b.bounds,
		tags:     b.tags.Copy(),
		data:     make([]float64, len(b.rows)*len(b.colTimes)),
		rowTags:  make([]Tags, len(b.rowTags)),
		colTimes: make([]Time, len(b.colTimes)),
	}

	copy(db.rowTags, b.rowTags)
	copy(db.colTimes, b.colTimes)
	stride := len(db.colTimes)
	for i, r := range b.rows {
		copy(db.data[i*stride:(i+1)*stride], r)
	}
	return db
}

type rowListValueIterator struct {
	blk *rowListBlock

	currentRow int
}

func (vi *rowListValueIterator) Do(f func([]float64)) {
	for _, row := range vi.blk.rows {
		f(row)
	}
}

type rowListCellIterator struct {
	blk *rowListBlock
}

func (ci *rowListCellIterator) Do(f func([]Cell)) {
	for r, row := range ci.blk.rows {
		for c, value := range row {
			f([]Cell{{
				Value: value,
				Time:  ci.blk.colTimes[c],
				Tags:  ci.blk.rowTags[r],
			}})
		}
	}
}

type denseBlock struct {
	bounds Bounds
	tags   Tags

	data []float64

	rowTags  []Tags
	colTimes []Time
}

func (b *denseBlock) Bounds() Bounds {
	return b.bounds
}

func (b *denseBlock) Tags() Tags {
	return b.tags
}

func (b *denseBlock) Values() ValueIterator {
	return &denseBlockValueIterator{blk: b}
}

func (b *denseBlock) Cells() CellIterator {
	return &denseBlockCellIterator{blk: b}
}

func (b *denseBlock) at(r, c int) float64 {
	return b.data[r*len(b.colTimes)+c]
}

type denseBlockValueIterator struct {
	blk *denseBlock
}

func (vi *denseBlockValueIterator) Do(f func([]float64)) {
	f(vi.blk.data)
}

type denseBlockCellIterator struct {
	blk *denseBlock
}

func (ci *denseBlockCellIterator) Do(f func([]Cell)) {
	stride := len(ci.blk.colTimes)
	for i, value := range ci.blk.data {
		r := i / stride
		c := i % stride
		f([]Cell{{
			Value: value,
			Time:  ci.blk.colTimes[c],
			Tags:  ci.blk.rowTags[r],
		}})
	}
}
