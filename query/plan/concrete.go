package plan

import "time"

type PlanSpec struct {
	Operations []Operation
	Datasets   []Dataset
	Now        time.Time
}

type Planner interface {
	// Plan create a plan from the abstract plan and available storage.
	Plan(p *AbstractPlanSpec, s Storage) (*PlanSpec, error)
}

type planner struct {
	plan *PlanSpec
}

func (p *planner) Plan(ap *AbstractPlanSpec, s Storage) (*PlanSpec, error) {
	p.plan = new(PlanSpec)

	return p.plan, nil
}
