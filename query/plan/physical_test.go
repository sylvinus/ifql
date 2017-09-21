package plan_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

func TestPhysicalPlanner_Plan(t *testing.T) {
	testCases := []struct {
		lp *plan.LogicalPlanSpec
		cp *plan.PlanSpec
	}{
		{
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
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
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: nil,
					},
				},
			},
			cp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
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
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("select")),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("count")),
				},
			},
		},
		{
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
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
						Spec: &plan.LimitProcedureSpec{
							Limit: 10,
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
			},
			cp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select"): {
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
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
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("select")),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("count")),
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
			if !cmp.Equal(got, tc.cp) {
				t.Errorf("unexpected physical plan:\n%s", cmp.Diff(got, tc.cp))
			}
		})
	}
}
