package functions_test

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestSortOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"sort","kind":"sort","spec":{"cols":["t1","t2"],"desc":true}}`)
	op := &query.Operation{
		ID: "sort",
		Spec: &functions.SortOpSpec{
			Cols: []string{"t1", "t2"},
			Desc: true,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSort_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
		s := functions.NewSortTransformation(
			d,
			c,
			&functions.SortProcedureSpec{
				Cols: []string{"value"},
				Desc: true,
			},
		)
		return s
	})
}

func TestSort_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.SortProcedureSpec
		data []*executetest.Block
		want []*executetest.Block
	}{
		{
			name: "one block",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"value"},
				Desc: false,
			},
			data: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 1.0},
					{execute.Time(1), 2.0},
				},
			}},
		},
		{
			name: "one block descending",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"value"},
				Desc: true,
			},
			data: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0},
					{execute.Time(1), 1.0},
				},
			}},
		},
		{
			name: "one block multiple columns",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"value", "time"},
				Desc: false,
			},
			data: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0},
					{execute.Time(1), 1.0},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 1.0},
					{execute.Time(2), 2.0},
				},
			}},
		},
		{
			name: "one block multiple columns descending",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"value", "time"},
				Desc: true,
			},
			data: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 1.0},
					{execute.Time(2), 2.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(1), 1.0},
				},
			}},
		},
		{
			name: "multiple blocks",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"value"},
				Desc: false,
			},
			data: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 3.0},
						{execute.Time(2), 2.0},
						{execute.Time(2), 1.0},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0},
						{execute.Time(3), 2.0},
						{execute.Time(4), 1.0},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(1), 3.0},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(4), 1.0},
						{execute.Time(3), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
		},
		{
			name: "one block multiple columns with tags",
			spec: &functions.SortProcedureSpec{
				Cols: []string{"_field", "value"},
				Desc: false,
			},
			data: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "host", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "_field", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "hostA", "F1"},
						{execute.Time(1), 2.0, "hostA", "F2"},
						{execute.Time(1), 3.0, "hostA", "F3"},
						{execute.Time(2), 4.0, "hostA", "F1"},
						{execute.Time(2), 5.0, "hostA", "F2"},
						{execute.Time(2), 6.0, "hostA", "F3"},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "host", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "_field", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "hostB", "F1"},
						{execute.Time(1), 2.0, "hostB", "F2"},
						{execute.Time(1), 3.0, "hostB", "F3"},
						{execute.Time(2), 4.0, "hostB", "F1"},
						{execute.Time(2), 5.0, "hostB", "F2"},
						{execute.Time(2), 6.0, "hostB", "F3"},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "host", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "_field", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "hostA", "F1"},
						{execute.Time(2), 4.0, "hostA", "F1"},
						{execute.Time(1), 2.0, "hostA", "F2"},
						{execute.Time(2), 5.0, "hostA", "F2"},
						{execute.Time(1), 3.0, "hostA", "F3"},
						{execute.Time(2), 6.0, "hostA", "F3"},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "host", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "_field", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "hostB", "F1"},
						{execute.Time(2), 4.0, "hostB", "F1"},
						{execute.Time(1), 2.0, "hostB", "F2"},
						{execute.Time(2), 5.0, "hostB", "F2"},
						{execute.Time(1), 3.0, "hostB", "F3"},
						{execute.Time(2), 6.0, "hostB", "F3"},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewBlockBuilderCache()
			c.SetTriggerSpec(execute.DefaultTriggerSpec)

			st := functions.NewSortTransformation(
				d,
				c,
				tc.spec,
			)

			parentID := executetest.RandomDatasetID()
			for _, b := range tc.data {
				st.Process(parentID, b)
			}

			got := executetest.BlocksFromCache(c)

			sort.Sort(executetest.SortedBlocks(got))
			sort.Sort(executetest.SortedBlocks(tc.want))

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
