package functions

import (
	"errors"
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
	Except []string `json:"except"`
}

func init() {
	query.RegisterMethod(GroupKind, createGroupOpSpec)
	query.RegisterOpSpec(GroupKind, newGroupOp)
	plan.RegisterProcedureSpec(GroupKind, newGroupProcedure, GroupKind)
	execute.RegisterTransformation(GroupKind, createGroupTransformation)
}

func createGroupOpSpec(args query.Arguments, ctx *query.Context) (query.OperationSpec, error) {
	spec := new(GroupOpSpec)
	if array, ok, err := args.GetArray("by", ifql.TString); err != nil {
		return nil, err
	} else if ok {
		spec.By = array.AsStrings()
	}
	if array, ok, err := args.GetArray("keep", ifql.TString); err != nil {
		return nil, err
	} else if ok {
		spec.Keep = array.AsStrings()
	}
	if array, ok, err := args.GetArray("except", ifql.TString); err != nil {
		return nil, err
	} else if ok {
		spec.Except = array.AsStrings()
	}

	if len(spec.By) > 0 && len(spec.Except) > 0 {
		return nil, errors.New(`cannot specify both "by" and "except" keyword arguments`)
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
	Except []string
	Keep   []string
}

func newGroupProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*GroupOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &GroupProcedureSpec{
		By:     spec.By,
		Except: spec.Except,
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

	ns.Except = make([]string, len(s.Except))
	copy(ns.Except, s.Except)

	ns.Keep = make([]string, len(s.Keep))
	copy(ns.Keep, s.Keep)

	return ns
}

func (s *GroupProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    FromKind,
		Through: []plan.ProcedureKind{LimitKind, RangeKind, FilterKind},
		Match: func(spec plan.ProcedureSpec) bool {
			selectSpec := spec.(*FromProcedureSpec)
			return !selectSpec.AggregateSet
		},
	}
}

func (s *GroupProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*FromProcedureSpec)
	if selectSpec.GroupingSet {
		root = dup()
		selectSpec = root.Spec.(*FromProcedureSpec)
		selectSpec.OrderByTime = false
		selectSpec.GroupingSet = false
		selectSpec.MergeAll = false
		selectSpec.GroupKeys = nil
		selectSpec.GroupExcept = nil
		selectSpec.GroupKeep = nil
		return
	}
	selectSpec.GroupingSet = true
	// TODO implement OrderByTime
	//selectSpec.OrderByTime = true

	// Merge all series into a single group if we have no specific grouping dimensions.
	selectSpec.MergeAll = len(s.By) == 0 && len(s.Except) == 0
	selectSpec.GroupKeys = s.By
	selectSpec.GroupExcept = s.Except
	selectSpec.GroupKeep = s.Keep
}

func createGroupTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*GroupProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache(ctx.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewGroupTransformation(d, cache, s)
	return t, d, nil
}

type groupTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	keys   []string
	except []string
	keep   []string

	// Ignoring is true of len(keys) == 0 && len(except) > 0
	ignoring bool
}

func NewGroupTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *GroupProcedureSpec) *groupTransformation {
	t := &groupTransformation{
		d:        d,
		cache:    cache,
		keys:     spec.By,
		except:   spec.Except,
		keep:     spec.Keep,
		ignoring: len(spec.By) == 0 && len(spec.Except) > 0,
	}
	sort.Strings(t.keys)
	sort.Strings(t.except)
	sort.Strings(t.keep)
	return t
}

func (t *groupTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) (err error) {
	//TODO(nathanielc): Investigate if this can be smarter and not retract all blocks with the same time bounds.
	t.cache.ForEachBuilder(func(bk execute.BlockKey, builder execute.BlockBuilder) {
		if err != nil {
			return
		}
		if meta.Bounds().Equal(builder.Bounds()) {
			err = t.d.RetractBlock(bk)
		}
	})
	return
}

func (t *groupTransformation) Process(id execute.DatasetID, b execute.Block) error {
	isFanIn := false
	var tags execute.Tags
	if t.ignoring {
		// Assume we can fan in, we check for the false condition below
		isFanIn = true
		blockTags := b.Tags()
		tags = make(execute.Tags, len(blockTags))
		cols := b.Cols()
		for _, c := range cols {
			if c.IsTag {
				found := false
				for _, tag := range t.except {
					if tag == c.Label {
						found = true
						break
					}
				}
				if !found {
					if !c.IsCommon {
						isFanIn = false
						break
					}
					tags[c.Label] = blockTags[c.Label]
				}
			}
		}
	} else {
		tags, isFanIn = b.Tags().Subset(t.keys)
	}
	if isFanIn {
		return t.processFanIn(b, tags)
	} else {
		return t.processFanOut(b)
	}
}

// processFanIn assumes that all rows of b will be placed in the same builder.
func (t *groupTransformation) processFanIn(b execute.Block, tags execute.Tags) error {
	builder, new := t.cache.BlockBuilder(blockMetadata{
		tags:   tags,
		bounds: b.Bounds(),
	})
	if new {
		// Determine columns of new block.

		// Add existing columns, skipping tags.
		for _, c := range b.Cols() {
			if !c.IsTag {
				builder.AddCol(c)
			}
		}

		// Add tags.
		execute.AddTags(tags, builder)

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
	return nil
}

type tagMeta struct {
	idx      int
	isCommon bool
}

// processFanOut assumes each row of b could end up in a different builder.
func (t *groupTransformation) processFanOut(b execute.Block) error {
	cols := b.Cols()
	tagMap := make(map[string]tagMeta, len(cols))
	for j, c := range cols {
		if c.IsTag {
			ignoreTag := false
			for _, tag := range t.except {
				if tag == c.Label {
					ignoreTag = true
					break
				}
			}
			byTag := false
			for _, tag := range t.keys {
				if tag == c.Label {
					byTag = true
					break
				}
			}
			keepTag := false
			for _, tag := range t.keep {
				if tag == c.Label {
					keepTag = true
					break
				}
			}
			if (t.ignoring && !ignoreTag) || byTag || keepTag {
				tagMap[c.Label] = tagMeta{
					idx:      j,
					isCommon: (t.ignoring && !keepTag) || (!t.ignoring && byTag),
				}
			}
		}
	}

	// Iterate over each row and append to specific builder
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			tags := t.determineRowTags(tagMap, i, rr)
			builder, new := t.cache.BlockBuilder(blockMetadata{
				tags:   tags,
				bounds: b.Bounds(),
			})
			if new {
				// Add existing columns, skipping tags.
				for _, c := range cols {
					if !c.IsTag {
						builder.AddCol(c)
						continue
					}
					if meta, ok := tagMap[c.Label]; ok {
						j := builder.AddCol(execute.ColMeta{
							Label:    c.Label,
							Type:     execute.TString,
							IsTag:    true,
							IsCommon: meta.isCommon,
						})
						if meta.isCommon {
							builder.SetCommonString(j, tags[c.Label])
						}
					}
				}
			}
			// Construct map of builder column index to block column index.
			builderCols := builder.Cols()
			colMap := make([]int, len(builderCols))
			for j, c := range builderCols {
				for nj, nc := range cols {
					if c.Label == nc.Label {
						colMap[j] = nj
						break
					}
				}
			}

			// Add row to builder
			execute.AppendRow(i, rr, builder, colMap)
		}
	})
	return nil
}

func (t *groupTransformation) determineRowTags(tagMap map[string]tagMeta, i int, rr execute.RowReader) execute.Tags {
	cols := rr.Cols()
	tags := make(execute.Tags, len(cols))
	for t, meta := range tagMap {
		if meta.isCommon {
			tags[t] = rr.AtString(i, meta.idx)
		}
	}
	return tags
}

func (t *groupTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *groupTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
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
