package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFromOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"from","kind":"from","spec":{"database":"mydb"}}`)
	op := &query.Operation{
		ID: "from",
		Spec: &functions.FromOpSpec{
			Database: "mydb",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
