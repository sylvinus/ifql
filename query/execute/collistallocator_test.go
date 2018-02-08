package execute_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
)

func assertEqual(t *testing.T, exp, got interface{}) {
	if !cmp.Equal(exp, got) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(exp, got))
	}
}

func TestColListAllocatorBools(t *testing.T) {
	var assert executetest.Assert

	res := &execute.Resource{Limit: 1000}
	ca := execute.NewColListAllocator(execute.NewLimitedAllocator(executetest.UnlimitedAllocator, res))
	b := ca.Bools(5, 5)
	exp := 5
	got := len(b)
	assert.Equal(t, exp, got)
	assert.Equal(t, int64(5), res.Reserved())

	b[0] = true
	b[4] = true

	b = ca.AppendBools(b, false, false, true, true)
	assert.Equal(t, []bool{true, false, false, false, true, false, false, true, true}, b)
	assert.Equal(t, int64(9), res.Reserved())

	ca.FreeBools(b)
	assert.Equal(t, int64(0), res.Reserved())
}
