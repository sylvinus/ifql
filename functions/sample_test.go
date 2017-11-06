package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestSampleOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"sample","kind":"sample","spec":{"useRowTime":true, "n":5, "pos":0}}`)
	op := &query.Operation{
		ID: "sample",
		Spec: &functions.SampleOpSpec{
			UseRowTime: true,
			N:          5,
			Pos:        0,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSample_Process(t *testing.T) {
	testCases := []struct {
		name     string
		data     *executetest.Block
		want     []execute.Row
		selector *functions.SampleSelector
	}{
		{
			selector: &functions.SampleSelector{
				N:   1,
				Pos: 0,
			},
			name: "everything",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(0), 7.0, "a", "y"}},
				{Values: []interface{}{execute.Time(10), 5.0, "a", "x"}},
				{Values: []interface{}{execute.Time(20), 9.0, "a", "y"}},
				{Values: []interface{}{execute.Time(30), 4.0, "a", "x"}},
				{Values: []interface{}{execute.Time(40), 6.0, "a", "y"}},
				{Values: []interface{}{execute.Time(50), 8.0, "a", "x"}},
				{Values: []interface{}{execute.Time(60), 1.0, "a", "y"}},
				{Values: []interface{}{execute.Time(70), 2.0, "a", "x"}},
				{Values: []interface{}{execute.Time(80), 3.0, "a", "y"}},
				{Values: []interface{}{execute.Time(90), 10.0, "a", "x"}},
			},
		},
		{
			selector: &functions.SampleSelector{
				N:   2,
				Pos: 0,
			},
			name: "every-other-even",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(0), 7.0, "a", "y"}},
				{Values: []interface{}{execute.Time(20), 9.0, "a", "y"}},
				{Values: []interface{}{execute.Time(40), 6.0, "a", "y"}},
				{Values: []interface{}{execute.Time(60), 1.0, "a", "y"}},
				{Values: []interface{}{execute.Time(80), 3.0, "a", "y"}},
			},
		},
		{
			selector: &functions.SampleSelector{
				N:   2,
				Pos: 1,
			},
			name: "every-other-odd",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(10), 5.0, "a", "x"}},
				{Values: []interface{}{execute.Time(30), 4.0, "a", "x"}},
				{Values: []interface{}{execute.Time(50), 8.0, "a", "x"}},
				{Values: []interface{}{execute.Time(70), 2.0, "a", "x"}},
				{Values: []interface{}{execute.Time(90), 10.0, "a", "x"}},
			},
		},
		{
			selector: &functions.SampleSelector{
				N:   3,
				Pos: 0,
			},
			name: "every-third-0",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(0), 7.0, "a", "y"}},
				{Values: []interface{}{execute.Time(30), 4.0, "a", "x"}},
				{Values: []interface{}{execute.Time(60), 1.0, "a", "y"}},
				{Values: []interface{}{execute.Time(90), 10.0, "a", "x"}},
			},
		},
		{
			selector: &functions.SampleSelector{
				N:   3,
				Pos: 1,
			},
			name: "every-third-1",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(10), 5.0, "a", "x"}},
				{Values: []interface{}{execute.Time(40), 6.0, "a", "y"}},
				{Values: []interface{}{execute.Time(70), 2.0, "a", "x"}},
			},
		},
		{
			selector: &functions.SampleSelector{
				N:   3,
				Pos: 2,
			},
			name: "every-third-2",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 10.0, "a", "x"},
				},
			},
			want: []execute.Row{
				{Values: []interface{}{execute.Time(20), 9.0, "a", "y"}},
				{Values: []interface{}{execute.Time(50), 8.0, "a", "x"}},
				{Values: []interface{}{execute.Time(80), 3.0, "a", "y"}},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.SelectorFuncTestHelper(
				t,
				tc.selector,
				tc.data,
				tc.want,
			)
		})
	}
}

func BenchmarkSample(b *testing.B) {
	ss := &functions.SampleSelector{
		N:   10,
		Pos: 0,
	}
	executetest.SelectorFuncBenchmarkHelper(b, ss, NormalBlock)
}
