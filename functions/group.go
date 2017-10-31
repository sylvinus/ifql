package functions

import (
	"fmt"
	"sort"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const GroupKind = "group"

type GroupOpSpec struct {
	By     []string `json:"by"`
	Keep   []string `json:"keep"`
	Ignore []string `json:"ignore"`
}

func init() {
	ifql.RegisterFunction(GroupKind, createGroupOpSpec)
	query.RegisterOpSpec(GroupKind, newGroupOp)
	plan.RegisterProcedureSpec(GroupKind, newGroupProcedure, GroupKind)
	execute.RegisterTransformation(GroupKind, createGroupTransformation)
}

func createGroupOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(GroupOpSpec)
	if len(args) == 0 {
		return spec, nil
	}

	if value, ok := args["by"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf("'by' argument must be a list of strings got %v", value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf("'by' argument must be a list of strings, got list of %v", list.Type)
		}
		spec.By = list.Elements.([]string)
	}

	if value, ok := args["keep"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf("keep argument must be a list of strings got %v", value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf("keep argument must be a list of strings, got list of %v", list.Type)
		}
		spec.Keep = list.Elements.([]string)
	}
	if value, ok := args["ignore"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf("ignore argument must be a list of strings got %v", value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf("ignore argument must be a list of strings, got list of %v", list.Type)
		}
		spec.Ignore = list.Elements.([]string)
	}
	return spec, nil
}

func newGroupOp() query.OperationSpec {
	return new(GroupOpSpec)
}

func (s *GroupOpSpec) Kind() query.OperationKind {
	return GroupKind
}

type GroupProcedureSpec struct {
	By     []string
	Ignore []string
	Keep   []string
}

func newGroupProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*GroupOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &GroupProcedureSpec{
		By:     spec.By,
		Ignore: spec.Ignore,
		Keep:   spec.Keep,
	}
	return p, nil
}

func (s *GroupProcedureSpec) Kind() plan.ProcedureKind {
	return GroupKind
}
func (s *GroupProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(GroupProcedureSpec)

	ns.By = make([]string, len(s.By))
	copy(ns.By, s.By)

	ns.Ignore = make([]string, len(s.Ignore))
	copy(ns.Ignore, s.Ignore)

	ns.Keep = make([]string, len(s.Keep))
	copy(ns.Keep, s.Keep)

	return ns
}

func (s *GroupProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: []plan.ProcedureKind{LimitKind, RangeKind, WhereKind},
	}
}

func (s *GroupProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.GroupingSet {
		root = dup()
		selectSpec = root.Spec.(*SelectProcedureSpec)
		selectSpec.OrderByTime = false
		selectSpec.GroupingSet = false
		selectSpec.MergeAll = false
		selectSpec.GroupKeys = nil
		selectSpec.GroupIgnore = nil
		selectSpec.GroupKeep = nil
		return
	}
	selectSpec.GroupingSet = true
	// TODO implement OrderByTime
	//selectSpec.OrderByTime = true

	// Merge all series into a single group if we have no specific grouping dimensions.
	selectSpec.MergeAll = len(s.By) == 0 && len(s.Ignore) == 0
	selectSpec.GroupKeys = s.By
	selectSpec.GroupIgnore = s.Ignore
	selectSpec.GroupKeep = s.Keep
}

func createGroupTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*GroupProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := newGroupTransformation(d, cache, s)
	return t, d, nil
}

type groupTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	keys []string
	keep []string
}

func newGroupTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *GroupProcedureSpec) *groupTransformation {
	sort.Strings(spec.By)
	t := &groupTransformation{
		d:     d,
		cache: cache,
		keys:  spec.By,
		keep:  spec.Keep,
	}
	sort.Strings(t.keys)
	sort.Strings(t.keep)
	return t
}

func (t *groupTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.cache.ForEachBuilder(func(bk execute.BlockKey, builder execute.BlockBuilder) {
		if meta.Bounds().Equal(builder.Bounds()) {
			t.d.RetractBlock(bk)
		}
	})
}

func (t *groupTransformation) Process(id execute.DatasetID, b execute.Block) {
	tags := b.Tags().Subset(t.keys)
	builder, new := t.cache.BlockBuilder(blockMetadata{
		tags:   tags,
		bounds: b.Bounds(),
	})
	if new {
		// Determine columns of new block.

		// Add tags.
		execute.AddTags(tags, builder)

		// Add other existing columns, skipping tags.
		for _, c := range b.Cols() {
			if !c.IsTag {
				builder.AddCol(c)
			}
		}

		// Add columns for tags that are to be kept.
		for _, k := range t.keep {
			builder.AddCol(execute.ColMeta{
				Label: k,
				Type:  execute.TString,
				IsTag: true,
			})
		}
	}

	// Construct map of builder column index to block column index.
	builderCols := builder.Cols()
	blockCols := b.Cols()
	colMap := make([]int, len(builderCols))
	for j, c := range builderCols {
		for nj, nc := range blockCols {
			if c.Label == nc.Label {
				colMap[j] = nj
				break
			}
		}
	}

	execute.AppendBlock(b, builder, colMap)
}

func (t *groupTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.d.UpdateWatermark(mark)
}
func (t *groupTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *groupTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
func (t *groupTransformation) SetParents(ids []execute.DatasetID) {
}

type blockMetadata struct {
	tags   execute.Tags
	bounds execute.Bounds
}

func (m blockMetadata) Tags() execute.Tags {
	return m.tags
}
func (m blockMetadata) Bounds() execute.Bounds {
	return m.bounds
}
