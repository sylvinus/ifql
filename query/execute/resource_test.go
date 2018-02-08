package execute_test

import (
	"testing"

	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
)

func TestResource_ReserveLimit(t *testing.T) {
	var assert executetest.Assert
	r := &execute.Resource{Limit: 1000}
	r.Reserve(1000, 1)
	assert.Equal(t, int64(1000), r.Reserved())
}

func TestResource_ReserveExceedLimit(t *testing.T) {
	var assert executetest.Assert
	r := &execute.Resource{Limit: 10}
	exp := execute.ResourceError{
		Limit:     10,
		Allocated: 0,
		Wanted:    1000,
	}
	assert.PanicsWithValue(t, exp, func() {
		r.Reserve(1000, 1)
	})
}
