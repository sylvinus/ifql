package execute_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

var allowUnexported = cmp.AllowUnexported(blockList{}, block{})

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
					points: []point{
						{Value: 1.0, Time: 0},
						{Value: 2.0, Time: 1},
						{Value: 3.0, Time: 2},
						{Value: 4.0, Time: 3},
						{Value: 5.0, Time: 4},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ColMeta{
							Label: execute.ValueColLabel,
							Type:  execute.TFloat,
						},
					},
				},
			},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Bounds: plan.BoundsSpec{
					Start: query.Time{Absolute: time.Unix(0, 1)},
					Stop:  query.Time{Absolute: time.Unix(0, 5)},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									Relative:   -5,
									IsRelative: true,
								},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{}, Parents: []plan.ProcedureID{
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
					points: []point{
						{Value: 15.0, Time: 5},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ColMeta{
							Label: execute.ValueColLabel,
							Type:  execute.TFloat,
						},
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
				points: []point{
					{Value: int64(1), Time: 0},
					{Value: int64(2), Time: 1},
					{Value: int64(3), Time: 2},
					{Value: int64(4), Time: 3},
					{Value: int64(5), Time: 4},
				},
				cols: []execute.ColMeta{
					execute.TimeCol,
					execute.ColMeta{
						Label: execute.ValueColLabel,
						Type:  execute.TInt,
					},
				},
			}},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Bounds: plan.BoundsSpec{
					Start: query.Time{Absolute: time.Unix(0, 1)},
					Stop:  query.Time{Absolute: time.Unix(0, 5)},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									Relative:   -5,
									IsRelative: true,
								},
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
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
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
					points: []point{
						{Value: int64(3), Time: 5},
					},
					cols: []execute.ColMeta{
						execute.TimeCol,
						execute.ColMeta{
							Label: execute.ValueColLabel,
							Type:  execute.TInt,
						},
					},
				}},
			}},
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := execute.Config{
				StorageReader: &storageReader{blocks: tc.src},
			}
			exe := execute.NewExecutor(c)
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

func (bi *storageBlockIterator) Do(f func(execute.Block)) error {
	for _, b := range bi.s.blocks {
		f(b)
	}
	return nil
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

func (bi *blockIterator) Do(f func(execute.Block)) error {
	for _, b := range bi.blocks {
		f(b)
	}
	return nil
}

type block struct {
	bounds execute.Bounds
	tags   execute.Tags
	points []point
	cols   []execute.ColMeta
}

type point struct {
	Value interface{}
	Time  execute.Time
	Tags  execute.Tags
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
	return &valueIterator{points: b.points, cols: b.cols}
}
func (b block) Values() execute.ValueIterator {
	return &valueIterator{points: b.points, cols: b.cols}
}
func (b block) Times() execute.ValueIterator {
	return &valueIterator{points: b.points, cols: b.cols}
}

type valueIterator struct {
	points []point
	cols   []execute.ColMeta
}

func (itr *valueIterator) Cols() []execute.ColMeta {
	return itr.cols
}
func (itr *valueIterator) DoBool(f func([]bool, execute.RowReader)) {
	for _, p := range itr.points {
		f([]bool{p.Value.(bool)}, itr)
	}
}
func (itr *valueIterator) DoInt(f func([]int64, execute.RowReader)) {
	for _, p := range itr.points {
		f([]int64{p.Value.(int64)}, itr)
	}
}
func (itr *valueIterator) DoUInt(f func([]uint64, execute.RowReader)) {
	for _, p := range itr.points {
		f([]uint64{p.Value.(uint64)}, itr)
	}
}
func (itr *valueIterator) DoFloat(f func([]float64, execute.RowReader)) {
	for _, p := range itr.points {
		f([]float64{p.Value.(float64)}, itr)
	}
}
func (itr *valueIterator) DoString(f func([]string, execute.RowReader)) {
}
func (itr *valueIterator) DoTime(f func([]execute.Time, execute.RowReader)) {
	for _, p := range itr.points {
		f([]execute.Time{p.Time}, itr)
	}
}
func (itr *valueIterator) AtBool(i, j int) bool {
	return itr.points[i].Value.(bool)
}
func (itr *valueIterator) AtInt(i, j int) int64 {
	return itr.points[i].Value.(int64)
}
func (itr *valueIterator) AtUInt(i, j int) uint64 {
	return itr.points[i].Value.(uint64)
}
func (itr *valueIterator) AtFloat(i, j int) float64 {
	return itr.points[i].Value.(float64)
}
func (itr *valueIterator) AtString(i, j int) string {
	return itr.points[i].Tags[itr.cols[j].Label]
}
func (itr *valueIterator) AtTime(i, j int) execute.Time {
	return itr.points[i].Time
}

func convertToTestBlock(b execute.Block) block {
	cols := b.Cols()
	blk := block{
		bounds: b.Bounds(),
		tags:   b.Tags(),
		cols:   cols,
	}
	valueIdx := execute.ValueIdx(cols)
	valueType := cols[valueIdx].Type
	times := b.Times()
	times.DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i, time := range ts {
			var v interface{}
			switch valueType {
			case execute.TBool:
				v = rr.AtBool(i, valueIdx)
			case execute.TInt:
				v = rr.AtInt(i, valueIdx)
			case execute.TUInt:
				v = rr.AtUInt(i, valueIdx)
			case execute.TFloat:
				v = rr.AtFloat(i, valueIdx)
			case execute.TString:
				v = rr.AtString(i, valueIdx)
			}
			tags := execute.TagsForRow(i, rr)
			blk.points = append(blk.points, point{
				Time:  time,
				Value: v,
				Tags:  tags,
			})
		}
	})
	return blk
}
