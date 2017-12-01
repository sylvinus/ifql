package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/ifql/ifqltest"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFilter_NewQuery(t *testing.T) {
	tests := []ifqltest.NewQueryTestCase{
		{
			Name: "from with database filter and range",
			Raw:  `from(db:"mydb").filter(exp:{("t1"=="val1") and ("t2"=="val2")}).range(start:-4h, stop:-2h).count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.AndOperator,
									Left: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "t1",
											Kind: "tag",
										},
										Right: &expression.StringLiteralNode{
											Value: "val1",
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "t2",
											Kind: "tag",
										},
										Right: &expression.StringLiteralNode{
											Value: "val2",
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter (and with or) and range",
			Raw: `from(db:"mydb")
						.filter(exp:{
								(
									("t1"=="val1")
									and
									("t2"=="val2")
								)
								or
								("t3"=="val3")
							})
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.OrOperator,
									Left: &expression.BinaryNode{
										Operator: expression.AndOperator,
										Left: &expression.BinaryNode{
											Operator: expression.EqualOperator,
											Left: &expression.ReferenceNode{
												Name: "t1",
												Kind: "tag",
											},
											Right: &expression.StringLiteralNode{
												Value: "val1",
											},
										},
										Right: &expression.BinaryNode{
											Operator: expression.EqualOperator,
											Left: &expression.ReferenceNode{
												Name: "t2",
												Kind: "tag",
											},
											Right: &expression.StringLiteralNode{
												Value: "val2",
											},
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "t3",
											Kind: "tag",
										},
										Right: &expression.StringLiteralNode{
											Value: "val3",
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter including fields",
			Raw: `from(db:"mydb")
						.filter(exp:{
							("t1"=="val1")
							and
							($ == 10)
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.AndOperator,
									Left: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "t1",
											Kind: "tag",
										},
										Right: &expression.StringLiteralNode{
											Value: "val1",
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "$",
											Kind: "field",
										},
										Right: &expression.IntegerLiteralNode{
											Value: 10,
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including fields",
			Raw: `from(db:"mydb")
						.filter(exp:{
							"t1"=="val1"
							and
							$ == 10
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.AndOperator,
									Left: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "t1",
											Kind: "tag",
										},
										Right: &expression.StringLiteralNode{
											Value: "val1",
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "$",
											Kind: "field",
										},
										Right: &expression.IntegerLiteralNode{
											Value: 10,
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including regex and field",
			Raw: `from(db:"mydb")
						.filter(exp:{
							"t1"==/val1/
							and
							$ == 10.5
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.AndOperator,
									Left: &expression.BinaryNode{
										Operator: expression.RegexpMatchOperator,
										Left: &expression.ReferenceNode{
											Name: "t1",
											Kind: "tag",
										},
										Right: &expression.RegexpLiteralNode{
											Value: "val1",
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.EqualOperator,
										Left: &expression.ReferenceNode{
											Name: "$",
											Kind: "field",
										},
										Right: &expression.FloatLiteralNode{
											Value: 10.5,
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database regex with escape",
			Raw: `from(db:"mydb")
						.filter(exp:{
							"t1"==/va\/l1/
						})`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.RegexpMatchOperator,
									Left: &expression.ReferenceNode{
										Name: "t1",
										Kind: "tag",
									},
									Right: &expression.RegexpLiteralNode{
										Value: `va/l1`,
									},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with database with two regex",
			Raw: `from(db:"mydb")
						.filter(exp:{
							"t1"==/va\/l1/
							and
							"t2" != /val2/
						})`,
			Want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Expression: expression.Expression{
								Root: &expression.BinaryNode{
									Operator: expression.AndOperator,
									Left: &expression.BinaryNode{
										Operator: expression.RegexpMatchOperator,
										Left: &expression.ReferenceNode{
											Name: "t1",
											Kind: "tag",
										},
										Right: &expression.RegexpLiteralNode{
											Value: `va/l1`,
										},
									},
									Right: &expression.BinaryNode{
										Operator: expression.RegexpNotMatchOperator,
										Left: &expression.ReferenceNode{
											Name: "t2",
											Kind: "tag",
										},
										Right: &expression.RegexpLiteralNode{
											Value: `val2`,
										},
									},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ifqltest.NewQueryTestHelper(t, tc)
		})
	}
}
func TestFilterOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"filter",
		"kind":"filter",
		"spec":{
			"expression":{
				"root":{
					"type":"binary",
					"operator": "!=",
					"left":{
						"type":"reference",
						"name":"_measurement",
						"kind":"tag"
					},
					"right":{
						"type":"stringLiteral",
						"value":"mem"
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "filter",
		Spec: &functions.FilterOpSpec{
			Expression: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.NotEqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "tag",
					},
					Right: &expression.StringLiteralNode{
						Value: "mem",
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFilter_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.FilterProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: "$>5",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.GreaterThanOperator,
						Left: &expression.ReferenceNode{
							Name: "$",
						},
						Right: &expression.FloatLiteralNode{
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
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
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
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "$>5 multiple blocks",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.GreaterThanOperator,
						Left: &expression.ReferenceNode{
							Name: "$",
						},
						Right: &expression.FloatLiteralNode{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{
				&executetest.Block{
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
						{execute.Time(2), 6.0},
						{execute.Time(2), 1.0},
					},
				},
				&executetest.Block{
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
						{execute.Time(4), 8.0},
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
						{execute.Time(2), 6.0},
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
						{execute.Time(4), 8.0},
					},
				},
			},
		},
		{
			name: "$>5 and t1 = a and t2 = y",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.AndOperator,
						Left: &expression.BinaryNode{
							Operator: expression.GreaterThanOperator,
							Left: &expression.ReferenceNode{
								Name: "$",
							},
							Right: &expression.FloatLiteralNode{
								Value: 5,
							},
						},
						Right: &expression.BinaryNode{
							Operator: expression.AndOperator,
							Left: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "t1",
								},
								Right: &expression.StringLiteralNode{
									Value: "a",
								},
							},
							Right: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "t2",
								},
								Right: &expression.StringLiteralNode{
									Value: "y",
								},
							},
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
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a", "x"},
					{execute.Time(2), 6.0, "a", "x"},
					{execute.Time(3), 8.0, "a", "y"},
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
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(3), 8.0, "a", "y"},
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
					return functions.NewFilterTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestFilter_PushDown(t *testing.T) {
	spec := &functions.FilterProcedureSpec{
		Expression: expression.Expression{
			Root: &expression.BinaryNode{
				Operator: expression.NotEqualOperator,
				Left: &expression.ReferenceNode{
					Name: "_measurement",
					Kind: "tag",
				},
				Right: &expression.StringLiteralNode{
					Value: "mem",
				},
			},
		},
	}
	root := &plan.Procedure{
		Spec: new(functions.FromProcedureSpec),
	}
	want := &plan.Procedure{
		Spec: &functions.FromProcedureSpec{
			FilterSet: true,
			Filter: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.NotEqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "tag",
					},
					Right: &expression.StringLiteralNode{
						Value: "mem",
					},
				},
			},
		},
	}

	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, false, want)
}

func TestFilter_PushDown_Duplicate(t *testing.T) {
	spec := &functions.FilterProcedureSpec{
		Expression: expression.Expression{
			Root: &expression.BinaryNode{
				Operator: expression.NotEqualOperator,
				Left: &expression.ReferenceNode{
					Name: "_measurement",
					Kind: "tag",
				},
				Right: &expression.StringLiteralNode{
					Value: "mem",
				},
			},
		},
	}
	root := &plan.Procedure{
		Spec: &functions.FromProcedureSpec{
			FilterSet: true,
			Filter: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.NotEqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "tag",
					},
					Right: &expression.StringLiteralNode{
						Value: "cpu",
					},
				},
			},
		},
	}
	want := &plan.Procedure{
		// Expect the duplicate has been reset to zero values
		Spec: new(functions.FromProcedureSpec),
	}

	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, true, want)
}
