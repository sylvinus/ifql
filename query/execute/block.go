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

	// Times returns an iterator to consume the values for the "_time" column.
	Times() ValueIterator
	// Values returns an iterator to consume the values for the "_value" column.
	// If no column exists and error is returned
	Values() (ValueIterator, error)

	// Retain increases the reference count of the Block by 1.
	// Retain may be called simultaneously from multiple goroutines.
	Retain()

	// Release decreases the reference count of the Block by 1.
	// When the reference count goes to zero, the block is freed.
	// Release may be called simultaneously from multiple goroutines.
	Release()
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
func CacheOneTimeBlock(b Block, a Allocator) Block {
	_, ok := b.(OneTimeBlock)
	if !ok {
		return b
	}
	return CopyBlock(b, a)
}

// CopyBlock returns a copy of the block and is OneTimeBlock safe.
func CopyBlock(b Block, a Allocator) Block {
	builder := NewColListBlockBuilder(&ColListAllocator{a})
	builder.SetBounds(b.Bounds())

	cols := b.Cols()
	colMap := make([]int, len(cols))
	for j, c := range cols {
		colMap[j] = j
		builder.AddCol(c)
		if c.IsTag() && c.Common {
			builder.SetCommonString(j, b.Tags()[c.Label])
		}
	}

	AppendBlock(b, builder, colMap)
	// ColListBlockBuilders do not error
	nb, _ := builder.Block()
	return nb
}

// AddBlockCols adds the columns of b onto builder.
func AddBlockCols(b Block, builder BlockBuilder) {
	cols := b.Cols()
	for j, c := range cols {
		builder.AddCol(c)
		if c.IsTag() && c.Common {
			builder.SetCommonString(j, b.Tags()[c.Label])
		}
	}
}

// AddNewCols adds the columns of b onto builder that did not already exist.
// Returns the mapping of builder cols to block cols.
func AddNewCols(b Block, builder BlockBuilder) []int {
	cols := b.Cols()
	existing := builder.Cols()
	colMap := make([]int, len(existing))
	for j, c := range cols {
		found := false
		for ej, ec := range existing {
			if c.Label == ec.Label {
				colMap[ej] = j
				found = true
				break
			}
		}
		if !found {
			builder.AddCol(c)
			colMap = append(colMap, j)

			if c.IsTag() && c.Common {
				builder.SetCommonString(j, b.Tags()[c.Label])
			}
		}
	}
	return colMap
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
			if j == timeIdx || c.Common {
				continue
			}
			for i := range ts {
				switch c.Type {
				case TBool:
					builder.AppendBool(j, rr.AtBool(i, colMap[j]))
				case TInt:
					builder.AppendInt(j, rr.AtInt(i, colMap[j]))
				case TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, colMap[j]))
				case TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, colMap[j]))
				case TString:
					builder.AppendString(j, rr.AtString(i, colMap[j]))
				case TTime:
					builder.AppendTime(j, rr.AtTime(i, colMap[j]))
				default:
					PanicUnknownType(c.Type)
				}
			}
		}
	})
}

// AppendRow appends a single row from rr onto builder.
// The colMap is a map of builder columnm index to rr column index.
func AppendRow(i int, rr RowReader, builder BlockBuilder, colMap []int) {
	for j, c := range builder.Cols() {
		switch c.Type {
		case TBool:
			builder.AppendBool(j, rr.AtBool(i, colMap[j]))
		case TInt:
			builder.AppendInt(j, rr.AtInt(i, colMap[j]))
		case TUInt:
			builder.AppendUInt(j, rr.AtUInt(i, colMap[j]))
		case TFloat:
			builder.AppendFloat(j, rr.AtFloat(i, colMap[j]))
		case TString:
			builder.AppendString(j, rr.AtString(i, colMap[j]))
		case TTime:
			builder.AppendTime(j, rr.AtTime(i, colMap[j]))
		default:
			PanicUnknownType(c.Type)
		}
	}
}

// AppendRowForCols appends a single row from rr onto builder for the specified cols.
// The colMap is a map of builder columnm index to rr column index.
func AppendRowForCols(i int, rr RowReader, builder BlockBuilder, cols []ColMeta, colMap []int) {
	for j, c := range cols {
		switch c.Type {
		case TBool:
			builder.AppendBool(j, rr.AtBool(i, colMap[j]))
		case TInt:
			builder.AppendInt(j, rr.AtInt(i, colMap[j]))
		case TUInt:
			builder.AppendUInt(j, rr.AtUInt(i, colMap[j]))
		case TFloat:
			builder.AppendFloat(j, rr.AtFloat(i, colMap[j]))
		case TString:
			builder.AppendString(j, rr.AtString(i, colMap[j]))
		case TTime:
			builder.AppendTime(j, rr.AtTime(i, colMap[j]))
		default:
			PanicUnknownType(c.Type)
		}
	}
}

// AddTags add columns to the builder for the given tags.
// It is assumed that all tags are common to all rows of this block.
func AddTags(t Tags, b BlockBuilder) {
	keys := t.Keys()
	for _, k := range keys {
		j := b.AddCol(ColMeta{
			Label:  k,
			Type:   TString,
			Kind:   TagColKind,
			Common: true,
		})
		b.SetCommonString(j, t[k])
	}
}

var NoDefaultValueColumn = fmt.Errorf("no default value column %q found.", DefaultValueColLabel)

func ValueCol(cols []ColMeta) (ColMeta, error) {
	for _, c := range cols {
		if c.Label == DefaultValueColLabel {
			return c, nil
		}
	}
	return ColMeta{}, NoDefaultValueColumn
}
func ValueIdx(cols []ColMeta) int {
	return ColIdx(DefaultValueColLabel, cols)
}
func TimeIdx(cols []ColMeta) int {
	return ColIdx(TimeColLabel, cols)
}
func ColIdx(label string, cols []ColMeta) int {
	for j, c := range cols {
		if c.Label == label {
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

	// SetCommonString sets a single value for the entire column.
	SetCommonString(j int, value string)

	AppendBool(j int, value bool)
	AppendInt(j int, value int64)
	AppendUInt(j int, value uint64)
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
	Block() (Block, error)
}

type DataType int

const (
	TInvalid DataType = iota
	TBool
	TInt
	TUInt
	TFloat
	TString
	TTime
)

func (t DataType) String() string {
	switch t {
	case TInvalid:
		return "invalid"
	case TBool:
		return "bool"
	case TInt:
		return "int"
	case TUInt:
		return "uint"
	case TFloat:
		return "float"
	case TString:
		return "string"
	case TTime:
		return "time"
	default:
		return "unknown"
	}
}

type ColMeta struct {
	Label string
	Type  DataType
	Kind  ColKind
	// Common indicates that the value for the column is shared by all rows.
	Common bool
}

func (c ColMeta) IsTime() bool {
	return c.Kind == TimeColKind
}
func (c ColMeta) IsTag() bool {
	return c.Kind == TagColKind
}
func (c ColMeta) IsValue() bool {
	return c.Kind == ValueColKind
}

const (
	DefaultValueColLabel = "_value"
	TimeColLabel         = "_time"
)

type ColKind int

const (
	InvalidColKind = iota
	TimeColKind
	TagColKind
	ValueColKind
)

func (k ColKind) String() string {
	switch k {
	case InvalidColKind:
		return "invalid"
	case TimeColKind:
		return "time"
	case TagColKind:
		return "tag"
	case ValueColKind:
		return "value"
	default:
		return "unknown"
	}
}

var (
	TimeCol = ColMeta{
		Label: TimeColLabel,
		Type:  TTime,
		Kind:  TimeColKind,
	}
)

type BlockIterator interface {
	Do(f func(Block) error) error
}

type ValueIterator interface {
	DoBool(f func([]bool, RowReader))
	DoInt(f func([]int64, RowReader))
	DoUInt(f func([]uint64, RowReader))
	DoFloat(f func([]float64, RowReader))
	DoString(f func([]string, RowReader))
	DoTime(f func([]Time, RowReader))
}

type RowReader interface {
	Cols() []ColMeta
	// AtBool returns the bool value of another column and given index.
	AtBool(i, j int) bool
	// AtInt returns the int value of another column and given index.
	AtInt(i, j int) int64
	// AtUInt returns the uint value of another column and given index.
	AtUInt(i, j int) uint64
	// AtFloat returns the float value of another column and given index.
	AtFloat(i, j int) float64
	// AtString returns the string value of another column and given index.
	AtString(i, j int) string
	// AtTime returns the time value of another column and given index.
	AtTime(i, j int) Time
}

func TagsForRow(i int, rr RowReader) Tags {
	cols := rr.Cols()
	tags := make(Tags, len(cols))
	for j, c := range cols {
		if c.IsTag() {
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

// Subset creates a new Tags that is a subset of t, using the list of keys.
// If a keys is provided that does not exist on t, then a subset is not possible and
// the boolean return value is false.
func (t Tags) Subset(keys []string) (Tags, bool) {
	subset := make(Tags, len(keys))
	for _, k := range keys {
		v, ok := t[k]
		if !ok {
			return nil, false
		}
		subset[k] = v
	}
	return subset, true
}

func (t Tags) IntersectingSubset(keys []string) Tags {
	subset := make(Tags, len(keys))
	for _, k := range keys {
		v, ok := t[k]
		if ok {
			subset[k] = v
		}
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

type BlockBuilderCache interface {
	// BlockBuilder returns an existing or new BlockBuilder for the given meta data.
	// The boolean return value indicates if BlockBuilder is new.
	BlockBuilder(meta BlockMetadata) (BlockBuilder, bool)
	ForEachBuilder(f func(BlockKey, BlockBuilder))
}

type blockBuilderCache struct {
	blocks map[BlockKey]blockState
	alloc  Allocator

	triggerSpec query.TriggerSpec
}

func NewBlockBuilderCache(a Allocator) *blockBuilderCache {
	return &blockBuilderCache{
		blocks: make(map[BlockKey]blockState),
		alloc:  a,
	}
}

type blockState struct {
	builder BlockBuilder
	trigger Trigger
}

func (d *blockBuilderCache) SetTriggerSpec(ts query.TriggerSpec) {
	d.triggerSpec = ts
}

func (d *blockBuilderCache) Block(key BlockKey) (Block, error) {
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
		bb := NewColListBlockBuilder(&ColListAllocator{d.alloc})
		//bb := NewArrowBlockBuilder(d.alloc)
		bb.SetBounds(meta.Bounds())
		t := NewTriggerFromSpec(d.triggerSpec)
		b = blockState{
			builder: bb,
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
	d.blocks[key].builder.ClearData()
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
