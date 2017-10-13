package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
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
		10,
	)
}
func BenchmarkCount(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		new(functions.CountAgg),
		NormalData,
		float64(len(NormalData)),
	)
}
