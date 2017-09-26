package ifql_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "select",
			raw:     `select()`,
			wantErr: true,
		},
		{
			name: "select with database",
			raw:  `select(db:"mydb").range(start:-4h, stop:-2h).sum()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "sum2",
						Spec: &functions.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
		{
			name: "select with database with range",
			raw:  `select(db:"mydb").range(start:-4h, stop:-2h).sum()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "sum2",
						Spec: &functions.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
		{
			name: "select with database with range and count",
			raw:  `select(db:"mydb").range(start:-4h, stop:-2h).count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count2",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "range1"},
					{Parent: "range1", Child: "count2"},
				},
			},
		},
		{
			name: "select with database where and range",
			raw:  `select(db:"mydb").where(exp:{("t1"=="val1") and ("t2"=="val2")}).range(start:-4h, stop:-2h).count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			name: "select with database where (and with or) and range",
			raw: `select(db:"mydb")
						.where(exp:{
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
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			name: "select with database where including fields",
			raw: `select(db:"mydb")
						.where(exp:{
							("t1"=="val1")
							and
							($ == 10)
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			name: "select with database where with no parens including fields",
			raw: `select(db:"mydb")
						.where(exp:{
							"t1"=="val1"
							and
							$ == 10
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			name: "select with database where with no parens including regex and field",
			raw: `select(db:"mydb")
						.where(exp:{
							"t1"==/val1/
							and
							$ == 10.5
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			name: "select with database regex with escape",
			raw: `select(db:"mydb")
						.where(exp:{
							"t1"==/va\/l1/
						})`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
				},
			},
		},
		{
			name: "select with database with two regex",
			raw: `select(db:"mydb")
						.where(exp:{
							"t1"==/va\/l1/
							and
							"t2" != /val2/
						})`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
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
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
				},
			},
		},
		{
			name: "select with window",
			raw:  `select(db:"mydb").window(start:-4h, every:1h)`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "window1",
						Spec: &functions.WindowOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Every:  query.Duration(time.Hour),
							Period: query.Duration(time.Hour),
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "window1"},
				},
			},
		},
		{
			name: "select with join",
			raw: `
var a = select(db:"dbA").range(start:-1h)
var b = select(db:"dbB").range(start:-1h)
a.join(keys:["host"], exp:{a + b})`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "dbA",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "select2",
						Spec: &functions.SelectOpSpec{
							Database: "dbB",
						},
					},
					{
						ID: "range3",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "join4",
						Spec: &functions.JoinOpSpec{
							Keys: []string{"host"},
							Expression: &expression.BinaryNode{
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
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "range1"},
					{Parent: "select2", Child: "range3"},
					{Parent: "range1", Child: "join4"},
					{Parent: "range3", Child: "join4"},
				},
			},
		},
		{
			name: "select with join and anonymous",
			raw: `var a = select(db:"ifql").where(exp:{"_measurement" == "a"}).range(start:-1h)
			select(db:"ifql").where(exp:{"_measurement" == "b"}).range(start:-1h).join(keys:["t1"], exp:{a/$})`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "_measurement",
									Kind: "tag",
								},
								Right: &expression.StringLiteralNode{
									Value: "a",
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "select3",
						Spec: &functions.SelectOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "where4",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "_measurement",
									Kind: "tag",
								},
								Right: &expression.StringLiteralNode{
									Value: "b",
								},
							},
						},
					},
					{
						ID: "range5",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "join6",
						Spec: &functions.JoinOpSpec{
							Keys: []string{"t1"},
							Expression: &expression.BinaryNode{
								Operator: expression.DivisionOperator,
								Left: &expression.ReferenceNode{
									Name: "a",
									Kind: "identifier",
								},
								Right: &expression.ReferenceNode{
									Name: "$",
									Kind: "field",
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "select3", Child: "where4"},
					{Parent: "where4", Child: "range5"},
					{Parent: "range5", Child: "join6"},
					{Parent: "range2", Child: "join6"},
				},
			},
		},
		{
			name: "select with join with complex expression",
			raw: `var a = select(db:"ifql").where(exp:{"_measurement" == "a"}).range(start:-1h)
			select(db:"ifql").where(exp:{"_measurement" == "b"}).range(start:-1h).join(keys:["t1"], exp:{(a-$)/$})`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select0",
						Spec: &functions.SelectOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "where1",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "_measurement",
									Kind: "tag",
								},
								Right: &expression.StringLiteralNode{
									Value: "a",
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "select3",
						Spec: &functions.SelectOpSpec{
							Database: "ifql",
						},
					},
					{
						ID: "where4",
						Spec: &functions.WhereOpSpec{
							Exp: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "_measurement",
									Kind: "tag",
								},
								Right: &expression.StringLiteralNode{
									Value: "b",
								},
							},
						},
					},
					{
						ID: "range5",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative: -1 * time.Hour,
							},
						},
					},
					{
						ID: "join6",
						Spec: &functions.JoinOpSpec{
							Keys: []string{"t1"},
							Expression: &expression.BinaryNode{
								Operator: expression.DivisionOperator,
								Left: &expression.BinaryNode{
									Operator: expression.SubtractionOperator,
									Left: &expression.ReferenceNode{
										Name: "a",
										Kind: "identifier",
									},
									Right: &expression.ReferenceNode{
										Name: "$",
										Kind: "field",
									},
								},
								Right: &expression.ReferenceNode{
									Name: "$",
									Kind: "field",
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "select0", Child: "where1"},
					{Parent: "where1", Child: "range2"},
					{Parent: "select3", Child: "where4"},
					{Parent: "where4", Child: "range5"},
					{Parent: "range5", Child: "join6"},
					{Parent: "range2", Child: "join6"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ifql.NewQuery(tt.raw, ifql.Debug(false))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			opts := []cmp.Option{cmp.AllowUnexported(query.QuerySpec{}), cmpopts.IgnoreUnexported(query.QuerySpec{})}
			if !cmp.Equal(tt.want, got, opts...) {
				t.Errorf("NewQuery() = -want/+got %s", cmp.Diff(tt.want, got, opts...))
			}
		})
	}
}
