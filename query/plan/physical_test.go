package plan_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

func TestPhysicalPlanner_Plan(t *testing.T) {
	testCases := []struct {
		lp *plan.LogicalPlanSpec
		pp *plan.PlanSpec
	}{
		{
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select"),
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
					plan.ProcedureIDFromOperationID("select"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Bounds: plan.BoundsSpec{
					Start: query.Time{Relative: -1 * time.Hour},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("select")),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("count")),
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("select"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
		},
		{
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("select")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					plan.ProcedureIDFromOperationID("limit"): {
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &functions.LimitProcedureSpec{
							Limit: 10,
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("select"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("limit"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Bounds: plan.BoundsSpec{
					Start: query.Time{Relative: -1 * time.Hour},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &functions.SelectProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
							LimitSet: true,
							Limit:    10,
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("select")),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("count")),
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("select"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			planner := plan.NewPlanner()
			got, err := planner.Plan(tc.lp, nil, time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC))
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tc.pp) {
				t.Errorf("unexpected physical plan -want/+got %s", cmp.Diff(tc.pp, got))
			}
		})
	}
}

var benchmarkPhysicalPlan *plan.PlanSpec

func BenchmarkPhysicalPlan(b *testing.B) {
	var err error
	lp, err := plan.NewLogicalPlanner().Plan(benchmarkQuery)
	if err != nil {
		b.Fatal(err)
	}
	planner := plan.NewPlanner()
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	for n := 0; n < b.N; n++ {
		benchmarkPhysicalPlan, err = planner.Plan(lp, nil, now)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var benchmarkQueryToPhysicalPlan *plan.PlanSpec

func BenchmarkQueryToPhysicalPlan(b *testing.B) {
	lp := plan.NewLogicalPlanner()
	pp := plan.NewPlanner()
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	for n := 0; n < b.N; n++ {
		lp, err := lp.Plan(benchmarkQuery)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkQueryToPhysicalPlan, err = pp.Plan(lp, nil, now)
		if err != nil {
			b.Fatal(err)
		}
	}
}
