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
				Procedures: []*plan.Procedure{
					{
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					{
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
					},
					{
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source:       plan.ProcedureIDFromOperationID("select"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					{
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
						Source:       plan.ProcedureIDFromOperationID("range"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						Source: plan.ProcedureIDFromOperationID("count"),
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
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")): {
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source:       plan.ProcedureIDFromOperationID("select"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						Source: plan.ProcedureIDFromOperationID("count"),
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
				},
			},
		},
		{
			ap: &plan.AbstractPlanSpec{
				Procedures: []*plan.Procedure{
					{
						ID: plan.ProcedureIDFromOperationID("select"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					{
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
					},
					{
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &plan.LimitProcedureSpec{
							Limit: 10,
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit")),
					},
					{
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source:       plan.ProcedureIDFromOperationID("select"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					{
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range")),
						Source:       plan.ProcedureIDFromOperationID("range"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					{
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit")),
						Source:       plan.ProcedureIDFromOperationID("limit"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						Source: plan.ProcedureIDFromOperationID("count"),
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
						Parents: nil,
						Child:   plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						},
						Child: plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")): {
						ID:           plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select")),
						Source:       plan.ProcedureIDFromOperationID("select"),
						Destinations: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
						Source: plan.ProcedureIDFromOperationID("count"),
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count")),
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
