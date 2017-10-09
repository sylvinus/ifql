package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestWindowOperation_Marshaling(t *testing.T) {
	//TODO: Test marshalling of triggerspec
	data := []byte(`{"id":"window","kind":"window","spec":{"every":"1m","period":"1h","start":"-4h","round":"1s"}}`)
	op := &query.Operation{
		ID: "window",
		Spec: &functions.WindowOpSpec{
			Every:  query.Duration(time.Minute),
			Period: query.Duration(time.Hour),
			Start: query.Time{
				Relative: -4 * time.Hour,
			},
			Round: query.Duration(time.Second),
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}
