package plantest

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query/plan"
)

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
		t.Errorf("unexpected physical plan -want/+got:\n%s", cmp.Diff(want, got))
	}
}
