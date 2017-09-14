package plan

import (
	"time"
)

type PlanSpec struct {
	// Now represents the relative currentl time of the plan.
	Now time.Time
	// Procedures is a set of all operations
	Procedures map[ProcedureID]*Procedure
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
		Procedures: make(map[ProcedureID]*Procedure, len(ap.Procedures)),
		Datasets:   make(map[DatasetID]*Dataset, len(ap.Datasets)),
	}

	// Find the datasets that are results and populate mappings
	childCount := make(map[DatasetID]int)
	for _, pr := range ap.Procedures {
		p.plan.Procedures[pr.ID] = pr

		for _, d := range pr.Parents {
			childCount[d]++
		}
	}
	for _, d := range ap.Datasets {
		p.plan.Datasets[d.ID] = d

		if childCount[d.ID] == 0 {
			p.plan.Results = append(p.plan.Results, d.ID)
		}
	}

	// Find Limit+Where+Range+Select to push down time bounds and predicate
	for _, pr := range ap.Procedures {
		switch spec := pr.Spec.(type) {
		case *RangeProcedureSpec:
			p.pushDownRange(pr, spec)
		case *WhereProcedureSpec:
			p.pushDownWhere(pr, spec)
		case *LimitProcedureSpec:
			p.pushDownLimit(pr, spec)
		}
	}

	return p.plan, nil
}

func hasKind(kind ProcedureKind, kinds []ProcedureKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func (p *planner) pushDownAndSearch(pr *Procedure, kind ProcedureKind, do func(parent *Procedure), validPushThroughKinds ...ProcedureKind) {
	for _, parent := range pr.Parents {
		pp := p.plan.Procedures[p.plan.Datasets[parent].Source]
		pk := pp.Spec.Kind()
		if pk == kind {
			do(pp)
		} else if hasKind(pk, validPushThroughKinds) {
			p.pushDownAndSearch(pp, kind, do)
		} else {
			// Cannot push down
			// TODO: create new branch since procedure cannot be pushed down
		}
	}
	// Remove procedure since it has been pushed down
	p.removeProcedure(pr)
}

func (p *planner) removeProcedure(pr *Procedure) {
	delete(p.plan.Procedures, pr.ID)

	childDS := p.plan.Datasets[pr.Child]
	delete(p.plan.Datasets, pr.Child)

	for _, dest := range childDS.Destinations {
		p.plan.Procedures[dest].Parents = pr.Parents
	}
	for _, parentDS := range pr.Parents {
		p.plan.Datasets[parentDS].Destinations = childDS.Destinations
	}
}

func (p *planner) pushDownRange(pr *Procedure, spec *RangeProcedureSpec) {
	p.pushDownAndSearch(pr, SelectKind, func(parent *Procedure) {
		selectSpec := parent.Spec.(*SelectProcedureSpec)
		if selectSpec.BoundsSet {
			// TODO: create copy of select spec and set new bounds
			//
			// Example case where this matters
			//    var data = select(database: "mydb")
			//    var past = data.range(start:-2d,stop:-1d)
			//    var current = data.range(start:-1d,stop:now)
		}
		selectSpec.BoundsSet = true
		selectSpec.Bounds = spec.Bounds
		// Update child dataset with bounds
		p.plan.Datasets[parent.Child].Bounds = spec.Bounds
	},
		WhereKind, LimitKind,
	)
}

func (p *planner) pushDownWhere(pr *Procedure, spec *WhereProcedureSpec) {
	p.pushDownAndSearch(pr, SelectKind, func(parent *Procedure) {
		selectSpec := parent.Spec.(*SelectProcedureSpec)
		if selectSpec.WhereSet {
			// TODO: create copy of select spec and set new where expression
		}
		selectSpec.WhereSet = true
		selectSpec.Where = spec.Exp.Exp.Predicate
	},
		LimitKind, RangeKind,
	)
}

func (p *planner) pushDownLimit(pr *Procedure, spec *LimitProcedureSpec) {
	p.pushDownAndSearch(pr, SelectKind, func(parent *Procedure) {
		selectSpec := parent.Spec.(*SelectProcedureSpec)
		if selectSpec.LimitSet {
			// TODO: create copy of select spec and set new limit
		}
		selectSpec.LimitSet = true
		selectSpec.Limit = spec.Limit
		selectSpec.Offset = spec.Offset
	},
		WhereKind, RangeKind,
	)
}
