package plan_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

func TestAbstractPlanner_Plan(t *testing.T) {
	testCases := []struct {
		q  *query.QuerySpec
		ap *plan.AbstractPlanSpec
	}{
		{
			q: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "0",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "1",
						Spec: &query.RangeOpSpec{
							Start: query.Time{Relative: -1 * time.Hour},
							Stop:  query.Time{},
						},
					},
					{
						ID:   "2",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "0", Child: "1"},
					{Parent: "1", Child: "2"},
				},
			},
			ap: &plan.AbstractPlanSpec{
				Procedures: []*plan.Procedure{
					{
						ID: plan.ProcedureIDFromOperationID("0"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents: nil,
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("0"), "0"),
						},
					},
					{
						ID: plan.ProcedureIDFromOperationID("1"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("0"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("1"), "0"),
						},
					},
					{
						ID:   plan.ProcedureIDFromOperationID("2"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("1"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("2"), "0"),
						},
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("0"), "0"),
						Source: plan.ProcedureIDFromOperationID("0"),
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("1"), "0"),
						Source: plan.ProcedureIDFromOperationID("1"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("2"), "0"),
						Source: plan.ProcedureIDFromOperationID("2"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			planner := plan.NewAbstractPlanner()
			got, err := planner.Plan(tc.q)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tc.ap) {
				t.Errorf("unexpected abstract plan:\n%s", cmp.Diff(got, tc.ap))
			}
		})
	}
}
