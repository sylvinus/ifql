package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestSumOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"sum","kind":"sum"}`)
	op := &query.Operation{
		ID:   "sum",
		Spec: &functions.SumOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSum_Process(t *testing.T) {
	executetest.AggFuncTestHelper(t,
		new(functions.SumAgg),
		[]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		45,
	)
}

func BenchmarkSum(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		new(functions.SumAgg),
		NormalData,
		10000816.9673,
	)
}
