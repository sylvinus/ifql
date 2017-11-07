package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

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
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
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
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("count"),
		},
	}

	want := &plan.PlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database:      "mydb",
					AggregateSet:  true,
					AggregateType: "count",
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			(plan.ProcedureIDFromOperationID("select")),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestCount_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("sum"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			plan.ProcedureIDFromOperationID("sum"): {
				ID:   plan.ProcedureIDFromOperationID("sum"),
				Spec: &functions.SumProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("select")),
				},
				Children: nil,
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
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("sum"),
			plan.ProcedureIDFromOperationID("count"), // Count is last so it will get duplicated
		},
	}

	selectID := plan.ProcedureIDFromOperationID("select")
	selectIDDup := plan.ProcedureIDForDuplicate(selectID)
	want := &plan.PlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			selectID: {
				ID: selectID,
				Spec: &functions.SelectProcedureSpec{
					Database:      "mydb",
					AggregateSet:  true,
					AggregateType: "sum",
				},
				Children: []plan.ProcedureID{},
			},
			selectIDDup: {
				ID: selectIDDup,
				Spec: &functions.SelectProcedureSpec{
					Database:      "mydb",
					AggregateSet:  true,
					AggregateType: "count",
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			selectID,
			selectIDDup,
		},
		Order: []plan.ProcedureID{
			selectID,
			selectIDDup,
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}
