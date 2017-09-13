package execute_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/ifql/query/plan"
)

var allowUnexportedDataFrame = cmp.AllowUnexported(blockList{}, block{})

var epoch = time.Unix(0, 0)

func TestExecutor_Execute(t *testing.T) {
	testCases := []struct {
		src  []block
		plan *plan.PlanSpec
		exp  []blockList
	}{
		{
			src: []block{
				{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					cells: []execute.Cell{
						{Value: 1, Time: 0},
						{Value: 2, Time: 1},
						{Value: 3, Time: 2},
						{Value: 4, Time: 3},
						{Value: 5, Time: 4},
					},
				},
			},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -5},
							},
						},
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &plan.SumProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source: plan.ProcedureIDFromOperationID("select"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
						Source: plan.ProcedureIDFromOperationID("sum"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
				},
			},
			exp: []blockList{{
				blocks: []block{{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					cells: []execute.Cell{
						{Value: 15, Time: 5, Tags: execute.Tags{}},
					},
				}},
			}},
		},
		{
			src: []block{
				{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					cells: []execute.Cell{
						{Value: 1, Time: 0},
						{Value: 2, Time: 1},
						{Value: 3, Time: 2},
						{Value: 4, Time: 3},
						{Value: 5, Time: 4},
					},
				},
			},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -5},
							},
						},
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &plan.SumProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID:   plan.ProcedureIDFromOperationID("join"),
						Spec: &plan.MergeJoinProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("join")),
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source: plan.ProcedureIDFromOperationID("select"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("sum")),
						Source: plan.ProcedureIDFromOperationID("sum"),
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						Source: plan.ProcedureIDFromOperationID("count"),
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("join")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("join")),
						Source: plan.ProcedureIDFromOperationID("join"),
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("join")),
				},
			},
			exp: []blockList{{
				blocks: []block{{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					cells: []execute.Cell{
						{Value: 3, Time: 5, Tags: execute.Tags{}},
					},
				}},
			}},
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			exe := execute.NewExecutor(&storageReader{
				blocks: tc.src,
			})
			results, err := exe.Execute(context.Background(), tc.plan)
			if err != nil {
				t.Fatal(err)
			}
			got := make([]blockList, len(results))
			for i, r := range results {
				got[i] = convertToBlockList(r)
			}

			if !cmp.Equal(got, tc.exp, allowUnexportedDataFrame) {
				t.Error("unexpected results", cmp.Diff(got, tc.exp, allowUnexportedDataFrame))
			}
		})
	}
}

type storageReader struct {
	blocks []block
}

func (s storageReader) Close() {}
func (s storageReader) Read(string, *storage.Predicate, int64, bool, execute.Time, execute.Time) (execute.BlockIterator, error) {
	return &storageBlockIterator{
		s: s,
	}, nil
}

type storageBlockIterator struct {
	s storageReader
}

func (bi *storageBlockIterator) Do(f func(execute.Block)) {
	for _, b := range bi.s.blocks {
		f(b)
	}
}

type blockList struct {
	blocks []block
	bounds execute.Bounds
}

func (df blockList) Bounds() execute.Bounds {
	return df.bounds
}

func (df blockList) Blocks() execute.BlockIterator {
	return &blockIterator{df.blocks}
}

func convertToBlockList(r execute.Result) blockList {
	bl := blockList{}
	blocks := r.Blocks()
	blocks.Do(func(b execute.Block) {
		bl.blocks = append(bl.blocks, convertToTestBlock(b))
	})
	return bl
}

type blockIterator struct {
	blocks []block
}

func (bi *blockIterator) Do(f func(execute.Block)) {
	for _, b := range bi.blocks {
		f(b)
	}
}

type block struct {
	bounds execute.Bounds
	tags   execute.Tags
	cells  []execute.Cell
}

func (b block) Bounds() execute.Bounds {
	return b.bounds
}

func (b block) Tags() execute.Tags {
	return b.tags
}

func (b block) Values() execute.ValueIterator {
	return &valueIterator{b.cells}
}

func (b block) Cells() execute.CellIterator {
	return &cellIterator{b.cells}
}

type valueIterator struct {
	cells []execute.Cell
}

func (vi *valueIterator) Do(f func([]float64)) {
	for _, c := range vi.cells {
		f([]float64{c.Value})
	}
}

type cellIterator struct {
	cells []execute.Cell
}

func (ci *cellIterator) Do(f func([]execute.Cell)) {
	f(ci.cells)
}

func convertToTestBlock(b execute.Block) block {
	blk := block{
		bounds: b.Bounds(),
		tags:   b.Tags(),
	}
	cells := b.Cells()

	cells.Do(func(cs []execute.Cell) {
		blk.cells = append(blk.cells, cs...)
	})
	return blk
}
