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
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("0"): {
						ID: plan.ProcedureIDFromOperationID("0"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("1")},
					},
					plan.ProcedureIDFromOperationID("1"): {
						ID: plan.ProcedureIDFromOperationID("1"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("2")},
					},
					plan.ProcedureIDFromOperationID("2"): {
						ID:   plan.ProcedureIDFromOperationID("2"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("1"),
						},
						Children: nil,
					},
				},
			},
		},
		{
			q: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range0",
						Spec: &query.RangeOpSpec{
							Start: query.Time{Relative: -1 * time.Hour},
							Stop:  query.Time{},
						},
					},
					{
						ID:   "count0",
						Spec: &query.CountOpSpec{},
					},
					{
						ID: "select1",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &query.RangeOpSpec{
							Start: query.Time{Relative: -1 * time.Hour},
							Stop:  query.Time{},
						},
					},
					{
						ID:   "sum1",
						Spec: &query.SumOpSpec{},
					},
					{
						ID:   "join",
						Spec: &query.JoinOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "range0"},
					{Parent: "range0", Child: "count0"},
					{Parent: "select1", Child: "range1"},
					{Parent: "range1", Child: "sum1"},
					{Parent: "count0", Child: "join"},
					{Parent: "sum1", Child: "join"},
				},
			},
			ap: &plan.AbstractPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select0"): {
						ID: plan.ProcedureIDFromOperationID("select0"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range0")},
					},
					plan.ProcedureIDFromOperationID("range0"): {
						ID: plan.ProcedureIDFromOperationID("range0"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count0")},
					},
					plan.ProcedureIDFromOperationID("count0"): {
						ID:   plan.ProcedureIDFromOperationID("count0"),
						Spec: &plan.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("select1"): {
						ID: plan.ProcedureIDFromOperationID("select1"),
						Spec: &plan.SelectProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range1")},
					},
					plan.ProcedureIDFromOperationID("range1"): {
						ID: plan.ProcedureIDFromOperationID("range1"),
						Spec: &plan.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{Relative: -1 * time.Hour},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select1"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum1")},
					},
					plan.ProcedureIDFromOperationID("sum1"): {
						ID:   plan.ProcedureIDFromOperationID("sum1"),
						Spec: &plan.SumProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range1"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID:   plan.ProcedureIDFromOperationID("join"),
						Spec: &plan.MergeJoinProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("count0"),
							plan.ProcedureIDFromOperationID("sum1"),
						},
						Children: nil,
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
