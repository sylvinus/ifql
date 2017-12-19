package functions

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

const JoinKind = "join"
const MergeJoinKind = "merge-join"

type JoinOpSpec struct {
	// On is a list of tags on which to join.
	On []string `json:"on"`
	// Fn is a function accepting a single parameter.
	// The parameter is map if records for each of the parent operations.
	Fn *ast.ArrowFunctionExpression `json:"fn"`
	// TableNames are the names to give to each parent when populating the parameter for the function.
	// The first parent is referenced by the first name and so forth.
	// TODO(nathanielc): Change this to a map of parent operation IDs to names.
	// Then make it possible for the transformation to map operation IDs to parent IDs.
	TableNames map[query.OperationID]string `json:"table_names"`
}

func init() {
	query.RegisterFunction(JoinKind, createJoinOpSpec)
	query.RegisterOpSpec(JoinKind, newJoinOp)
	//TODO(nathanielc): Allow for other types of join implementations
	plan.RegisterProcedureSpec(MergeJoinKind, newMergeJoinProcedure, JoinKind)
	execute.RegisterTransformation(MergeJoinKind, createMergeJoinTransformation)
}

func createJoinOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}
	resolved, err := f.Resolve()
	if err != nil {
		return nil, err
	}
	spec := &JoinOpSpec{
		Fn:         resolved,
		TableNames: make(map[query.OperationID]string),
	}

	if array, ok, err := args.GetArray("on", ifql.TString); err != nil {
		return nil, err
	} else if ok {
		spec.On = array.AsStrings()
	}

	if m, ok, err := args.GetMap("tables"); err != nil {
		return nil, err
	} else if ok {
		for k, t := range m.Elements {
			if t.Type() != query.TTable {
				return nil, fmt.Errorf("tables key %q must be a table: got %v", k, t.Type())
			}
			id := t.Value().(query.Table).ID
			a.AddParent(id)
			spec.TableNames[id] = k
		}
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
	On         []string                     `json:"keys"`
	Fn         *ast.ArrowFunctionExpression `json:"f"`
	TableNames map[plan.ProcedureID]string  `json:"table_names"`
}

func newMergeJoinProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*JoinOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	tableNames := make(map[plan.ProcedureID]string, len(spec.TableNames))
	for qid, name := range spec.TableNames {
		pid := pa.ConvertID(qid)
		tableNames[pid] = name
	}

	p := &MergeJoinProcedureSpec{
		On:         spec.On,
		Fn:         spec.Fn,
		TableNames: tableNames,
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

	ns.Fn = s.Fn.Copy().(*ast.ArrowFunctionExpression)

	return ns
}

func (s *MergeJoinProcedureSpec) ParentChanged(old, new plan.ProcedureID) {
	if v, ok := s.TableNames[old]; ok {
		delete(s.TableNames, old)
		s.TableNames[new] = v
	}
}

func createMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeJoinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	parents := a.Parents()
	if len(parents) != 2 {
		//TODO(nathanielc): Support n-way joins
		return nil, nil, errors.New("joins currently must only have two parents")
	}
	leftID := parents[0]
	rightID := parents[1]

	tableNames := make(map[execute.DatasetID]string, len(s.TableNames))
	for pid, name := range s.TableNames {
		id := a.ConvertID(pid)
		tableNames[id] = name
	}

	joinFn, err := NewJoinFunction(s.Fn, parents, tableNames)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid expression")
	}
	cache := NewMergeJoinCache(joinFn, a.Allocator(), leftID, rightID)
	d := execute.NewDataset(id, mode, cache)
	t := NewMergeJoinTransformation(d, cache, s, parents)
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

func NewMergeJoinTransformation(d execute.Dataset, cache MergeJoinCache, spec *MergeJoinProcedureSpec, parents []execute.DatasetID) *mergeJoinTransformation {
	t := &mergeJoinTransformation{
		d:       d,
		cache:   cache,
		keys:    spec.On,
		leftID:  parents[0],
		rightID: parents[1],
	}
	t.parentState = make(map[execute.DatasetID]*mergeJoinParentState)
	for _, id := range parents {
		t.parentState[id] = new(mergeJoinParentState)
	}
	return t
}

type mergeJoinParentState struct {
	mark       execute.Time
	processing execute.Time
	finished   bool
}

func (t *mergeJoinTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	bm := blockMetadata{
		tags:   meta.Tags().IntersectingSubset(t.keys),
		bounds: meta.Bounds(),
	}
	return t.d.RetractBlock(execute.ToBlockKey(bm))
}

func (t *mergeJoinTransformation) Process(id execute.DatasetID, b execute.Block) error {
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
	return nil
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

func (t *mergeJoinTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.parentState[id].mark = mark

	min := execute.Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.mark < min {
			min = state.mark
		}
	}

	return t.d.UpdateWatermark(min)
}

func (t *mergeJoinTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.parentState[id].processing = pt

	min := execute.Time(math.MaxInt64)
	for _, state := range t.parentState {
		if state.processing < min {
			min = state.processing
		}
	}

	return t.d.UpdateProcessingTime(min)
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

type MergeJoinCache interface {
	Tables(execute.BlockMetadata) *joinTables
}

type mergeJoinCache struct {
	data  map[execute.BlockKey]*joinTables
	alloc *execute.Allocator

	leftID, rightID execute.DatasetID

	triggerSpec query.TriggerSpec

	joinFn *joinFunc
}

func NewMergeJoinCache(joinFn *joinFunc, a *execute.Allocator, leftID, rightID execute.DatasetID) *mergeJoinCache {
	return &mergeJoinCache{
		data:    make(map[execute.BlockKey]*joinTables),
		joinFn:  joinFn,
		alloc:   a,
		leftID:  leftID,
		rightID: rightID,
	}
}

func (c *mergeJoinCache) BlockMetadata(key execute.BlockKey) execute.BlockMetadata {
	return c.data[key]
}

func (c *mergeJoinCache) Block(key execute.BlockKey) (execute.Block, error) {
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
			alloc:   c.alloc,
			left:    execute.NewColListBlockBuilder(c.alloc),
			right:   execute.NewColListBlockBuilder(c.alloc),
			leftID:  c.leftID,
			rightID: c.rightID,
			trigger: execute.NewTriggerFromSpec(c.triggerSpec),
			joinFn:  c.joinFn.Copy(),
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

	alloc *execute.Allocator

	left, right     *execute.ColListBlockBuilder
	leftID, rightID execute.DatasetID

	trigger execute.Trigger

	joinFn *joinFunc
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
	t.left = execute.NewColListBlockBuilder(t.alloc)
	t.right = execute.NewColListBlockBuilder(t.alloc)
}

// Join performs a sort-merge join
func (t *joinTables) Join() (execute.Block, error) {
	// First determine new value type
	newType, err := t.joinFn.Compile(map[execute.DatasetID]execute.DataType{
		t.leftID:  execute.ValueCol(t.left.Cols()).Type,
		t.rightID: execute.ValueCol(t.right.Cols()).Type,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile join function")
	}
	// Create a builder to the result of the join
	builder := execute.NewColListBlockBuilder(t.alloc)
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
					v, err := t.joinFn.Eval(map[execute.DatasetID]execute.Value{
						t.leftID:  lv,
						t.rightID: rv,
					})
					if err != nil {
						return nil, errors.Wrap(err, "failed to evaluate join function")
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

func (t *joinTables) eval(values map[execute.DatasetID]execute.Value) (execute.Value, error) {
	return t.joinFn.Eval(values)
}

type joinFunc struct {
	f              *ast.ArrowFunctionExpression
	referencePaths map[execute.DatasetID]execute.ReferencePath

	scope execute.Scope

	ce    execute.CompiledExpression
	types map[execute.DatasetID]execute.DataType
}

func NewJoinFunction(f *ast.ArrowFunctionExpression, parentIDs []execute.DatasetID, tableNames map[execute.DatasetID]string) (*joinFunc, error) {
	if len(f.Params) != 1 {
		return nil, errors.New("join function should only have one parameter for the map of tables")
	}
	tableParam := f.Params[0].Name
	referencePaths := make(map[execute.DatasetID]execute.ReferencePath, len(parentIDs))
	for _, pid := range parentIDs {
		referencePaths[pid] = execute.Reference{
			tableParam,
			tableNames[pid],
			"_value",
		}.Path()
	}
	return &joinFunc{
		f:              f,
		scope:          make(execute.Scope, 2),
		referencePaths: referencePaths,
	}, nil
}

func (s *joinFunc) Copy() *joinFunc {
	cpy := new(joinFunc)
	*cpy = *s
	cpy.scope = make(execute.Scope, 2)
	return cpy
}

func (s *joinFunc) Compile(types map[execute.DatasetID]execute.DataType) (execute.DataType, error) {
	if s.ce != nil && equalTypes(types, s.types) {
		// Nothing to do, we already have a compiled expression
		return s.ce.Type(), nil
	}
	referenceTypes := make(map[execute.ReferencePath]execute.DataType, len(types))
	for pid, typ := range types {
		referenceTypes[s.referencePaths[pid]] = typ
	}
	ce, err := execute.CompileExpression(s.f, referenceTypes)
	if err != nil {
		s.ce = nil
		s.types = nil
		return execute.TInvalid, err
	}
	s.ce = ce
	// Copy types
	s.types = make(map[execute.DatasetID]execute.DataType, len(types))
	for k, v := range types {
		s.types[k] = v
	}
	return s.ce.Type(), nil
}

func equalTypes(a, b map[execute.DatasetID]execute.DataType) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func (s *joinFunc) Eval(values map[execute.DatasetID]execute.Value) (execute.Value, error) {
	if s.ce == nil {
		return execute.Value{}, errors.New("expression has not been compiled")
	}
	for pid, v := range values {
		s.scope[s.referencePaths[pid]] = v
	}
	return s.ce.Eval(s.scope)
}
