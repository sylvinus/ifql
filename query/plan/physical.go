package plan

import (
	"math"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type PlanSpec struct {
	// Now represents the relative currentl time of the plan.
	Now    time.Time
	Bounds BoundsSpec
	// Procedures is a set of all operations
	Procedures map[ProcedureID]*Procedure
	Order      []ProcedureID
	// Results is a list of datasets that are the result of the plan
	Results []ProcedureID

	Resources query.ResourceManagement
}

func (p *PlanSpec) Do(f func(pr *Procedure)) {
	for _, id := range p.Order {
		f(p.Procedures[id])
	}
}
func (p *PlanSpec) lookup(id ProcedureID) *Procedure {
	return p.Procedures[id]
}

type Planner interface {
	// Plan create a plan from the logical plan and available storage.
	Plan(p *LogicalPlanSpec, s Storage, now time.Time) (*PlanSpec, error)
}

type planner struct {
	plan *PlanSpec

	modified bool
}

func NewPlanner() Planner {
	return new(planner)
}

func (p *planner) Plan(lp *LogicalPlanSpec, s Storage, now time.Time) (*PlanSpec, error) {
	p.plan = &PlanSpec{
		Now:        now,
		Procedures: make(map[ProcedureID]*Procedure, len(lp.Procedures)),
		Order:      make([]ProcedureID, 0, len(lp.Order)),
		Resources:  lp.Resources,
	}

	// Find the datasets that are results and populate mappings
	lp.Do(func(pr *Procedure) {
		p.plan.Procedures[pr.ID] = pr
		p.plan.Order = append(p.plan.Order, pr.ID)

	})

	// Find Limit+Where+Range+Select to push down time bounds and predicate
	var order []ProcedureID
	p.modified = true
	for p.modified {
		p.modified = false
		if cap(order) < len(p.plan.Order) {
			order = make([]ProcedureID, len(p.plan.Order))
		} else {
			order = order[:len(p.plan.Order)]
		}
		copy(order, p.plan.Order)
		for _, id := range order {
			pr := p.plan.Procedures[id]
			if pd, ok := pr.Spec.(PushDownProcedureSpec); ok {
				rules := pd.PushDownRules()
				for _, rule := range rules {
					if p.pushDownAndSearch(pr, rule, pd.PushDown) {
						if err := p.removeProcedure(pr); err != nil {
							return nil, errors.Wrap(err, "failed to remove procedure")
						}
					}
				}
			}
		}
	}

	// Now that plan is complete find results and time bounds
	p.plan.Do(func(pr *Procedure) {
		if bounded, ok := pr.Spec.(BoundedProcedureSpec); ok {
			bounds := bounded.TimeBounds()
			p.plan.Bounds = p.plan.Bounds.Union(bounds, now)
		}
		if len(pr.Children) == 0 {
			p.plan.Results = append(p.plan.Results, pr.ID)
		}
	})

	if p.plan.Bounds.Start.IsZero() && p.plan.Bounds.Stop.IsZero() {
		return nil, errors.New("unbounded queries are not supported. Add a '.range' call to bound the query.")
	}

	// Update concurrency quota
	if p.plan.Resources.ConcurrencyQuota == 0 {
		p.plan.Resources.ConcurrencyQuota = len(p.plan.Procedures)
	}
	// Update memory quota
	if p.plan.Resources.MemoryBytesQuota == 0 {
		p.plan.Resources.MemoryBytesQuota = math.MaxInt64
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

func (p *planner) pushDownAndSearch(pr *Procedure, rule PushDownRule, do func(parent *Procedure, dup func() *Procedure)) bool {
	matched := false
	for _, parent := range pr.Parents {
		pp := p.plan.Procedures[parent]
		pk := pp.Spec.Kind()
		if pk == rule.Root {
			if rule.Match == nil || rule.Match(pp.Spec) {
				do(pp, func() *Procedure { return p.duplicate(pp, false) })
				matched = true
			}
		} else if hasKind(pk, rule.Through) {
			p.pushDownAndSearch(pp, rule, do)
		}
	}
	return matched
}

func (p *planner) removeProcedure(pr *Procedure) error {
	// It only makes sense to remove a procedure that has a single parent.
	if len(pr.Parents) > 1 {
		return errors.New("cannot remove a procedure that has more than one parent")
	}

	p.modified = true
	delete(p.plan.Procedures, pr.ID)
	p.plan.Order = removeID(p.plan.Order, pr.ID)

	for _, id := range pr.Parents {
		parent := p.plan.Procedures[id]
		parent.Children = removeID(parent.Children, pr.ID)
		parent.Children = append(parent.Children, pr.Children...)
	}
	for _, id := range pr.Children {
		child := p.plan.Procedures[id]
		child.Parents = removeID(child.Parents, pr.ID)
		child.Parents = append(child.Parents, pr.Parents...)

		if len(pr.Parents) == 1 {
			if pa, ok := child.Spec.(ParentAwareProcedureSpec); ok {
				pa.ParentChanged(pr.ID, pr.Parents[0])
			}
		}
	}
	return nil
}

func ProcedureIDForDuplicate(id ProcedureID) ProcedureID {
	return ProcedureID(uuid.NewV5(RootUUID, id.String()))
}

func (p *planner) duplicate(pr *Procedure, skipParents bool) *Procedure {
	p.modified = true
	np := pr.Copy()
	np.ID = ProcedureIDForDuplicate(pr.ID)
	p.plan.Procedures[np.ID] = np
	p.plan.Order = insertAfter(p.plan.Order, pr.ID, np.ID)

	if !skipParents {
		for _, id := range np.Parents {
			parent := p.plan.Procedures[id]
			parent.Children = append(parent.Children, np.ID)
		}
	}

	newChildren := make([]ProcedureID, len(np.Children))
	for i, id := range np.Children {
		child := p.plan.Procedures[id]
		newChild := p.duplicate(child, true)
		newChild.Parents = removeID(newChild.Parents, pr.ID)
		newChild.Parents = append(newChild.Parents, np.ID)

		newChildren[i] = newChild.ID

		if pa, ok := newChild.Spec.(ParentAwareProcedureSpec); ok {
			pa.ParentChanged(pr.ID, np.ID)
		}
	}
	np.Children = newChildren
	return np
}

func removeID(ids []ProcedureID, remove ProcedureID) []ProcedureID {
	filtered := ids[0:0]
	for i, id := range ids {
		if id == remove {
			filtered = append(filtered, ids[0:i]...)
			filtered = append(filtered, ids[i+1:]...)
			break
		}
	}
	return filtered
}
func insertAfter(ids []ProcedureID, after, new ProcedureID) []ProcedureID {
	var newIds []ProcedureID
	for i, id := range ids {
		if id == after {
			newIds = append(newIds, ids[:i+1]...)
			newIds = append(newIds, new)
			if i+1 < len(ids) {
				newIds = append(newIds, ids[i+1:]...)
			}
			break
		}
	}
	return newIds
}
