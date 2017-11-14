package functions_test

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestJoinOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"join",
		"kind":"join",
		"spec":{
			"on":["t1","t2"],
			"expression":{
				"root":{
					"type":"binary",
					"operator": "+",
					"left":{
						"type":"reference",
						"name":"a",
						"kind":"identifier"
					},
					"right":{
						"type":"reference",
						"name":"b",
						"kind":"identifier"
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "join",
		Spec: &functions.JoinOpSpec{
			On: []string{"t1", "t2"},
			Expression: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.AdditionOperator,
					Left: &expression.ReferenceNode{
						Name: "a",
						Kind: "identifier",
					},
					Right: &expression.ReferenceNode{
						Name: "b",
						Kind: "identifier",
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMergeJoin_Process(t *testing.T) {
	addExpression := expression.Expression{
		Root: &expression.BinaryNode{
			Operator: expression.AdditionOperator,
			Left: &expression.ReferenceNode{
				Name: "$",
				Kind: "identifier",
			},
			Right: &expression.ReferenceNode{
				Name: "b",
				Kind: "identifier",
			},
		},
	}
	testCases := []struct {
		skip  bool
		name  string
		spec  *functions.MergeJoinProcedureSpec
		data0 []*executetest.Block // data from parent 0
		data1 []*executetest.Block // data from parent 1
		want  []*executetest.Block
	}{
		{
			name: "simple inner",
			spec: &functions.MergeJoinProcedureSpec{
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0},
						{execute.Time(2), 20.0},
						{execute.Time(3), 30.0},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0},
						{execute.Time(2), 22.0},
						{execute.Time(3), 33.0},
					},
				},
			},
		},
		{
			name: "simple inner with ints",
			spec: &functions.MergeJoinProcedureSpec{
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(1)},
						{execute.Time(2), int64(2)},
						{execute.Time(3), int64(3)},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(10)},
						{execute.Time(2), int64(20)},
						{execute.Time(3), int64(30)},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(11)},
						{execute.Time(2), int64(22)},
						{execute.Time(3), int64(33)},
					},
				},
			},
		},
		{
			name: "inner with missing values",
			spec: &functions.MergeJoinProcedureSpec{
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0},
						{execute.Time(3), 30.0},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0},
						{execute.Time(3), 33.0},
					},
				},
			},
		},
		{
			name: "inner with multiple matches",
			spec: &functions.MergeJoinProcedureSpec{
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0},
						{execute.Time(1), 10.1},
						{execute.Time(2), 20.0},
						{execute.Time(3), 30.0},
						{execute.Time(3), 30.1},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0},
						{execute.Time(1), 11.1},
						{execute.Time(2), 22.0},
						{execute.Time(3), 33.0},
						{execute.Time(3), 33.1},
					},
				},
			},
		},
		{
			name: "inner with common tags",
			spec: &functions.MergeJoinProcedureSpec{
				On:         []string{"t1"},
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a"},
						{execute.Time(2), 2.0, "a"},
						{execute.Time(3), 3.0, "a"},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0, "a"},
						{execute.Time(2), 20.0, "a"},
						{execute.Time(3), 30.0, "a"},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0, "a"},
						{execute.Time(2), 22.0, "a"},
						{execute.Time(3), 33.0, "a"},
					},
				},
			},
		},
		{
			name: "inner with extra attributes",
			spec: &functions.MergeJoinProcedureSpec{
				On:         []string{"t1"},
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a"},
						{execute.Time(1), 1.5, "b"},
						{execute.Time(2), 2.0, "a"},
						{execute.Time(2), 2.5, "b"},
						{execute.Time(3), 3.0, "a"},
						{execute.Time(3), 3.5, "b"},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0, "a"},
						{execute.Time(1), 10.1, "b"},
						{execute.Time(2), 20.0, "a"},
						{execute.Time(2), 20.1, "b"},
						{execute.Time(3), 30.0, "a"},
						{execute.Time(3), 30.1, "b"},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0, "a"},
						{execute.Time(1), 11.6, "b"},
						{execute.Time(2), 22.0, "a"},
						{execute.Time(2), 22.6, "b"},
						{execute.Time(3), 33.0, "a"},
						{execute.Time(3), 33.6, "b"},
					},
				},
			},
		},
		{
			name: "inner with tags and extra attributes",
			spec: &functions.MergeJoinProcedureSpec{
				On:         []string{"t1", "t2"},
				Expression: addExpression,
			},
			data0: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a", "x"},
						{execute.Time(1), 1.5, "a", "y"},
						{execute.Time(2), 2.0, "a", "x"},
						{execute.Time(2), 2.5, "a", "y"},
						{execute.Time(3), 3.0, "a", "x"},
						{execute.Time(3), 3.5, "a", "y"},
					},
				},
			},
			data1: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0, "a", "x"},
						{execute.Time(1), 10.1, "a", "y"},
						{execute.Time(2), 20.0, "a", "x"},
						{execute.Time(2), 20.1, "a", "y"},
						{execute.Time(3), 30.0, "a", "x"},
						{execute.Time(3), 30.1, "a", "y"},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  10,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
						{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
					},
					Data: [][]interface{}{
						{execute.Time(1), 11.0, "a", "x"},
						{execute.Time(1), 11.6, "a", "y"},
						{execute.Time(2), 22.0, "a", "x"},
						{execute.Time(2), 22.6, "a", "y"},
						{execute.Time(3), 33.0, "a", "x"},
						{execute.Time(3), 33.6, "a", "y"},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			d := executetest.NewDataset(executetest.RandomDatasetID())
			joinExpr, err := functions.NewExpressionSpec(tc.spec.Expression)
			if err != nil {
				t.Fatal(err)
			}
			c := functions.NewMergeJoinCache(joinExpr)
			c.SetTriggerSpec(execute.DefaultTriggerSpec)
			jt := functions.NewMergeJoinTransformation(d, c, tc.spec)

			parentID0 := executetest.RandomDatasetID()
			parentID1 := executetest.RandomDatasetID()
			jt.SetParents([]execute.DatasetID{parentID0, parentID1})

			l := len(tc.data0)
			if len(tc.data1) > l {
				l = len(tc.data1)
			}
			for i := 0; i < l; i++ {
				if i < len(tc.data0) {
					jt.Process(parentID0, tc.data0[i])
				}
				if i < len(tc.data1) {
					jt.Process(parentID1, tc.data1[i])
				}
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
