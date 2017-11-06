package executetest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/query/execute"
)

// AggFuncTestHelper splits the data in half, runs Do over each split and compares
// the Value to want.
func AggFuncTestHelper(t *testing.T, aggF execute.AggFunc, data []float64, want float64) {
	t.Helper()

	// Call Do twice, since this is possible according to the interface.
	h := len(data) / 2
	aggF.Do(data[:h])
	if h < len(data) {
		aggF.Do(data[h:])
	}

	got := aggF.Value()

	if !cmp.Equal(want, got, cmpopts.EquateNaNs()) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
	}
}

const small = 1e-5

// AggFuncBenchmarkHelper benchmarks the aggregate function over data and compares to wantValue
func AggFuncBenchmarkHelper(b *testing.B, aggF execute.AggFunc, data []float64, wantValue float64) {
	b.Helper()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		aggF.Reset()
		aggF.Do(data)
		v := aggF.Value()
		if diff := v - wantValue; diff > small || diff < -small {
			b.Fatalf("unexpected result: got: %f want: %f", v, wantValue)
		}
	}
}
