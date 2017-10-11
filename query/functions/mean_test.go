package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestMeanOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"mean","kind":"mean"}`)
	op := &query.Operation{
		ID:   "mean",
		Spec: &functions.MeanOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMean_Process(t *testing.T) {
	executetest.AggregateProcessTestHelper(t,
		new(functions.MeanAgg),
		[]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		4.5,
	)
}
