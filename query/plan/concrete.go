package plan

import "time"

type PlanSpec struct {
	// Now represents the relative currentl time of the plan.
	Now time.Time
	// Operations is a set of all operations
	Operations map[OperationID]*Operation
	// Datasets is a set of all datasets
	Datasets map[DatasetID]*Dataset
	// Results is a list of datasets that are the result of the plan
	Results []DatasetID
}

type Planner interface {
	// Plan create a plan from the abstract plan and available storage.
	Plan(p *AbstractPlanSpec, s Storage, now time.Time) (*PlanSpec, error)
}

type planner struct {
	plan *PlanSpec
}

func NewPlanner() Planner {
	return new(planner)
}

func (p *planner) Plan(ap *AbstractPlanSpec, s Storage, now time.Time) (*PlanSpec, error) {
	p.plan = &PlanSpec{
		Now:        now,
		Operations: make(map[OperationID]*Operation, len(ap.Operations)),
		Datasets:   make(map[DatasetID]*Dataset, len(ap.Datasets)),
	}

	// Find the datasets that are results and populate mappings
	childCount := make(map[DatasetID]int)
	for _, o := range ap.Operations {
		p.plan.Operations[o.ID] = o

		for _, d := range o.Parents {
			childCount[d]++
		}
	}
	for _, d := range ap.Datasets {
		p.plan.Datasets[d.ID] = d

		if childCount[d.ID] == 0 {
			p.plan.Results = append(p.plan.Results, d.ID)
		}
	}

	return p.plan, nil
}
