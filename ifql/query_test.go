package ifql

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
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
			raw:  `select(database:"mydb").range(start:-4h, stop:-2h).clear()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "clear",
						Spec: &query.ClearOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "clear"},
				},
			},
		},
		{
			name: "select with database with range",
			raw:  `select(database:"mydb").range(start:-4h, stop:-2h).sum()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "sum",
						Spec: &query.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "sum"},
				},
			},
		},
		{
			name: "select with database with range and count",
			raw:  `select(database:"mydb").range(start:-4h, stop:-2h).count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "count"},
				},
			},
		},
		{
			name: "select with database where and range",
			raw:  `select(database:"mydb").where(exp:{("t1"="val1") and ("t2"="val2")}).range(start:-4h, stop:-2h).count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where",
						Spec: &query.WhereOpSpec{
							Exp: &query.WhereExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeGroupExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeBooleanExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeRef,
														Value: &storage.Node_RefValue{
															RefValue: "t1",
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
												NodeType: storage.NodeTypeBooleanExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeRef,
														Value: &storage.Node_RefValue{
															RefValue: "t2",
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
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "where"},
					{Parent: "where", Child: "range"},
					{Parent: "range", Child: "count"},
				},
			},
		},
		{
			name: "select with database where (and with or) and range",
			raw: `select(database:"mydb")
						.where(exp:{
								(
									("t1"="val1")
									and
									("t2"="val2")
								)
								or
								("t3"="val3")
							})
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where",
						Spec: &query.WhereOpSpec{
							Exp: &query.WhereExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeGroupExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalOr},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeGroupExpression,
												Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeBooleanExpression,
														Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
														Children: []*storage.Node{
															&storage.Node{
																NodeType: storage.NodeTypeRef,
																Value: &storage.Node_RefValue{
																	RefValue: "t1",
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
														NodeType: storage.NodeTypeBooleanExpression,
														Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
														Children: []*storage.Node{
															&storage.Node{
																NodeType: storage.NodeTypeRef,
																Value: &storage.Node_RefValue{
																	RefValue: "t2",
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
												NodeType: storage.NodeTypeBooleanExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeRef,
														Value: &storage.Node_RefValue{
															RefValue: "t3",
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
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "where"},
					{Parent: "where", Child: "range"},
					{Parent: "range", Child: "count"},
				},
			},
		},
		{
			name: "select with database where including fields",
			raw: `select(database:"mydb")
						.where(exp:{
							("t1"="val1")
							and
							($ = 10)
						})
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "where",
						Spec: &query.WhereOpSpec{
							Exp: &query.WhereExpressionSpec{
								Predicate: &storage.Predicate{
									Root: &storage.Node{
										NodeType: storage.NodeTypeGroupExpression,
										Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
										Children: []*storage.Node{
											&storage.Node{
												NodeType: storage.NodeTypeBooleanExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeRef,
														Value: &storage.Node_RefValue{
															RefValue: "t1",
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
												NodeType: storage.NodeTypeBooleanExpression,
												Value:    &storage.Node_Comparison_{Comparison: storage.ComparisonEqual},
												Children: []*storage.Node{
													&storage.Node{
														NodeType: storage.NodeTypeRef,
														Value: &storage.Node_RefValue{
															RefValue: "_field",
														},
													},
													&storage.Node{
														NodeType: storage.NodeTypeLiteral,
														Value: &storage.Node_FloatValue{
															FloatValue: 10.0,
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
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "where"},
					{Parent: "where", Child: "range"},
					{Parent: "range", Child: "count"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewQuery(tt.raw, Debug(false))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			opts := []cmp.Option{cmp.AllowUnexported(query.QuerySpec{}), cmpopts.IgnoreUnexported(query.QuerySpec{})}
			if !cmp.Equal(tt.want, got, opts...) {
				t.Errorf("%q. NewQuery() = -got/+want %s", tt.name, cmp.Diff(tt.want, got, opts...))
			}
		})
	}
}
