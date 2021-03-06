package executetest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/query/execute"
)

// AggFuncTestHelper splits the data in half, runs Do over each split and compares
// the Value to want.
func AggFuncTestHelper(t *testing.T, agg execute.Aggregate, data []float64, want interface{}) {
	t.Helper()

	// Call Do twice, since this is possible according to the interface.
	h := len(data) / 2
	vf := agg.NewFloatAgg()
	vf.DoFloat(data[:h])
	if h < len(data) {
		vf.DoFloat(data[h:])
	}

	var got interface{}
	switch vf.Type() {
	case execute.TBool:
		got = vf.(execute.BoolValueFunc).ValueBool()
	case execute.TInt:
		got = vf.(execute.IntValueFunc).ValueInt()
	case execute.TUInt:
		got = vf.(execute.UIntValueFunc).ValueUInt()
	case execute.TFloat:
		got = vf.(execute.FloatValueFunc).ValueFloat()
	case execute.TString:
		got = vf.(execute.StringValueFunc).ValueString()
	}

	if !cmp.Equal(want, got, cmpopts.EquateNaNs()) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
	}
}

// AggFuncBenchmarkHelper benchmarks the aggregate function over data and compares to wantValue
func AggFuncBenchmarkHelper(b *testing.B, agg execute.Aggregate, data []float64, want interface{}) {
	b.Helper()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		vf := agg.NewFloatAgg()
		vf.DoFloat(data)
		var got interface{}
		switch vf.Type() {
		case execute.TBool:
			got = vf.(execute.BoolValueFunc).ValueBool()
		case execute.TInt:
			got = vf.(execute.IntValueFunc).ValueInt()
		case execute.TUInt:
			got = vf.(execute.UIntValueFunc).ValueUInt()
		case execute.TFloat:
			got = vf.(execute.FloatValueFunc).ValueFloat()
		case execute.TString:
			got = vf.(execute.StringValueFunc).ValueString()
		}
		if !cmp.Equal(want, got) {
			b.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
		}
	}
}
