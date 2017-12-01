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

func TestCount_PushDown_Match(t *testing.T) {
	spec := new(functions.CountProcedureSpec)
	from := new(functions.FromProcedureSpec)

	// Should not match when an aggregate is set
	from.GroupingSet = true
	plantest.PhysicalPlan_PushDown_Match_TestHelper(t, spec, from, false)

	// Should match when no aggregate is set
	from.GroupingSet = false
	plantest.PhysicalPlan_PushDown_Match_TestHelper(t, spec, from, true)
}

func TestCount_PushDown(t *testing.T) {
	spec := new(functions.CountProcedureSpec)
	root := &plan.Procedure{
		Spec: new(functions.FromProcedureSpec),
	}
	want := &plan.Procedure{
		Spec: &functions.FromProcedureSpec{
			AggregateSet:  true,
			AggregateType: functions.CountKind,
		},
	}

	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, false, want)
}

func TestCount_PushDown_Duplicate(t *testing.T) {
	spec := new(functions.CountProcedureSpec)
	root := &plan.Procedure{
		Spec: &functions.FromProcedureSpec{
			AggregateSet:  true,
			AggregateType: functions.CountKind,
		},
	}
	want := &plan.Procedure{
		// Expect the duplicate has been reset to zero values
		Spec: new(functions.FromProcedureSpec),
	}

	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, true, want)
}
