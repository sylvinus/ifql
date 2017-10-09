package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestSelectOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"select","kind":"select","spec":{"database":"mydb"}}`)
	op := &query.Operation{
		ID: "select",
		Spec: &functions.SelectOpSpec{
			Database: "mydb",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
