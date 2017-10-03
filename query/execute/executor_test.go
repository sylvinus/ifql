package execute_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/plan"
)

var allowUnexported = cmp.AllowUnexported(blockList{}, block{}, execute.Row{})

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
					rows: []execute.Row{
						{Values: []interface{}{1.0, execute.Time(0)}},
						{Values: []interface{}{2.0, execute.Time(1)}},
						{Values: []interface{}{3.0, execute.Time(2)}},
						{Values: []interface{}{4.0, execute.Time(4)}},
						{Values: []interface{}{5.0, execute.Time(5)}},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ValueCol,
					},
				},
			},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -5},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select"),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("sum"),
				},
			},
			exp: []blockList{{
				blocks: []block{{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					rows: []execute.Row{
						{Values: []interface{}{15.0, execute.Time(5)}},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ValueCol,
					},
				}},
			}},
		},
		{
			src: []block{{
				bounds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				tags: execute.Tags{},
				rows: []execute.Row{
					{Values: []interface{}{1.0, execute.Time(0)}},
					{Values: []interface{}{2.0, execute.Time(1)}},
					{Values: []interface{}{3.0, execute.Time(2)}},
					{Values: []interface{}{4.0, execute.Time(3)}},
					{Values: []interface{}{5.0, execute.Time(4)}},
				},
				cols: []execute.ColMeta{
					execute.TimeCol,
					execute.ValueCol,
				},
			}},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -5},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID: plan.ProcedureIDFromOperationID("join"),
						Spec: &functions.MergeJoinProcedureSpec{
							Expression: &expression.BinaryNode{
								Operator: expression.DivisionOperator,
								Left: &expression.ReferenceNode{
									Name: "$",
									Kind: "field",
								},
								Right: &expression.ReferenceNode{
									Name: "b",
									Kind: "identifier",
								},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
							plan.ProcedureIDFromOperationID("count"),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("join"),
				},
			},
			exp: []blockList{{
				blocks: []block{{
					bounds: execute.Bounds{
						Start: 1,
						Stop:  5,
					},
					tags: execute.Tags{},
					rows: []execute.Row{
						{Values: []interface{}{3.0, execute.Time(5)}},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ValueCol,
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

			if !cmp.Equal(got, tc.exp, allowUnexported) {
				t.Error("unexpected results -want/+got", cmp.Diff(tc.exp, got, allowUnexported))
			}
		})
	}
}

type storageReader struct {
	blocks []block
}

func (s storageReader) Close() {}
func (s storageReader) Read(execute.ReadSpec, execute.Time, execute.Time) (execute.BlockIterator, error) {
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
	rows   []execute.Row
	cols   []execute.ColMeta
}

func (b block) Bounds() execute.Bounds {
	return b.bounds
}

func (b block) Tags() execute.Tags {
	return b.tags
}
func (b block) Cols() []execute.ColMeta {
	return b.cols
}

func (b block) Col(c int) execute.ValueIterator {
	return &valueIterator{b.rows}
}
func (b block) Values() execute.ValueIterator {
	return &valueIterator{b.rows}
}

func (b block) Rows() execute.RowIterator {
	return &rowIterator{b.rows}
}

type valueIterator struct {
	rows []execute.Row
}

func (vi *valueIterator) DoFloat(f func([]float64)) {
	for _, r := range vi.rows {
		f([]float64{r.Value()})
	}
}
func (vi *valueIterator) DoString(f func([]string)) {
}
func (vi *valueIterator) DoTime(f func([]execute.Time)) {
}

type rowIterator struct {
	rows []execute.Row
}

func (ci *rowIterator) Do(f func([]execute.Row)) {
	f(ci.rows)
}

func convertToTestBlock(b execute.Block) block {
	blk := block{
		bounds: b.Bounds(),
		tags:   b.Tags(),
		cols:   b.Cols(),
	}
	rows := b.Rows()

	rows.Do(func(rs []execute.Row) {
		blk.rows = append(blk.rows, rs...)
	})
	return blk
}
