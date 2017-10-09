package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestStddevOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"stddev","kind":"stddev"}`)
	op := &query.Operation{
		ID:   "stddev",
		Spec: &functions.StddevOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}
