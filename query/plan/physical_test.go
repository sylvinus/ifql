package plan_test

import (
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

func TestPhysicalPlanner_Plan(t *testing.T) {
	testCases := []struct {
		name string
		lp   *plan.LogicalPlanSpec
		pp   *plan.PlanSpec
	}{
		{
			name: "single push down",
			lp: &plan.LogicalPlanSpec{
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Bounds: plan.BoundsSpec{
					Start: query.Time{
						IsRelative: true,
						Relative:   -1 * time.Hour,
					},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
							},
							AggregateSet:  true,
							AggregateType: "count",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{},
					},
				},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
			},
		},
		{
			name: "single push down with match",
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("last")},
					},
					plan.ProcedureIDFromOperationID("last"): {
						ID:   plan.ProcedureIDFromOperationID("last"),
						Spec: &functions.LastProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("last"),
				},
			},
			pp: &plan.PlanSpec{
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Bounds: plan.BoundsSpec{
					Start: query.MinTime,
					Stop:  query.Now,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.MinTime,
								Stop:  query.Now,
							},
							LimitSet:      true,
							PointsLimit:   1,
							DescendingSet: true,
							Descending:    true,
						},
						Parents:  nil,
						Children: []plan.ProcedureID{},
					},
				},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
			},
		},
		{
			name: "multiple push down",
			lp: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database: "mydb",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
							},
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					plan.ProcedureIDFromOperationID("limit"): {
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &functions.LimitProcedureSpec{
							N: 10,
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("limit"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 2,
					MemoryBytesQuota: math.MaxInt64,
				},
				Bounds: plan.BoundsSpec{
					Start: query.Time{
						IsRelative: true,
						Relative:   -1 * time.Hour,
					},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
							},
							LimitSet:    true,
							PointsLimit: 10,
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("mean")),
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			PhysicalPlanTestHelper(t, tc.lp, tc.pp)
		})
	}
}

func TestPhysicalPlanner_Plan_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("first"),
					plan.ProcedureIDFromOperationID("last"),
				},
			},
			plan.ProcedureIDFromOperationID("first"): {
				ID:       plan.ProcedureIDFromOperationID("first"),
				Spec:     &functions.FirstProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("last"): {
				ID:       plan.ProcedureIDFromOperationID("last"),
				Spec:     &functions.LastProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("first"),
			plan.ProcedureIDFromOperationID("last"), // last is last so it will be duplicated
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.MinTime,
			Stop:  query.Now,
		},
		Resources: query.ResourceManagement{
			ConcurrencyQuota: 2,
			MemoryBytesQuota: math.MaxInt64,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.MinTime,
						Stop:  query.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    false, // first
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.MinTime,
						Stop:  query.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    true, // last
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
	}

	PhysicalPlanTestHelper(t, lp, want)
}

func TestPhysicalPlanner_Plan_PushDown_Mixed(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							IsRelative: true,
							Relative:   -1 * time.Hour,
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("from")),
				},
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("sum"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
			plan.ProcedureIDFromOperationID("sum"): {
				ID:       plan.ProcedureIDFromOperationID("sum"),
				Spec:     &functions.SumProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("mean"): {
				ID:       plan.ProcedureIDFromOperationID("mean"),
				Spec:     &functions.MeanProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("sum"),
			plan.ProcedureIDFromOperationID("mean"), // Mean can't be pushed down, but sum can
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	meanID := plan.ProcedureIDFromOperationID("mean")
	meanIDDup := plan.ProcedureIDForDuplicate(meanID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.Time{
				IsRelative: true,
				Relative:   -1 * time.Hour,
			},
		},
		Resources: query.ResourceManagement{
			ConcurrencyQuota: 3,
			MemoryBytesQuota: math.MaxInt64,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							IsRelative: true,
							Relative:   -1 * time.Hour,
						},
					},
					AggregateSet:  true,
					AggregateType: "sum",
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.Time{
							IsRelative: true,
							Relative:   -1 * time.Hour,
						},
					},
				},
				Parents: []plan.ProcedureID{},
				Children: []plan.ProcedureID{
					meanIDDup,
				},
			},
			meanIDDup: {
				ID:       meanIDDup,
				Spec:     &functions.MeanProcedureSpec{},
				Parents:  []plan.ProcedureID{fromIDDup},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			fromID,
			meanIDDup,
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
			meanIDDup,
		},
	}

	PhysicalPlanTestHelper(t, lp, want)
}

func PhysicalPlanTestHelper(t *testing.T, lp *plan.LogicalPlanSpec, want *plan.PlanSpec) {
	t.Helper()
	// Setup expected now time
	now := time.Now()
	want.Now = now

	planner := plan.NewPlanner()
	got, err := planner.Plan(lp, nil, now)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(got, want) {
		t.Log(plan.Formatted(got))
		t.Errorf("unexpected physical plan -want/+got:\n%s", cmp.Diff(want, got))
	}
}

var benchmarkPhysicalPlan *plan.PlanSpec

func BenchmarkPhysicalPlan(b *testing.B) {
	var err error
	lp, err := plan.NewLogicalPlanner().Plan(benchmarkQuery)
	if err != nil {
		b.Fatal(err)
	}
	planner := plan.NewPlanner()
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	for n := 0; n < b.N; n++ {
		benchmarkPhysicalPlan, err = planner.Plan(lp, nil, now)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var benchmarkQueryToPhysicalPlan *plan.PlanSpec

func BenchmarkQueryToPhysicalPlan(b *testing.B) {
	lp := plan.NewLogicalPlanner()
	pp := plan.NewPlanner()
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	for n := 0; n < b.N; n++ {
		lp, err := lp.Plan(benchmarkQuery)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkQueryToPhysicalPlan, err = pp.Plan(lp, nil, now)
		if err != nil {
			b.Fatal(err)
		}
	}
}
