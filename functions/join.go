package functions

import (
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
	On   []string              `json:"on"`
	Eval expression.Expression `json:"eval"`
}

func init() {
	ifql.RegisterFunction(JoinKind, createJoinOpSpec)
	query.RegisterOpSpec(JoinKind, newJoinOp)
	//TODO(nathanielc): Allow for other types of join implementations
	plan.RegisterProcedureSpec(MergeJoinKind, newMergeJoinProcedure, JoinKind)
	execute.RegisterTransformation(MergeJoinKind, createMergeJoinTransformation)
}

func createJoinOpSpec(args ifql.Arguments, ctx ifql.Context) (query.OperationSpec, error) {
	eval, err := args.GetRequiredExpression("eval")
	if err != nil {
		return nil, err
	}
	spec := &JoinOpSpec{
		Eval: eval,
	}

	if array, ok, err := args.GetArray("on", ifql.TString); err != nil {
		return nil, err
	} else if ok {
		spec.On = array.Elements.([]string)
	}

	// Find identifier of parent nodes
	err = expression.Walk(spec.Eval.Root, func(n expression.Node) error {
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
	On   []string              `json:"keys"`
	Eval expression.Expression `json:"eval"`
}

func newMergeJoinProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*JoinOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &MergeJoinProcedureSpec{
		On:   spec.On,
		Eval: spec.Eval,
	}
	sort.Strings(p.On)
	return p, nil
}

func (s *MergeJoinProcedureSpec) Kind() plan.ProcedureKind {
	return MergeJoinKind
}
func (s *MergeJoinProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MergeJoinProcedureSpec)

	ns.On = make([]string, len(s.On))
	copy(ns.On, s.On)

	// TODO Copy Expression
	ns.Eval = s.Eval

	return ns
}

func createMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeJoinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	joinEval, err := NewExpressionSpec(s.Eval)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid expression")
	}
	cache := NewMergeJoinCache(joinEval)
	d := execute.NewDataset(id, mode, cache)
	t := NewMergeJoinTransformation(d, cache, s)
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

func NewMergeJoinTransformation(d execute.Dataset, cache MergeJoinCache, spec *MergeJoinProcedureSpec) *mergeJoinTransformation {
	return &mergeJoinTransformation{
		d:           d,
		cache:       cache,
		parentState: make(map[execute.DatasetID]*mergeJoinParentState),
		keys:        spec.On,
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
		tags:   meta.Tags().IntersectingSubset(t.keys),
		bounds: meta.Bounds(),
	}
	t.d.RetractBlock(execute.ToBlockKey(bm))
}

func (t *mergeJoinTransformation) Process(id execute.DatasetID, b execute.Block) {
	t.mu.Lock()
	defer t.mu.Unlock()

	bm := blockMetadata{
		tags:   b.Tags().IntersectingSubset(t.keys),
		bounds: b.Bounds(),
	}
	tables := t.cache.Tables(bm)

	var table execute.BlockBuilder
	switch id {
	case t.leftID:
		table = tables.left
	case t.rightID:
		table = tables.right
	}

	colMap := t.addNewCols(b, table)

	times := b.Times()
	times.DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			execute.AppendRow(i, rr, table, colMap)
		}
	})
}

// addNewCols adds column to builder that exist on b and are part of the join keys.
// This method ensures that the left and right tables always have the same columns.
// A colMap is returned mapping cols of builder to cols of b.
func (t *mergeJoinTransformation) addNewCols(b execute.Block, builder execute.BlockBuilder) []int {
	cols := b.Cols()
	existing := builder.Cols()
	colMap := make([]int, len(existing))
	for j, c := range cols {
		// Skip common tags or tags that are not one of the join keys.
		if c.IsTag {
			if c.IsCommon {
				continue
			}
			found := false
			for _, k := range t.keys {
				if c.Label == k {
					found = true
					break
				}
			}
			// Column is not one of the join keys
			if !found {
				continue
			}
		}
		// Check if column already exists
		found := false
		for ej, ec := range existing {
			if c.Label == ec.Label {
				colMap[ej] = j
				found = true
				break
			}
		}
		// Add new column
		if !found {
			builder.AddCol(c)
			colMap = append(colMap, j)
		}
	}
	return colMap
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

func (t *mergeJoinTransformation) Finish(id execute.DatasetID, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err != nil {
		t.d.Finish(err)
	}

	t.parentState[id].finished = true
	finished := true
	for _, state := range t.parentState {
		finished = finished && state.finished
	}

	if finished {
		t.d.Finish(nil)
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

type MergeJoinCache interface {
	Tables(execute.BlockMetadata) *joinTables
}

type mergeJoinCache struct {
	data map[execute.BlockKey]*joinTables

	triggerSpec query.TriggerSpec

	joinEval *expressionSpec
}

func NewMergeJoinCache(joinEval *expressionSpec) *mergeJoinCache {
	return &mergeJoinCache{
		data:     make(map[execute.BlockKey]*joinTables),
		joinEval: joinEval,
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
			left:     execute.NewColListBlockBuilder(),
			right:    execute.NewColListBlockBuilder(),
			trigger:  execute.NewTriggerFromSpec(c.triggerSpec),
			joinEval: c.joinEval.Copy(),
		}
		tables.left.AddCol(execute.TimeCol)
		tables.right.AddCol(execute.TimeCol)
		c.data[key] = tables
	}
	return tables
}

type joinTables struct {
	tags   execute.Tags
	bounds execute.Bounds

	left, right *execute.ColListBlockBuilder

	trigger execute.Trigger

	joinEval *expressionSpec
}

func (t *joinTables) Bounds() execute.Bounds {
	return t.bounds
}
func (t *joinTables) Tags() execute.Tags {
	return t.tags
}
func (t *joinTables) Size() int {
	return t.left.NRows() + t.right.NRows()
}

func (t *joinTables) ClearData() {
	t.left = execute.NewColListBlockBuilder()
	t.right = execute.NewColListBlockBuilder()
}

// Join performs a sort-merge join
func (t *joinTables) Join() execute.Block {
	// First determine new value type
	newType, err := t.joinEval.Compile(
		execute.ValueCol(t.left.Cols()).Type,
		execute.ValueCol(t.right.Cols()).Type,
	)
	if err != nil {
		//TODO add correct error handling here
		panic(err)
	}
	// Create a builder to the result of the join
	builder := execute.NewColListBlockBuilder()
	builder.SetBounds(t.bounds)
	builder.AddCol(execute.TimeCol)
	builder.AddCol(execute.ColMeta{
		Label: execute.ValueColLabel,
		Type:  newType,
	})
	execute.AddTags(t.tags, builder)
	// Add non common tags
	cols := t.left.Cols()
	for _, c := range cols {
		if c.IsTag {
			builder.AddCol(c)
		}
	}

	// Build colMap from t.left.Cols() to builder.Cols()
	colMap := make([]int, len(cols))
	for j, c := range cols {
		for bj, bc := range builder.Cols() {
			if bc == c {
				colMap[j] = bj
				break
			}
		}
	}

	timeIdx := execute.TimeIdx(builder.Cols())
	valueIdx := execute.ValueIdx(builder.Cols())
	srcValueIdx := execute.ValueIdx(cols)

	// Determine sort order for the joining tables
	sortOrder := make([]string, len(cols))
	for i, c := range cols {
		sortOrder[i] = c.Label
	}
	t.left.Sort(sortOrder, false)
	t.right.Sort(sortOrder, false)

	var (
		left, right       *execute.ColListBlock
		leftSet, rightSet subset
		leftKey, rightKey joinKey
	)
	left = t.left.RawBlock()
	right = t.right.RawBlock()

	leftSet, leftKey = t.advance(leftSet.Stop, left)
	rightSet, rightKey = t.advance(rightSet.Stop, right)
	for !leftSet.Empty() && !rightSet.Empty() {
		if leftKey.Equal(rightKey) {
			// Inner join
			for l := leftSet.Start; l < leftSet.Stop; l++ {
				for r := rightSet.Start; r < rightSet.Stop; r++ {
					// Add time value
					builder.AppendTime(timeIdx, leftKey.Time)

					// Evaluate expression and add to block
					lv := readValue(l, srcValueIdx, left)
					rv := readValue(r, srcValueIdx, right)
					v, err := t.eval(lv, rv)
					if err != nil {
						//TODO add correct error handling here
						panic(err)
					}
					switch newType {
					case execute.TBool:
						builder.AppendBool(valueIdx, v.Bool())
					case execute.TInt:
						builder.AppendInt(valueIdx, v.Int())
					case execute.TUInt:
						builder.AppendUInt(valueIdx, v.UInt())
					case execute.TFloat:
						builder.AppendFloat(valueIdx, v.Float())
					case execute.TString:
						builder.AppendString(valueIdx, v.Str())
					}

					// Append noncommon tags
					for j, c := range cols {
						if c.IsTag && !c.IsCommon {
							builder.AppendString(colMap[j], leftKey.Tags[j-2])
						}
					}
				}
			}
			leftSet, leftKey = t.advance(leftSet.Stop, left)
			rightSet, rightKey = t.advance(rightSet.Stop, right)
		} else if leftKey.Less(rightKey) {
			leftSet, leftKey = t.advance(leftSet.Stop, left)
		} else {
			rightSet, rightKey = t.advance(rightSet.Stop, right)
		}
	}
	return builder.Block()
}

func (t *joinTables) advance(offset int, table *execute.ColListBlock) (subset, joinKey) {
	if n := table.NRows(); n == offset {
		return subset{Start: n, Stop: n}, joinKey{}
	}
	start := offset
	key := rowKey(start, table)
	s := subset{Start: start}
	offset++
	for offset < table.NRows() && equalRowKeys(start, offset, table) {
		offset++
	}
	s.Stop = offset
	return s, key
}

type subset struct {
	Start int
	Stop  int
}

func (s subset) Empty() bool {
	return s.Start == s.Stop
}

func rowKey(i int, table *execute.ColListBlock) (k joinKey) {
	for j, c := range table.Cols() {
		if c.Label == execute.TimeColLabel {
			k.Time = table.AtTime(i, j)
		} else if c.IsTag {
			k.Tags = append(k.Tags, table.AtString(i, j))
		}
	}
	return
}

func equalRowKeys(x, y int, table *execute.ColListBlock) bool {
	for j, c := range table.Cols() {
		if c.Label == execute.TimeColLabel {
			if table.AtTime(x, j) != table.AtTime(y, j) {
				return false
			}
		} else if c.IsTag {
			if table.AtString(x, j) != table.AtString(y, j) {
				return false
			}
		}
	}
	return true
}

func readValue(i, valueIdx int, table *execute.ColListBlock) execute.Value {
	var v interface{}
	cols := table.Cols()
	switch cols[valueIdx].Type {
	case execute.TBool:
		v = table.AtBool(i, valueIdx)
	case execute.TInt:
		v = table.AtInt(i, valueIdx)
	case execute.TUInt:
		v = table.AtUInt(i, valueIdx)
	case execute.TFloat:
		v = table.AtFloat(i, valueIdx)
	case execute.TString:
		v = table.AtString(i, valueIdx)
	}
	return execute.Value{
		Type:  cols[valueIdx].Type,
		Value: v,
	}
}

type joinKey struct {
	Time execute.Time
	Tags []string
}

func (k joinKey) Equal(o joinKey) bool {
	if k.Time == o.Time {
		for i := range k.Tags {
			if k.Tags[i] != o.Tags[i] {
				return false
			}
		}
		return true
	}
	return false
}
func (k joinKey) Less(o joinKey) bool {
	if k.Time == o.Time {
		for i := range k.Tags {
			if k.Tags[i] != o.Tags[i] {
				return k.Tags[i] < o.Tags[i]
			}
		}
	}
	return k.Time < o.Time
}

func (t *joinTables) eval(l, r execute.Value) (execute.Value, error) {
	return t.joinEval.Eval(l, r)
}

type expressionSpec struct {
	expr                expression.Expression
	leftName, rightName string

	scope execute.Scope

	leftType, rightType execute.DataType
	ce                  execute.CompiledExpression
}

func NewExpressionSpec(expr expression.Expression) (*expressionSpec, error) {
	names := execute.ExpressionNames(expr.Root)
	if len(names) != 2 {
		return nil, fmt.Errorf("join expression can only have two tables, got names: %v", names)
	}
	return &expressionSpec{
		leftName:  names[0],
		rightName: names[1],
		expr:      expr,
		scope:     make(execute.Scope, 2),
	}, nil
}

func (s *expressionSpec) Copy() *expressionSpec {
	cpy := new(expressionSpec)
	*cpy = *s
	cpy.scope = make(execute.Scope, 2)
	return cpy
}

func (s *expressionSpec) Compile(l, r execute.DataType) (execute.DataType, error) {
	if s.ce != nil && l == s.leftType && r == s.rightType {
		// Nothing to do, we already have a compiled expression
		return execute.TInvalid, nil
	}
	ce, err := execute.CompileExpression(s.expr, map[string]execute.DataType{s.leftName: l, s.rightName: r})
	if err != nil {
		s.ce = nil
		return execute.TInvalid, err
	}
	s.ce = ce
	s.leftType = l
	s.rightType = r
	return s.ce.Type(), nil
}

func (s *expressionSpec) Eval(l, r execute.Value) (execute.Value, error) {
	if s.ce == nil {
		return execute.Value{}, errors.New("expression has not been compiled")
	}
	s.scope[s.leftName] = l
	s.scope[s.rightName] = r
	return s.ce.Eval(s.scope)
}
