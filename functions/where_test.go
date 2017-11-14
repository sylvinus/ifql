package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestWhereOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"where",
		"kind":"where",
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
		ID: "where",
		Spec: &functions.WhereOpSpec{
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

func TestWhere_PushDown_Single(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("select")},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("where")},
			},
			plan.ProcedureIDFromOperationID("where"): {
				ID: plan.ProcedureIDFromOperationID("where"),
				Spec: &functions.WhereProcedureSpec{
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
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("where"),
		},
	}

	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					WhereSet: true,
					Where: expression.Expression{
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
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			(plan.ProcedureIDFromOperationID("select")),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestWhere_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("select")},
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("whereA"),
					plan.ProcedureIDFromOperationID("whereB"),
				},
			},
			plan.ProcedureIDFromOperationID("whereA"): {
				ID: plan.ProcedureIDFromOperationID("whereA"),
				Spec: &functions.WhereProcedureSpec{
					Expression: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "A",
							},
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("whereB"): {
				ID: plan.ProcedureIDFromOperationID("whereB"),
				Spec: &functions.WhereProcedureSpec{
					Expression: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "B",
							},
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("whereA"),
			plan.ProcedureIDFromOperationID("whereB"), // WhereB is last so it will be duplicated
		},
	}

	selectID := plan.ProcedureIDFromOperationID("select")
	selectIDDup := plan.ProcedureIDForDuplicate(selectID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			selectID: {
				ID: selectID,
				Spec: &functions.SelectProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					WhereSet: true,
					Where: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "A",
							},
						},
					},
				},
				Children: []plan.ProcedureID{},
			},
			selectIDDup: {
				ID: selectIDDup,
				Spec: &functions.SelectProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					WhereSet: true,
					Where: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "B",
							},
						},
					},
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			selectID,
			selectIDDup,
		},
		Order: []plan.ProcedureID{
			selectID,
			selectIDDup,
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}
