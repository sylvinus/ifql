package execute

import (
	"log"
	"math"
	"sort"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

type mergeJoinTransformation struct {
	parents []DatasetID

	d     Dataset
	cache *mergeJoinCache

	leftID  DatasetID
	rightID DatasetID

	parentState map[DatasetID]*mergeJoinParentState

	keys []string
}

func newMergeJoinTransformation(d Dataset, cache *mergeJoinCache, spec *plan.MergeJoinProcedureSpec) *mergeJoinTransformation {
	return &mergeJoinTransformation{
		d:           d,
		cache:       cache,
		parentState: make(map[DatasetID]*mergeJoinParentState),
	}
}

type mergeJoinParentState struct {
	mark       Time
	processing Time
	finished   bool
}

func (t *mergeJoinTransformation) RetractBlock(id DatasetID, meta BlockMetadata) {
	bm := blockMetadata{
		tags:   meta.Tags().Subset(t.keys),
		bounds: meta.Bounds(),
	}
	key := ToBlockKey(bm)
	t.cache.DiscardBlock(key)
	t.d.RetractBlock(key)
}

func (t *mergeJoinTransformation) Process(id DatasetID, b Block) {
	log.Println("Process", id)

	bm := blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	}
	tables := t.cache.Tables(bm)

	var table *mergeTable
	switch id {
	case t.leftID:
		table = tables.left
	case t.rightID:
		table = tables.right
	}

	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, c := range cs {
			table.Insert(c.Value, c.Tags.Subset(t.keys), c.Time)
		}
	})
}

func (t *mergeJoinTransformation) UpdateWatermark(id DatasetID, mark Time) {
	t.parentState[id].mark = mark

	min := Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.mark < min {
			min = state.mark
		}
	}

	//t.d.UpdateWatermark(min)
}

func (t *mergeJoinTransformation) UpdateProcessingTime(id DatasetID, pt Time) {
	t.parentState[id].processing = pt

	min := Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.processing < min {
			min = state.processing
		}
	}

	//t.d.UpdateProcessingTime(min)
}

func (t *mergeJoinTransformation) Finish(id DatasetID) {

	t.parentState[id].finished = true
	finished := true
	for _, state := range t.parentState {
		finished = finished && state.finished
	}

	if finished {
		t.d.Finish()
	}
}

func (t *mergeJoinTransformation) setParents(ids []DatasetID) {
	if len(ids) != 2 {
		panic("joins should only ever have two parents")
	}
	t.leftID = ids[0]
	t.rightID = ids[1]

	for _, id := range ids {
		t.parentState[id] = new(mergeJoinParentState)
	}
}

type joinCell struct {
	Key   joinKey
	Tags  Tags
	Value float64
}
type joinKey struct {
	Time    Time
	TagsKey TagsKey
}

func (k joinKey) Less(o joinKey) bool {
	if k.Time < o.Time {
		return true
	} else if k.Time == o.Time {
		return k.TagsKey < o.TagsKey
	}
	return false
}

type mergeJoinCache struct {
	data map[BlockKey]*joinTables

	triggerSpec query.TriggerSpec
}

func newMergeJoinCache() *mergeJoinCache {
	return &mergeJoinCache{
		data: make(map[BlockKey]*joinTables),
	}
}

func (c *mergeJoinCache) BlockMetadata(key BlockKey) BlockMetadata {
	return c.data[key]
}

func (c *mergeJoinCache) Block(key BlockKey) Block {
	return c.data[key].Join()
}

func (c *mergeJoinCache) ForEach(f func(BlockKey)) {
	for bk := range c.data {
		f(bk)
	}
}

func (c *mergeJoinCache) ForEachWithContext(f func(BlockKey, Trigger, BlockContext)) {
	for bk, tables := range c.data {
		bc := BlockContext{
			Bounds: tables.bounds,
			Count:  tables.Size(),
		}
		f(bk, tables.trigger, bc)
	}
}

func (c *mergeJoinCache) DiscardBlock(key BlockKey) {
	c.data[key].ClearData()
}

func (c *mergeJoinCache) ExpireBlock(key BlockKey) {
	delete(c.data, key)
}

func (c *mergeJoinCache) setTriggerSpec(spec query.TriggerSpec) {
	c.triggerSpec = spec
}

func (c *mergeJoinCache) Tables(bm BlockMetadata) *joinTables {
	key := ToBlockKey(bm)
	tables := c.data[key]
	if tables == nil {
		tables = &joinTables{
			tags:    bm.Tags(),
			bounds:  bm.Bounds(),
			left:    new(mergeTable),
			right:   new(mergeTable),
			trigger: newTriggerFromSpec(c.triggerSpec),
		}
		c.data[key] = tables
	}
	return tables
}

type joinTables struct {
	tags   Tags
	bounds Bounds

	left  *mergeTable
	right *mergeTable

	trigger Trigger
}

func (t *joinTables) Bounds() Bounds {
	return t.bounds
}
func (t *joinTables) Tags() Tags {
	return t.tags
}
func (t *joinTables) Size() int {
	return len(t.left.cells) + len(t.right.cells)
}

func (t *joinTables) ClearData() {
	t.left = new(mergeTable)
	t.right = new(mergeTable)
}

func (t *joinTables) Join() Block {
	// Perform sort-merge join

	builder := newRowListBlockBuilder()
	builder.SetBounds(t.bounds)
	builder.SetTags(t.tags)

	var left, leftSet, right, rightSet []joinCell
	var leftKey, rightKey joinKey
	left = t.left.Sorted()
	right = t.right.Sorted()

	left, leftSet, leftKey = t.advance(left)
	right, rightSet, rightKey = t.advance(right)
	for len(leftSet) > 0 && len(rightSet) > 0 {
		if leftKey == rightKey {
			// Inner join
			for _, l := range leftSet {
				for _, r := range rightSet {
					v := t.eval(l.Value, r.Value)
					builder.AddCell(Cell{
						Time:  l.Key.Time,
						Tags:  l.Tags,
						Value: v,
					})
				}
			}

			left, leftSet, leftKey = t.advance(left)
			right, rightSet, rightKey = t.advance(right)
		} else if leftKey.Less(rightKey) {
			left, leftSet, leftKey = t.advance(left)
		} else {
			right, rightSet, rightKey = t.advance(right)
		}
	}
	return builder.Block()
}

func (t *joinTables) advance(table []joinCell) ([]joinCell, []joinCell, joinKey) {
	if len(table) == 0 {
		return nil, nil, joinKey{}
	}
	key := table[0].Key
	var subset []joinCell
	for len(table) > 0 && table[0].Key == key {
		subset = append(subset, table[0])
		table = table[1:]
	}
	return table, subset, key
}

func (t *joinTables) eval(l, r float64) float64 {
	// TODO perform specified expression
	return l / r
}

type mergeTable struct {
	cells []joinCell
}

func (t *mergeTable) Insert(value float64, tags Tags, time Time) {
	cell := joinCell{
		Key: joinKey{
			Time:    time,
			TagsKey: tags.Key(),
		},
		Tags:  tags,
		Value: value,
	}
	t.cells = append(t.cells, cell)
}

func (t *mergeTable) Sorted() []joinCell {
	sort.Sort(cells(t.cells))
	return t.cells
}

type cells []joinCell

func (c cells) Len() int               { return len(c) }
func (c cells) Less(i int, j int) bool { return c[i].Key.Less(c[j].Key) }
func (c cells) Swap(i int, j int)      { c[i], c[j] = c[j], c[i] }
