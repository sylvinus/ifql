package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/ifql/ifqltest"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestCount_NewQuery(t *testing.T) {
	tests := []ifqltest.NewQueryTestCase{
		{
			Name: "from with range and count",
			Raw:  `from(db:"mydb").range(start:-4h, stop:-2h).count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count2",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "count2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ifqltest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestCountOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"count","kind":"count"}`)
	op := &query.Operation{
		ID:   "count",
		Spec: &functions.CountOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestCount_Process(t *testing.T) {
	executetest.AggFuncTestHelper(
		t,
		new(functions.CountAgg),
		[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		int64(10),
	)
}
func BenchmarkCount(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		new(functions.CountAgg),
		NormalData,
		int64(len(NormalData)),
	)
}

func TestCount_PushDown_Single(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
			},
			plan.ProcedureIDFromOperationID("count"): {
				ID:   plan.ProcedureIDFromOperationID("count"),
				Spec: &functions.CountProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("count"),
		},
	}

	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					AggregateSet:  true,
					AggregateType: "count",
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestCount_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("range"),
				},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("sum"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			plan.ProcedureIDFromOperationID("sum"): {
				ID:   plan.ProcedureIDFromOperationID("sum"),
				Spec: &functions.SumProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("count"): {
				ID:   plan.ProcedureIDFromOperationID("count"),
				Spec: &functions.CountProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("sum"),
			plan.ProcedureIDFromOperationID("count"), // Count is last so it will get duplicated
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					AggregateSet:  true,
					AggregateType: "sum",
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					AggregateSet:  true,
					AggregateType: "count",
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}
