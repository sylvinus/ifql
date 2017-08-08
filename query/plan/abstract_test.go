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
				Operations: []*plan.Operation{
					{
						ID: plan.OpIDFromQueryOpID("0"),
						Spec: &plan.SelectOpSpec{
							Database: "mydb",
						},
						Parents: nil,
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						},
					},
					{
						ID: plan.OpIDFromQueryOpID("1"),
						Spec: &plan.RangeOpSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						},
					},
					{
						ID:   plan.OpIDFromQueryOpID("2"),
						Spec: &plan.CountOpSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")),
						},
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:     plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")),
						Source: plan.OpIDFromQueryOpID("0"),
					},
					{
						ID:     plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")),
						Source: plan.OpIDFromQueryOpID("1"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					{
						ID:     plan.CreateDatasetID(plan.CreateDatasetID(plan.CreateDatasetID(plan.InvalidDatasetID, plan.OpIDFromQueryOpID("0")), plan.OpIDFromQueryOpID("1")), plan.OpIDFromQueryOpID("2")),
						Source: plan.OpIDFromQueryOpID("2"),
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
