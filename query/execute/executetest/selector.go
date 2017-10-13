package executetest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query/execute"
)

func SelectorFuncTestHelper(t *testing.T, selectorF execute.SelectorFunc, data execute.Block, want []execute.Row) {
	t.Helper()

	data.Values().DoFloat(selectorF.Do)

	got := selectorF.Rows()
	selectorF.Reset()

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
	}
}

var rows []execute.Row

func SelectorFuncBenchmarkHelper(b *testing.B, selectorF execute.SelectorFunc, data execute.Block) {
	b.Helper()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		selectorF.Reset()
		data.Values().DoFloat(selectorF.Do)
		rows = selectorF.Rows()
	}
}
