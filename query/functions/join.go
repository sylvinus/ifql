package functions

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

const JoinKind = "join"
const MergeJoinKind = "merge-join"

type JoinOpSpec struct {
	Keys       []string        `json:"keys"`
	Expression expression.Node `json:"expression"`
}

func init() {
	ifql.RegisterFunction(JoinKind, createJoinOpSpec)
	query.RegisterOpSpec(JoinKind, newJoinOp)
	//TODO(nathanielc): Allow for other types of join implementations
	plan.RegisterProcedureSpec(MergeJoinKind, newMergeJoinProcedure, JoinKind)
	execute.RegisterTransformation(MergeJoinKind, createMergeJoinTransformation)
}

func createJoinOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	expValue, ok := args["exp"]
	if !ok {
		return nil, errors.New(`join function requires an argument "exp"`)
	}
	if expValue.Type != ifql.TExpression {
		return nil, fmt.Errorf(`join function argument "exp" must be an expression, got %v`, expValue.Type)
	}
	expr := expValue.Value.(expression.Node)
	spec := &JoinOpSpec{
		Expression: expr,
	}

	if keysValue, ok := args["keys"]; ok {
		if keysValue.Type != ifql.TArray {
			return nil, fmt.Errorf(`join argument "keys" must be a list, got %v`, keysValue.Type)
		}
		list := keysValue.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf(`join argument "keys" must be a list of strings, got list of %v`, list.Type)
		}
		spec.Keys = list.Elements.([]string)
	}

	// Find identifier of parent nodes
	err := expression.Walk(expr, func(n expression.Node) error {
		if r, ok := n.(*expression.ReferenceNode); ok && r.Kind == "identifier" {
			id, err := ctx.LookupIDFromIdentifier(r.Name)
			if err != nil {
				return err
			}
			ctx.AdditionalParent(id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func newJoinOp() query.OperationSpec {
	return new(JoinOpSpec)
}

func (s *JoinOpSpec) Kind() query.OperationKind {
	return JoinKind
}

type MergeJoinProcedureSpec struct {
	Keys       []string        `json:"keys"`
	Expression expression.Node `json:"expression"`
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

func createMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeJoinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	joinExpr, err := newExpressionSpec(s.Expression)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid expression")
	}
	cache := newMergeJoinCache(joinExpr)
	d := execute.NewDataset(id, mode, cache)
	t := newMergeJoinTransformation(d, cache, s)
	return t, d, nil
}

type mergeJoinTransformation struct {
	parents []execute.DatasetID

	mu sync.Mutex

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
		keys:        spec.Keys,
	}
}

type mergeJoinParentState struct {
	mark       execute.Time
	processing execute.Time
	finished   bool
}

func (t *mergeJoinTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	t.mu.Lock()
	defer t.mu.Unlock()

	bm := blockMetadata{
		tags:   meta.Tags().Subset(t.keys),
		bounds: meta.Bounds(),
	}
	t.d.RetractBlock(execute.ToBlockKey(bm))
}

func (t *mergeJoinTransformation) Process(id execute.DatasetID, b execute.Block) {
	t.mu.Lock()
	defer t.mu.Unlock()

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

	cols := b.Cols()
	valueIdx := execute.ValueIdx(b)
	times := b.Times()
	times.DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i, time := range ts {
			v := rr.AtFloat(i, valueIdx)
			tags := execute.TagsForRow(cols, rr, i).Subset(t.keys)
			table.Insert(v, tags, time)
		}
	})
}

func (t *mergeJoinTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
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
	t.mu.Lock()
	defer t.mu.Unlock()
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
	t.mu.Lock()
	defer t.mu.Unlock()

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
	t.mu.Lock()
	defer t.mu.Unlock()

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

	joinExpr *expressionSpec
}

func newMergeJoinCache(joinExpr *expressionSpec) *mergeJoinCache {
	return &mergeJoinCache{
		data:     make(map[execute.BlockKey]*joinTables),
		joinExpr: joinExpr,
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
			tags:     bm.Tags(),
			bounds:   bm.Bounds(),
			left:     new(mergeTable),
			right:    new(mergeTable),
			trigger:  execute.NewTriggerFromSpec(c.triggerSpec),
			joinExpr: c.joinExpr.copy(),
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

	joinExpr *expressionSpec
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

	builder := execute.NewColListBlockBuilder()
	builder.SetBounds(t.bounds)
	builder.SetTags(t.tags)
	builder.AddCol(execute.TimeCol)
	builder.AddCol(execute.ValueCol)

	var left, leftSet, right, rightSet []joinCell
	var leftKey, rightKey joinKey
	left = t.left.Sorted()
	right = t.right.Sorted()

	//log.Println("tags", t.tags)
	//log.Println("left", tabularFmt(left))
	//log.Println("right", tabularFmt(right))

	left, leftSet, leftKey = t.advance(left)
	right, rightSet, rightKey = t.advance(right)
	for len(leftSet) > 0 && len(rightSet) > 0 {
		if leftKey == rightKey {
			// Inner join
			for _, l := range leftSet {
				for _, r := range rightSet {
					v := t.eval(l.Value, r.Value)
					builder.AppendTime(0, l.Key.Time)
					builder.AppendFloat(1, v)
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
	return t.joinExpr.eval(l, r)
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

type tabularFmt []joinCell

func (t tabularFmt) String() string {
	if len(t) == 0 {
		return "<empty table>"
	}
	var buf bytes.Buffer
	n := 0
	fmt.Fprintf(&buf, "Table:\n%5s", "#")
	n += 5
	fmt.Fprintf(&buf, "%31s", "Time")
	n += 31
	keys := t[0].Tags.Keys()
	for _, k := range keys {
		fmt.Fprintf(&buf, "%20s", k)
		n += 20
	}
	fmt.Fprintf(&buf, "%20s", "Value")
	buf.WriteRune('\n')
	n += 20
	for i := 0; i < n; i++ {
		buf.WriteRune('-')
	}
	buf.WriteRune('\n')

	for i, c := range t {
		fmt.Fprintf(&buf, "%5d", i)
		fmt.Fprintf(&buf, "%31v", c.Key.Time)
		for _, k := range keys {
			fmt.Fprintf(&buf, "%20s", c.Tags[k])
		}
		fmt.Fprintf(&buf, "%20v", c.Value)
		buf.WriteRune('\n')
	}
	return buf.String()
}

type expressionSpec struct {
	expr                expression.Node
	leftName, rightName string
	scope               execute.Scope
}

func newExpressionSpec(expr expression.Node) (*expressionSpec, error) {
	names := execute.ExpressionNames(expr)
	if len(names) != 2 {
		return nil, fmt.Errorf("join expression can only have two tables, got names: %v", names)
	}
	rightName := names[0]
	if rightName == "$" {
		rightName = names[1]
	}
	return &expressionSpec{
		leftName:  "$",
		rightName: rightName,
		scope:     make(execute.Scope, 2),
		expr:      expr,
	}, nil
}

func (s *expressionSpec) eval(l, r float64) float64 {
	s.scope[s.leftName] = l
	s.scope[s.rightName] = r
	// Ignore the error since we validated the names already
	v, _ := execute.EvalExpression(s.expr, s.scope)
	return v
}

func (s *expressionSpec) copy() *expressionSpec {
	cpy := new(expressionSpec)
	*cpy = *s
	cpy.scope = make(execute.Scope, 2)
	return cpy
}
