package ifql_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t1",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_StringValue{
															StringValue: "val1",
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t2",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_StringValue{
															StringValue: "val2",
														},
													},
												},
											},
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalOr},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeLogicalExpression,
												Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeComparisonExpression,
														Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
														Children: []*storage.Node{
															&storage.Node{
																NodeType: storage.NodeTypeTagRef,
																Value: &storage.Node_TagRefValue{
																	TagRefValue: "t1",
																},
															},
															&storage.Node{
																NodeType: storage.NodeTypeLiteral,
																Value: &storage.Node_StringValue{
																	StringValue: "val1",
																},
															},
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeComparisonExpression,
														Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
														Children: []*storage.Node{
															&storage.Node{
																NodeType: storage.NodeTypeTagRef,
																Value: &storage.Node_TagRefValue{
																	TagRefValue: "t2",
																},
															},
															&storage.Node{
																NodeType: storage.NodeTypeLiteral,
																Value: &storage.Node_StringValue{
																	StringValue: "val2",
																},
															},
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t3",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_StringValue{
															StringValue: "val3",
														},
													},
												},
											},
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t1",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_StringValue{
															StringValue: "val1",
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "_field",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_IntegerValue{
															IntegerValue: 10.0,
														},
													},
												},
											},
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t1",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_StringValue{
															StringValue: "val1",
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "_field",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_IntegerValue{
															IntegerValue: 10.0,
														},
													},
												},
											},
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonRegex},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t1",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_RegexValue{
															RegexValue: "val1",
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "_field",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_FloatValue{
															FloatValue: 10.5,
														},
													},
												},
											},
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeComparisonExpression,
										Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonRegex},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeTagRef,
												Value: &storage.Node_TagRefValue{
													TagRefValue: "t1",
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeLiteral,
												Value: &storage.Node_RegexValue{
													RegexValue: "va/l1",
												},
											},
										},
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
			name: "select with database with to regex",
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
							Exp: &query.ExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeLogicalExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonRegex},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t1",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_RegexValue{
															RegexValue: "va/l1",
														},
													},
												},
											},
											&storage.Node{
												NodeType: storage.NodeTypeComparisonExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonNotRegex},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeTagRef,
														Value: &storage.Node_TagRefValue{
															TagRefValue: "t2",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_RegexValue{
															RegexValue: "val2",
														},
													},
												},
											},
										},
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
