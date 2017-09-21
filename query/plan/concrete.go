package plan

import (
	"time"
)

type PlanSpec struct {
	// Now represents the relative currentl time of the plan.
	Now time.Time
	// Procedures is a set of all operations
	Procedures map[ProcedureID]*Procedure
	// Results is a list of datasets that are the result of the plan
	Results []ProcedureID
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
	}

	// Find the datasets that are results and populate mappings
	for id, pr := range ap.Procedures {
		p.plan.Procedures[id] = pr

		if len(pr.Children) == 0 {
			p.plan.Results = append(p.plan.Results, id)
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
		pp := p.plan.Procedures[parent]
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
}

func (p *planner) removeProcedure(pr *Procedure) {
	delete(p.plan.Procedures, pr.ID)

	for _, id := range pr.Parents {
		parent := p.plan.Procedures[id]
		parent.Children = removeID(parent.Children, pr.ID)
		parent.Children = append(parent.Children, pr.Children...)
	}
	for _, id := range pr.Children {
		child := p.plan.Procedures[id]
		child.Parents = removeID(child.Parents, pr.ID)
		child.Parents = append(child.Parents, pr.Parents...)
	}
}

func removeID(ids []ProcedureID, remove ProcedureID) []ProcedureID {
	filtered := ids[0:0]
	for _, id := range ids {
		if id != remove {
			filtered = append(filtered, id)
		}
	}
	return filtered
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
	},
		WhereKind, LimitKind,
	)
	// Remove procedure since it has been pushed down
	p.removeProcedure(pr)
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
	// Remove procedure since it has been pushed down
	p.removeProcedure(pr)
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
	// Remove procedure since it has been pushed down
	p.removeProcedure(pr)
}
