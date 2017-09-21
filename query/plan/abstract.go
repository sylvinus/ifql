package plan

import (
	"reflect"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

var NilUUID uuid.UUID
var RootUUID = NilUUID

type AbstractPlanSpec struct {
	Procedures map[ProcedureID]*Procedure
}

type AbstractPlanner interface {
	Plan(*query.QuerySpec) (*AbstractPlanSpec, error)
}

type abstractPlanner struct {
	plan *AbstractPlanSpec
	q    *query.QuerySpec
}

func NewAbstractPlanner() AbstractPlanner {
	return new(abstractPlanner)
}

func (p *abstractPlanner) Plan(q *query.QuerySpec) (*AbstractPlanSpec, error) {
	p.q = q
	p.plan = &AbstractPlanSpec{
		Procedures: make(map[ProcedureID]*Procedure),
	}
	err := q.Walk(p.walkQuery)
	if err != nil {
		return nil, err
	}
	return p.plan, nil
}

func ProcedureIDFromOperationID(id query.OperationID) ProcedureID {
	return ProcedureID(uuid.NewV5(RootUUID, string(id)))
}

func (p *abstractPlanner) walkQuery(o *query.Operation) error {
	spec := p.createSpec(o.Spec.Kind())
	if err := spec.SetSpec(o.Spec); err != nil {
		return err
	}

	pr := &Procedure{
		ID:   ProcedureIDFromOperationID(o.ID),
		Spec: spec,
	}
	p.plan.Procedures[pr.ID] = pr

	// Link parent/child relations
	parentOps := p.q.Parents(o.ID)
	for _, parentOp := range parentOps {
		parentID := ProcedureIDFromOperationID(parentOp.ID)
		parentPr := p.plan.Procedures[parentID]
		parentPr.Children = append(parentPr.Children, pr.ID)
		pr.Parents = append(pr.Parents, parentID)
	}

	return nil
}

func (p *abstractPlanner) createSpec(qk query.OperationKind) ProcedureSpec {
	k := opToProcedureKind[qk]
	typ := kindToGoType[k]
	return reflect.New(typ).Interface().(ProcedureSpec)
}
