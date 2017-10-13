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
	Keys []string `json:"keys"`
	Keep []string `json:"keep"`
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

	if value, ok := args["keys"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf("keys argument must be a list of strings got %v", value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf("keys argument must be a list of strings, got list of %v", list.Type)
		}
		spec.Keys = list.Elements.([]string)
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
	return spec, nil
}

func newGroupOp() query.OperationSpec {
	return new(GroupOpSpec)
}

func (s *GroupOpSpec) Kind() query.OperationKind {
	return GroupKind
}

type GroupProcedureSpec struct {
	Keys []string
	Keep []string
}

func newGroupProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*GroupOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &GroupProcedureSpec{
		Keys: spec.Keys,
		Keep: spec.Keep,
	}
	return p, nil
}

func (s *GroupProcedureSpec) Kind() plan.ProcedureKind {
	return GroupKind
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
	sort.Strings(spec.Keys)
	t := &groupTransformation{
		d:     d,
		cache: cache,
		keys:  spec.Keys,
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
func (t *groupTransformation) Finish(id execute.DatasetID) {
	t.d.Finish()
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
