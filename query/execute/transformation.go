package execute

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Transformation interface {
	RetractBlock(meta BlockMetadata)
	Process(id DatasetID, b Block)
	UpdateWatermark(t Time)
	UpdateProcessingTime(t Time)
	Finish()
}

func transformationFromProcedureSpec(d Dataset, spec plan.ProcedureSpec, now time.Time) Transformation {
	switch s := spec.(type) {
	case *plan.SumProcedureSpec:
		return &aggregateTransformation{
			d:   d,
			agg: new(sumAgg),
		}
	case *plan.CountProcedureSpec:
		return &aggregateTransformation{
			d:   d,
			agg: new(countAgg),
		}
	case *plan.MergeProcedureSpec:
		return newMergeTransformation(d, s)
	case *plan.MergeJoinProcedureSpec:
		return newMergeJoinTransformation(d, s)
	case *plan.WindowProcedureSpec:
		return newFixedWindowTransformation(d, Window{
			Every:  Duration(s.Window.Every),
			Period: Duration(s.Window.Period),
			Round:  Duration(s.Window.Round),
			Start:  Time(s.Window.Start.Time(now).UnixNano()),
		})
	default:
		//TODO add proper error handling
		panic(fmt.Sprintf("unsupported procedure %v", spec.Kind()))
	}
}

type fixedWindowTransformation struct {
	d Dataset
	w Window
}

func newFixedWindowTransformation(d Dataset, w Window) Transformation {
	return &fixedWindowTransformation{
		d: d,
		w: w,
	}
}

func (t *fixedWindowTransformation) RetractBlock(meta BlockMetadata) {
	tagKey := meta.Tags().Key()
	t.d.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
		if bld.Bounds().Overlaps(meta.Bounds()) && tagKey == bld.Tags().Key() {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *fixedWindowTransformation) Process(id DatasetID, b Block) {
	tagKey := b.Tags().Key()

	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, c := range cs {
			found := false
			t.d.ForEachBuilder(func(bk BlockKey, bld BlockBuilder) {
				if bld.Bounds().Contains(c.Time) && tagKey == bld.Tags().Key() {
					bld.AddCell(c)
					found = true
				}
			})
			if !found {
				builder := t.d.BlockBuilder(blockMetadata{
					tags:   b.Tags(),
					bounds: t.getWindowBounds(c.Time),
				})
				builder.AddCell(c)
			}
		}
	})
}

func (t *fixedWindowTransformation) getWindowBounds(time Time) Bounds {
	stop := time.Truncate(t.w.Every)
	stop += Time(t.w.Every)
	return Bounds{
		Stop:  stop,
		Start: stop - Time(t.w.Period),
	}
}

func (t *fixedWindowTransformation) UpdateWatermark(mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *fixedWindowTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *fixedWindowTransformation) Finish() {
	t.d.Finish()
}

type mergeTransformation struct {
	d    Dataset
	keys []string
}

func newMergeTransformation(d Dataset, spec *plan.MergeProcedureSpec) *mergeTransformation {
	sort.Strings(spec.Keys)
	return &mergeTransformation{
		d:    d,
		keys: spec.Keys,
	}
}

func (t *mergeTransformation) RetractBlock(meta BlockMetadata) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.d.ForEachBuilder(func(bk BlockKey, builder BlockBuilder) {
		if meta.Bounds().Equal(builder.Bounds()) {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *mergeTransformation) Process(id DatasetID, b Block) {
	builder := t.d.BlockBuilder(blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	})
	cells := b.Cells()
	cells.Do(func(cs []Cell) {
		for _, c := range cs {
			builder.AddCell(c)
		}
	})
}

func (t *mergeTransformation) UpdateWatermark(mark Time) {
	t.d.UpdateWatermark(mark)
}
func (t *mergeTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *mergeTransformation) Finish() {
	t.d.Finish()
}

type mergeJoinTransformation struct {
	d             Dataset
	finishedCount int

	leftID  DatasetID
	rightID DatasetID

	keys []string

	data map[Bounds]*joinTables
}

type joinTables struct {
	tags   Tags
	bounds Bounds

	left  *mergeTable
	right *mergeTable
}

func (t *joinTables) Bounds() Bounds {
	return t.bounds
}
func (t joinTables) Tags() Tags {
	return t.tags
}

func newMergeJoinTransformation(d Dataset, spec *plan.MergeJoinProcedureSpec) *mergeJoinTransformation {
	return &mergeJoinTransformation{
		d:    d,
		data: make(map[Bounds]*joinTables),
	}
}

func (t *mergeJoinTransformation) RetractBlock(meta BlockMetadata) {
	//TODO: How can we handle block retraction when we would have to cache lots of intermediate data?
}

func (t *mergeJoinTransformation) Process(id DatasetID, b Block) {
	log.Println("Process", id)

	//TODO(nathanielc): Which dataset ID is left or right?
	// Hack for now assume left is the first one
	if t.leftID.IsZero() {
		t.leftID = id
	} else if t.leftID != id && t.rightID.IsZero() {
		t.rightID = id
	}
	tables := t.data[b.Bounds()]
	if tables == nil {
		tables = &joinTables{
			tags:   b.Tags().Subset(t.keys),
			bounds: b.Bounds(),
			left:   new(mergeTable),
			right:  new(mergeTable),
		}
		t.data[b.Bounds()] = tables
	}

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

func (t *mergeJoinTransformation) UpdateWatermark(mark Time) {
	//TODO(nathanielc): This implementation assumes that triggering is based off watermarks and nothing else.
	// We need a way to consume triggers within the transformation.
	log.Println("mark", mark)

	//TODO Need to track watermark per parent

	t.d.UpdateWatermark(mark)
}
func (t *mergeJoinTransformation) UpdateProcessingTime(pt Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *mergeJoinTransformation) Finish() {
	t.finishedCount++
	// TODO find parent count.
	// TODO: Should all joins be 2 parent joins?
	// How do we do opaque functions that are of three or more fields?
	if t.finishedCount == 2 {
		// Join tables that are below the watermark
		for _, tables := range t.data {
			t.join(tables)
		}
		t.d.Finish()
	}
}

func (t *mergeJoinTransformation) join(tables *joinTables) {
	// Perform sort-merge join
	log.Println("join", tables)

	builder := t.d.BlockBuilder(tables)

	var left, leftSet, right, rightSet []joinCell
	var leftKey, rightKey joinKey
	left = tables.left.Sorted()
	right = tables.right.Sorted()

	left, leftSet, leftKey = t.advance(left)
	right, rightSet, rightKey = t.advance(right)
	for len(leftSet) > 0 && len(rightSet) > 0 {
		if leftKey == rightKey {
			log.Println("leftKey", leftKey)
			// Inner join
			for _, l := range leftSet {
				for _, r := range rightSet {
					v := t.eval(l.Value, r.Value)
					log.Println("value", v)
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
}

func (t *mergeJoinTransformation) advance(table cells) ([]joinCell, []joinCell, joinKey) {
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

func (t *mergeJoinTransformation) eval(l, r float64) float64 {
	// TODO perform specified expression
	log.Println(l, r)
	return r / l
}

type mergeTable struct {
	cells []joinCell
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

type cells []joinCell

func (c cells) Len() int               { return len(c) }
func (c cells) Less(i int, j int) bool { return c[i].Key.Less(c[j].Key) }
func (c cells) Swap(i int, j int)      { c[i], c[j] = c[j], c[i] }

func (t *mergeTable) Sorted() []joinCell {
	sort.Sort(cells(t.cells))
	return t.cells
}
