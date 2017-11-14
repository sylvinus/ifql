package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestLimitOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"limit","kind":"limit","spec":{"n":10}}`)
	op := &query.Operation{
		ID: "limit",
		Spec: &functions.LimitOpSpec{
			N: 10,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestLimit_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.LimitProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: "one block",
			spec: &functions.LimitProcedureSpec{
				N: 1,
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
					{execute.Time(1), 2.0},
				},
			}},
		},
		{
			name: "multiple blocks",
			spec: &functions.LimitProcedureSpec{
				N: 2,
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
						{execute.Time(2), 2.0},
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
						{execute.Time(1), 3.0},
						{execute.Time(2), 2.0},
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
					},
				},
			},
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
					return functions.NewLimitTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestLimit_PushDown_Single(t *testing.T) {
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
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
			},
			plan.ProcedureIDFromOperationID("limit"): {
				ID: plan.ProcedureIDFromOperationID("limit"),
				Spec: &functions.LimitProcedureSpec{
					N: 10,
				},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("limit"),
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
					LimitSet:    true,
					PointsLimit: 10,
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("select"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestLimit_PushDown_Branch(t *testing.T) {
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
					plan.ProcedureIDFromOperationID("limitA"),
					plan.ProcedureIDFromOperationID("limitB"),
				},
			},
			plan.ProcedureIDFromOperationID("limitA"): {
				ID: plan.ProcedureIDFromOperationID("limitA"),
				Spec: &functions.LimitProcedureSpec{
					N: 10,
				},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("limitB"): {
				ID: plan.ProcedureIDFromOperationID("limitB"),
				Spec: &functions.LimitProcedureSpec{
					N: 55,
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
			plan.ProcedureIDFromOperationID("limitA"),
			plan.ProcedureIDFromOperationID("limitB"), // limitB is last so it will be duplicated
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
					LimitSet:    true,
					PointsLimit: 10,
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
					LimitSet:    true,
					PointsLimit: 55,
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
