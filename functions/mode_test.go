package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestModeOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"mode","kind":"mode"}`)
	op := &query.Operation{
		ID:   "mode",
		Spec: &functions.ModeOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMode_ProcessFloat(t *testing.T) {
	testCases := []struct {
		name string
		data []float64
		want float64
	}{
		{
			name: "zero",
			data: []float64{0, 0, 0},
			want: 0,
		},
		{
			name: "nonzero",
			data: []float64{0, 1, 2, 2, 2, 3, 3, 3, 3, 3},
			want: 3,
		},
		{
			name: "0 for empty",
			data: []float64{},
			want: 0,
		},
		{
			name: "use truncated normal data",
			data: TruncatedData,
			want: 9.0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncTestHelper(
				t,
				new(functions.ModeAgg),
				tc.data,
				tc.want,
			)
		})
	}
}

func TestMode_ProcessInt(t *testing.T) {
	testCases := []struct {
		name string
		data []int64
		want int64
	}{
		{
			name: "zero",
			data: []int64{0, 0, 0},
			want: 0,
		},
		{
			name: "nonzero",
			data: []int64{0, 1, 2, 2, 2, 3, 3, 3, 3, 3},
			want: 3,
		},
		{
			name: "empty",
			data: []int64{},
			want: 0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncIntTestHelper(
				t,
				new(functions.ModeAgg),
				tc.data,
				tc.want,
			)
		})
	}
}

func TestMode_ProcessUInt(t *testing.T) {
	testCases := []struct {
		name string
		data []uint64
		want uint64
	}{
		{
			name: "zero",
			data: []uint64{0, 0, 0},
			want: 0,
		},
		{
			name: "nonzero",
			data: []uint64{0, 1, 2, 2, 2, 3, 3, 3, 3, 3},
			want: 3,
		},
		{
			name: "empty",
			data: []uint64{},
			want: 0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncUIntTestHelper(
				t,
				new(functions.ModeAgg),
				tc.data,
				tc.want,
			)
		})
	}
}

func TestMode_ProcessBool(t *testing.T) {
	testCases := []struct {
		name string
		data []bool
		want bool
	}{
		{
			name: "true",
			data: []bool{true, true, false},
			want: true,
		},
		{
			name: "empty",
			data: []bool{},
			want: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncBoolTestHelper(
				t,
				new(functions.ModeAgg),
				tc.data,
				tc.want,
			)
		})
	}
}

func TestMode_ProcessString(t *testing.T) {
	testCases := []struct {
		name string
		data []string
		want string
	}{
		{
			name: "howdy most common",
			data: []string{"howdy", "howdy", "doody"},
			want: "howdy",
		},
		{
			name: "empty",
			data: []string{},
			want: "",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncStringTestHelper(
				t,
				new(functions.ModeAgg),
				tc.data,
				tc.want,
			)
		})
	}
}

func BenchmarkMode(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		new(functions.ModeAgg),
		TruncatedData,
		9.0,
	)
}
