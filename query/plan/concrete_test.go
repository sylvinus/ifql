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
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
					},
					{
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						},
					},
					{
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						},
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						Source: plan.ProcedureIDFromOperationID("select"),
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						Source: plan.ProcedureIDFromOperationID("range"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						Source: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
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
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						},
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"): {
						ID:          plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						Source:      plan.ProcedureIDFromOperationID("select"),
						Destination: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						Source: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
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
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
					},
					{
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						},
					},
					{
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &plan.LimitProcedureSpec{
							Limit: 10,
						},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit"), "0"),
						},
					},
					{
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						},
					},
				},
				Datasets: []*plan.Dataset{
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						Source: plan.ProcedureIDFromOperationID("select"),
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("range"), "0"),
						Source: plan.ProcedureIDFromOperationID("range"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("limit"), "0"),
						Source: plan.ProcedureIDFromOperationID("limit"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					{
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						Source: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
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
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						},
						Children: []plan.DatasetID{
							plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						},
					},
				},
				Datasets: map[plan.DatasetID]*plan.Dataset{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"): {
						ID:          plan.CreateDatasetID(plan.ProcedureIDFromOperationID("select"), "0"),
						Source:      plan.ProcedureIDFromOperationID("select"),
						Destination: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"): {
						ID:     plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
						Source: plan.ProcedureIDFromOperationID("count"),
						Bounds: plan.BoundsSpec{
							Start: query.Time{Relative: -1 * time.Hour},
						},
					},
				},
				Results: []plan.DatasetID{
					plan.CreateDatasetID(plan.ProcedureIDFromOperationID("count"), "0"),
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
