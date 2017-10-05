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
	Keys []string `json:"keys"`
}

func newGroupProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*GroupOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &GroupProcedureSpec{
		Keys: spec.Keys,
	}
	sort.Strings(p.Keys)
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
	d       execute.Dataset
	cache   execute.BlockBuilderCache
	keys    []string
	parents []execute.DatasetID
}

func newGroupTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *GroupProcedureSpec) *groupTransformation {
	sort.Strings(spec.Keys)
	return &groupTransformation{
		d:     d,
		cache: cache,
		keys:  spec.Keys,
	}
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
	builder, new := t.cache.BlockBuilder(blockMetadata{
		tags:   b.Tags().Subset(t.keys),
		bounds: b.Bounds(),
	})
	cols := b.Cols()
	nj := 0
	for j, c := range cols {
		// TODO check the `keep` list to determine which tags are kept
		if c.IsTag {
			continue
		}
		if new {
			builder.AddCol(c)
		}

		values := b.Col(j)
		switch c.Type {
		case execute.TString:
			values.DoString(func(vs []string, _ execute.RowReader) {
				builder.AppendStrings(nj, vs)
			})
		case execute.TFloat:
			values.DoFloat(func(vs []float64, _ execute.RowReader) {
				builder.AppendFloats(nj, vs)
			})
		case execute.TTime:
			values.DoTime(func(vs []execute.Time, _ execute.RowReader) {
				builder.AppendTimes(nj, vs)
			})
		}
		nj++
	}
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
	t.parents = ids
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
