package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/querytest"
)

func TestLimitOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"limit","kind":"limit","spec":{"limit":10,"offset":5}}`)
	op := &query.Operation{
		ID: "limit",
		Spec: &functions.LimitOpSpec{
			Limit:  10,
			Offset: 5,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}
