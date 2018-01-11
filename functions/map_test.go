package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestMap_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "simple static map",
			Raw:  `from(db:"mydb").map(fn: (r) => r._value + 1)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "map1",
						Spec: &functions.MapOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left: &ast.MemberExpression{
										Object: &ast.Identifier{
											Name: "r",
										},
										Property: &ast.Identifier{Name: "_value"},
									},
									Right: &ast.IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "map1"},
				},
			},
		},
		{
			Name: "resolve map",
			Raw:  `x = 2 from(db:"mydb").map(fn: (r) => r._value + x)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "map1",
						Spec: &functions.MapOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left: &ast.MemberExpression{
										Object: &ast.Identifier{
											Name: "r",
										},
										Property: &ast.Identifier{Name: "_value"},
									},
									Right: &ast.IntegerLiteral{Value: 2},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "map1"},
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

func TestMapOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"map",
		"kind":"map",
		"spec":{
			"fn":{
				"type": "ArrowFunctionExpression",
				"params": [{"type":"Identifier","name":"r"}],
				"body":{
					"type":"BinaryExpression",
					"operator": "-",
					"left":{
						"type":"MemberExpression",
						"object": {
							"type": "Identifier",
							"name":"r"
						},
						"property": {"type":"StringLiteral","value":"_value"}
					},
					"right":{
						"type":"FloatLiteral",
						"value": 5.6
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "map",
		Spec: &functions.MapOpSpec{
			Fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Identifier{{Name: "r"}},
				Body: &ast.BinaryExpression{
					Operator: ast.SubtractionOperator,
					Left: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "r",
						},
						Property: &ast.StringLiteral{Value: "_value"},
					},
					Right: &ast.FloatLiteral{Value: 5.6},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
func TestMap_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.MapProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: `_value+5`,
			spec: &functions.MapProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Identifier{{Name: "r"}},
					Body: &ast.BinaryExpression{
						Operator: ast.AdditionOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_value"},
						},
						Right: &ast.FloatLiteral{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 6.0},
					{execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value*_value`,
			spec: &functions.MapProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Identifier{{Name: "r"}},
					Body: &ast.BinaryExpression{
						Operator: ast.MultiplicationOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_value"},
						},
						Right: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_value"},
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 36.0},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
					f, err := functions.NewMapTransformation(d, c, tc.spec)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}