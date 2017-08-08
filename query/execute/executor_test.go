package execute_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

var ignoreUnexportedDataFrame = cmpopts.IgnoreUnexported(dataFrame{})

var epoch = time.Unix(0, 0)

func TestExecutor_Execute(t *testing.T) {
	testCases := []struct {
		src  []dataFrame
		plan *plan.PlanSpec
		exp  []dataFrame
	}{
		{
			src: []dataFrame{{
				Data: []float64{1, 2, 3, 4, 5},
				Rows: []execute.Tags{
					{},
				},
				Cols: []execute.Time{
					1, 2, 3, 4, 5,
				},
				bounds: bounds{
					start: 1,
					stop:  5,
				},
			}},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Operations: map[plan.OperationID]*plan.Operation{
					plan.OpIDFromQueryOpID("0"): {
						ID: plan.OpIDFromQueryOpID("0"),
						Spec: &plan.SelectOpSpec{
							Database: "mydb",
						},
						Parents: nil,
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						},
					},
					plan.OpIDFromQueryOpID("1"): {
						ID: plan.OpIDFromQueryOpID("1"),
						Spec: &plan.RangeOpSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -5},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						},
					},
					plan.OpIDFromQueryOpID("2"): {
						ID:   plan.OpIDFromQueryOpID("2"),
						Spec: &plan.SumOpSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")),
						},
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")): {
						ID:     plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						Source: plan.OpIDFromQueryOpID("0"),
					},
					plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")): {
						ID:     plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						Source: plan.OpIDFromQueryOpID("1"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")): {
						ID:     plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")),
						Source: plan.OpIDFromQueryOpID("2"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")),
				},
			},
			exp: []dataFrame{{
				Data: []float64{15},
				Rows: []execute.Tags{
					{},
				},
				Cols: []execute.Time{
					5,
				},
				bounds: bounds{
					start: 1,
					stop:  5,
				},
			}},
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			exe := execute.NewExecutor(&storageReader{
				frames: tc.src,
			})
			results, err := exe.Execute(context.Background(), tc.plan)
			if err != nil {
				t.Fatal(err)
			}
			got := make([]dataFrame, len(results))
			for i, r := range results {
				got[i] = convertToLocalDataFrame(r)
			}

			if !cmp.Equal(got, tc.exp, ignoreUnexportedDataFrame) {
				t.Error("unexpected results", cmp.Diff(got, tc.exp))
			}
		})
	}
}

type storageReader struct {
	frames []dataFrame
	idx    int
}

func (s *storageReader) Read() (execute.DataFrame, bool) {
	idx := s.idx
	if idx >= len(s.frames) {
		return nil, false
	}
	s.idx++
	return s.frames[idx], true
}

type dataFrame struct {
	Data   []float64
	Rows   []execute.Tags
	Cols   []execute.Time
	bounds bounds
}

type bounds struct {
	start execute.Time
	stop  execute.Time
}

func (b bounds) Start() execute.Time {
	return b.start
}
func (b bounds) Stop() execute.Time {
	return b.stop
}

func (df dataFrame) Bounds() execute.Bounds {
	return df.bounds
}

func (df dataFrame) NRows() int {
	return len(df.Rows)
}

func (df dataFrame) NCols() int {
	return len(df.Cols)
}
func (df dataFrame) ColsIndex() []execute.Time {
	return df.Cols
}

func (df dataFrame) RowSlice(i int) ([]float64, execute.Tags) {
	s := len(df.Cols)
	return df.Data[i*s : (i+1)*s], df.Rows[i]
}

func convertToLocalDataFrame(d execute.DataFrame) dataFrame {
	df := dataFrame{
		Cols: d.ColsIndex(),
	}
	r := d.NRows()
	for i := 0; i < r; i++ {
		row, tags := d.RowSlice(i)
		df.Data = append(df.Data, row...)
		df.Rows = append(df.Rows, tags)
	}
	return df
}
