package functions

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const JoinKind = "join"
const MergeJoinKind = "merge-join"

type JoinOpSpec struct {
	Keys       []string             `json:"keys"`
	Expression query.ExpressionSpec `json:"expression"`
}

func init() {
	query.RegisterOpSpec(JoinKind, newJoinOp)
	//TODO(nathanielc): Allow for other types of join implementations
	plan.RegisterProcedureSpec(MergeJoinKind, newMergeJoinProcedure, JoinKind)
	execute.RegisterTransformation(MergeJoinKind, createMergeJoinTransformation)
}

func newJoinOp() query.OperationSpec {
	return new(JoinOpSpec)
}

func (s *JoinOpSpec) Kind() query.OperationKind {
	return JoinKind
}

type MergeJoinProcedureSpec struct {
	Keys       []string             `json:"keys"`
	Expression query.ExpressionSpec `json:"expression"`
}

func newMergeJoinProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*JoinOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &MergeJoinProcedureSpec{
		Keys:       spec.Keys,
		Expression: spec.Expression,
	}
	sort.Strings(p.Keys)
	return p, nil
}

func (s *MergeJoinProcedureSpec) Kind() plan.ProcedureKind {
	return MergeJoinKind
}

func createMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, now time.Time) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeJoinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := newMergeJoinCache()
	d := execute.NewDataset(id, mode, cache)
	t := newMergeJoinTransformation(d, cache, s)
	return t, d, nil
}

type mergeJoinTransformation struct {
	parents []execute.DatasetID

	d     execute.Dataset
	cache MergeJoinCache

	leftID  execute.DatasetID
	rightID execute.DatasetID

	parentState map[execute.DatasetID]*mergeJoinParentState

	keys []string
}

func newMergeJoinTransformation(d execute.Dataset, cache MergeJoinCache, spec *MergeJoinProcedureSpec) *mergeJoinTransformation {
	return &mergeJoinTransformation{
		d:           d,
		cache:       cache,
		parentState: make(map[execute.DatasetID]*mergeJoinParentState),
	}
}

type mergeJoinParentState struct {
	mark       execute.Time
	processing execute.Time
	finished   bool
}

func (t *mergeJoinTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	bm := blockMetadata{
		tags:   meta.Tags().Subset(t.keys),
		bounds: meta.Bounds(),
	}
	t.d.RetractBlock(execute.ToBlockKey(bm))
}

func (t *mergeJoinTransformation) Process(id execute.DatasetID, b execute.Block) {
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
	cells.Do(func(cs []execute.Cell) {
		for _, c := range cs {
			table.Insert(c.Value, c.Tags.Subset(t.keys), c.Time)
		}
	})
}

func (t *mergeJoinTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.parentState[id].mark = mark

	min := execute.Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.mark < min {
			min = state.mark
		}
	}

	t.d.UpdateWatermark(min)
}

func (t *mergeJoinTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.parentState[id].processing = pt

	min := execute.Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.processing < min {
			min = state.processing
		}
	}

	t.d.UpdateProcessingTime(min)
}

func (t *mergeJoinTransformation) Finish(id execute.DatasetID) {

	t.parentState[id].finished = true
	finished := true
	for _, state := range t.parentState {
		finished = finished && state.finished
	}

	if finished {
		t.d.Finish()
	}
}

func (t *mergeJoinTransformation) SetParents(ids []execute.DatasetID) {
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
	Tags  execute.Tags
	Value float64
}
type joinKey struct {
	Time    execute.Time
	TagsKey execute.TagsKey
}

func (k joinKey) Less(o joinKey) bool {
	if k.Time < o.Time {
		return true
	} else if k.Time == o.Time {
		return k.TagsKey < o.TagsKey
	}
	return false
}

type MergeJoinCache interface {
	Tables(execute.BlockMetadata) *joinTables
}

type mergeJoinCache struct {
	data map[execute.BlockKey]*joinTables

	triggerSpec query.TriggerSpec
}

func newMergeJoinCache() *mergeJoinCache {
	return &mergeJoinCache{
		data: make(map[execute.BlockKey]*joinTables),
	}
}

func (c *mergeJoinCache) BlockMetadata(key execute.BlockKey) execute.BlockMetadata {
	return c.data[key]
}

func (c *mergeJoinCache) Block(key execute.BlockKey) execute.Block {
	return c.data[key].Join()
}

func (c *mergeJoinCache) ForEach(f func(execute.BlockKey)) {
	for bk := range c.data {
		f(bk)
	}
}

func (c *mergeJoinCache) ForEachWithContext(f func(execute.BlockKey, execute.Trigger, execute.BlockContext)) {
	for bk, tables := range c.data {
		bc := execute.BlockContext{
			Bounds: tables.bounds,
			Count:  tables.Size(),
		}
		f(bk, tables.trigger, bc)
	}
}

func (c *mergeJoinCache) DiscardBlock(key execute.BlockKey) {
	c.data[key].ClearData()
}

func (c *mergeJoinCache) ExpireBlock(key execute.BlockKey) {
	delete(c.data, key)
}

func (c *mergeJoinCache) SetTriggerSpec(spec query.TriggerSpec) {
	c.triggerSpec = spec
}

func (c *mergeJoinCache) Tables(bm execute.BlockMetadata) *joinTables {
	key := execute.ToBlockKey(bm)
	tables := c.data[key]
	if tables == nil {
		tables = &joinTables{
			tags:    bm.Tags(),
			bounds:  bm.Bounds(),
			left:    new(mergeTable),
			right:   new(mergeTable),
			trigger: execute.NewTriggerFromSpec(c.triggerSpec),
		}
		c.data[key] = tables
	}
	return tables
}

type joinTables struct {
	tags   execute.Tags
	bounds execute.Bounds

	left  *mergeTable
	right *mergeTable

	trigger execute.Trigger
}

func (t *joinTables) Bounds() execute.Bounds {
	return t.bounds
}
func (t *joinTables) Tags() execute.Tags {
	return t.tags
}
func (t *joinTables) Size() int {
	return len(t.left.cells) + len(t.right.cells)
}

func (t *joinTables) ClearData() {
	t.left = new(mergeTable)
	t.right = new(mergeTable)
}

func (t *joinTables) Join() execute.Block {
	// Perform sort-merge join

	builder := execute.NewRowListBlockBuilder()
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
					builder.AddCell(execute.Cell{
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

func (t *mergeTable) Insert(value float64, tags execute.Tags, time execute.Time) {
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
