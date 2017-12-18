package functions_test

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestJoin_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "basic two-way join",
			Raw: `
var a = from(db:"dbA").range(start:-1h)
var b = from(db:"dbB").range(start:-1h)
join(tables:{a:a,b:b}, on:["host"], fn: (t) => t.a["_value"] + t.b["_value"])`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "dbA",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								IsRelative: true,
							},
						},
					},
					{
						ID: "from2",
						Spec: &functions.FromOpSpec{
							Database: "dbB",
						},
					},
					{
						ID: "range3",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								IsRelative: true,
							},
						},
					},
					{
						ID: "join4",
						Spec: &functions.JoinOpSpec{
							On:         []string{"host"},
							TableNames: map[query.OperationID]string{"range1": "a", "range3": "b"},
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "t"}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left: &ast.MemberExpression{
										Object: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "t",
											},
											Property: &ast.Identifier{Name: "a"},
										},
										Property: &ast.StringLiteral{Value: "_value"},
									},
									Right: &ast.MemberExpression{
										Object: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "t",
											},
											Property: &ast.Identifier{Name: "b"},
										},
										Property: &ast.StringLiteral{Value: "_value"},
									},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "from2", Child: "range3"},
					{Parent: "range1", Child: "join4"},
					{Parent: "range3", Child: "join4"},
				},
			},
		},
		{
			Name: "error: join as chain",
			Raw: `
				var a = from(db:"dbA").range(start:-1h)
				var b = from(db:"dbB").range(start:-1h)
				a.join(tables:{a:a,b:b}, on:["host"], fn: (t) => t.a["_value"] + t.b["_value"])
			`,
			WantErr: true,
		},
		{
			Name: "from with join with complex ast",
			Raw: `
				var a = from(db:"ifql").range(start:-1h)
				var b = from(db:"ifql").range(start:-1h)
				join(tables:{a:a,b:b}, on:["t1"], fn: (t) => (t.a["_value"]-t.b["_value"])/t.b["_value"])
			`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								IsRelative: true,
							},
						},
					},
					{
						ID: "from2",
						Spec: &functions.FromOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "range3",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								IsRelative: true,
							},
						},
					},
					{
						ID: "join4",
						Spec: &functions.JoinOpSpec{
							On:         []string{"t1"},
							TableNames: map[query.OperationID]string{"range1": "a", "range3": "b"},
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "t"}},
								Body: &ast.BinaryExpression{
									Operator: ast.DivisionOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.SubtractionOperator,
										Left: &ast.MemberExpression{
											Object: &ast.MemberExpression{
												Object: &ast.Identifier{
													Name: "t",
												},
												Property: &ast.Identifier{Name: "a"},
											},
											Property: &ast.StringLiteral{Value: "_value"},
										},
										Right: &ast.MemberExpression{
											Object: &ast.MemberExpression{
												Object: &ast.Identifier{
													Name: "t",
												},
												Property: &ast.Identifier{Name: "b"},
											},
											Property: &ast.StringLiteral{Value: "_value"},
										},
									},
									Right: &ast.MemberExpression{
										Object: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "t",
											},
											Property: &ast.Identifier{Name: "b"},
										},
										Property: &ast.StringLiteral{Value: "_value"},
									},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "from2", Child: "range3"},
					{Parent: "range1", Child: "join4"},
					{Parent: "range3", Child: "join4"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestJoinOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"join",
		"kind":"join",
		"spec":{
			"on":["t1","t2"],
			"table_names": {"sum1":"a","count3":"b"},
			"fn":{
				"params": [{"type":"Identifier","name":"t"}],
				"body":{
					"type":"BinaryExpression",
					"operator": "+",
					"left": {
						"type": "MemberExpression",
						"object": {
							"type":"Identifier",
							"name":"a"
						},
						"property": {"type":"StringLiteral","value":"_value"}
					},
					"right":{
						"type": "MemberExpression",
						"object": {
							"type":"Identifier",
							"name":"b"
						},
						"property": {"type":"StringLiteral","value":"_value"}
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "join",
		Spec: &functions.JoinOpSpec{
			On:         []string{"t1", "t2"},
			TableNames: map[query.OperationID]string{"sum1": "a", "count3": "b"},
			Fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Identifier{{Name: "t"}},
				Body: &ast.BinaryExpression{
					Operator: ast.AdditionOperator,
					Left: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "a",
						},
						Property: &ast.StringLiteral{Value: "_value"},
					},
					Right: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "b",
						},
						Property: &ast.StringLiteral{Value: "_value"},
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMergeJoin_Process(t *testing.T) {
	addFunction := &ast.ArrowFunctionExpression{
		Params: []*ast.Identifier{{Name: "t"}},
		Body: &ast.BinaryExpression{
			Operator: ast.AdditionOperator,
			Left: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "t",
					},
					Property: &ast.Identifier{Name: "a"},
				},
				Property: &ast.StringLiteral{Value: "_value"},
			},
			Right: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "t",
					},
					Property: &ast.Identifier{Name: "b"},
				},
				Property: &ast.StringLiteral{Value: "_value"},
			},
		},
	}
	parentID0 := plantest.RandomProcedureID()
	parentID1 := plantest.RandomProcedureID()
	addTableNames := map[plan.ProcedureID]string{
		parentID0: "a",
		parentID1: "b",
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
				Fn:         addFunction,
				TableNames: addTableNames,
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
			parents := []execute.DatasetID{execute.DatasetID(parentID0), execute.DatasetID(parentID1)}

			tableNames := make(map[execute.DatasetID]string, len(tc.spec.TableNames))
			for pid, name := range tc.spec.TableNames {
				tableNames[execute.DatasetID(pid)] = name
			}

			d := executetest.NewDataset(executetest.RandomDatasetID())
			joinExpr, err := functions.NewJoinFunction(tc.spec.Fn, parents, tableNames)
			if err != nil {
				t.Fatal(err)
			}
			c := functions.NewMergeJoinCache(joinExpr, executetest.UnlimitedAllocator, parents[0], parents[1])
			c.SetTriggerSpec(execute.DefaultTriggerSpec)
			jt := functions.NewMergeJoinTransformation(d, c, tc.spec, parents)

			l := len(tc.data0)
			if len(tc.data1) > l {
				l = len(tc.data1)
			}
			for i := 0; i < l; i++ {
				if i < len(tc.data0) {
					if err := jt.Process(parents[0], tc.data0[i]); err != nil {
						t.Fatal(err)
					}
				}
				if i < len(tc.data1) {
					if err := jt.Process(parents[1], tc.data1[i]); err != nil {
						t.Fatal(err)
					}
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
