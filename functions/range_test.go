package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/ifql/ifqltest"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestRange_NewQuery(t *testing.T) {
	tests := []ifqltest.NewQueryTestCase{
		{
			Name: "from with database with range",
			Raw:  `from(db:"mydb").range(start:-4h, stop:-2h).sum()`,
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
						ID:   "sum2",
						Spec: &functions.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
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

func TestRangeOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"range","kind":"range","spec":{"start":"-1h","stop":"2017-10-10T00:00:00Z"}}`)
	op := &query.Operation{
		ID: "range",
		Spec: &functions.RangeOpSpec{
			Start: query.Time{
				Relative:   -1 * time.Hour,
				IsRelative: true,
			},
			Stop: query.Time{
				Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestRange_PushDown_Single(t *testing.T) {
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
						Start: query.Time{
							Relative:   -1 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("from")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
		},
	}

	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.Time{
				Relative:   -1 * time.Hour,
				IsRelative: true,
			},
			Stop: query.Time{
				Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
			},
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							Relative:   -1 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			(plan.ProcedureIDFromOperationID("from")),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestRange_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("rangeA"),
					plan.ProcedureIDFromOperationID("rangeB"),
				},
			},
			plan.ProcedureIDFromOperationID("rangeA"): {
				ID: plan.ProcedureIDFromOperationID("rangeA"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							Relative:   -1 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("from")),
				},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("rangeB"): {
				ID: plan.ProcedureIDFromOperationID("rangeB"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							Relative:   -10 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2007, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("from")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("rangeA"),
			plan.ProcedureIDFromOperationID("rangeB"), // rangeB is last so it will be duplicated
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.Time{
				Relative:   -10 * time.Hour,
				IsRelative: true,
			},
			Stop: query.Time{
				Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
			},
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							Relative:   -1 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							Relative:   -10 * time.Hour,
							IsRelative: true,
						},
						Stop: query.Time{
							Absolute: time.Date(2007, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
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
