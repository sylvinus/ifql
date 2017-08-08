package plan_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

func TestConcretePlanner_Plan(t *testing.T) {
	testCases := []struct {
		ap *plan.AbstractPlanSpec
		cp *plan.PlanSpec
	}{
		{
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
			cp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
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
					plan.OpIDFromQueryOpID("2"): {
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
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			planner := plan.NewPlanner()
			got, err := planner.Plan(tc.ap, nil, time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC))
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tc.cp) {
				t.Errorf("unexpected concrete plan:\n%s", cmp.Diff(got, tc.cp))
			}
		})
	}
}
