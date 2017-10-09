package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
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
