package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestGroupOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"group","kind":"group","spec":{"keys":["t1","t2"],"keep":["t3","t4"]}}`)
	op := &query.Operation{
		ID: "group",
		Spec: &functions.GroupOpSpec{
			Keys: []string{"t1", "t2"},
			Keep: []string{"t3", "t4"},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
