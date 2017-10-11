package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/functions"
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
	executetest.AggregateProcessTestHelper(t, new(functions.SumAgg), 10, 45)
}
