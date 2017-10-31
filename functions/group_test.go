package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestGroupOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"group","kind":"group","spec":{"by":["t1","t2"],"keep":["t3","t4"]}}`)
	op := &query.Operation{
		ID: "group",
		Spec: &functions.GroupOpSpec{
			By:   []string{"t1", "t2"},
			Keep: []string{"t3", "t4"},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestGroup_PushDown_Single(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("group")},
			},
			plan.ProcedureIDFromOperationID("group"): {
				ID: plan.ProcedureIDFromOperationID("group"),
				Spec: &functions.GroupProcedureSpec{
					By:     []string{"a", "b"},
					Keep:   []string{"c", "d"},
					Ignore: []string{"e", "f"},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("select")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("group"),
		},
	}

	want := &plan.PlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database:    "mydb",
					GroupingSet: true,
					GroupKeys:   []string{"a", "b"},
					GroupKeep:   []string{"c", "d"},
					GroupIgnore: []string{"e", "f"},
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

func TestGroup_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("select"): {
				ID: plan.ProcedureIDFromOperationID("select"),
				Spec: &functions.SelectProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("groupA"),
					plan.ProcedureIDFromOperationID("groupB"),
				},
			},
			plan.ProcedureIDFromOperationID("groupA"): {
				ID: plan.ProcedureIDFromOperationID("groupA"),
				Spec: &functions.GroupProcedureSpec{
					By:     []string{"a", "b"},
					Keep:   []string{"c", "d"},
					Ignore: []string{"e", "f"},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("select")),
				},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("groupB"): {
				ID: plan.ProcedureIDFromOperationID("groupB"),
				Spec: &functions.GroupProcedureSpec{
					Keep: []string{"C", "D"},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("select")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("groupA"),
			plan.ProcedureIDFromOperationID("groupB"), // groupB is last so it will be duplicated
		},
	}

	selectID := plan.ProcedureIDFromOperationID("select")
	selectIDDup := plan.ProcedureIDForDuplicate(selectID)
	want := &plan.PlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			selectID: {
				ID: selectID,
				Spec: &functions.SelectProcedureSpec{
					Database:    "mydb",
					GroupingSet: true,
					GroupKeys:   []string{"a", "b"},
					GroupKeep:   []string{"c", "d"},
					GroupIgnore: []string{"e", "f"},
				},
				Children: []plan.ProcedureID{},
			},
			selectIDDup: {
				ID: selectIDDup,
				Spec: &functions.SelectProcedureSpec{
					Database:    "mydb",
					GroupingSet: true,
					MergeAll:    true,
					GroupKeys:   []string{},
					GroupKeep:   []string{"C", "D"},
					GroupIgnore: []string{},
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
