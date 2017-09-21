package plan

import (
	"reflect"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

var NilUUID uuid.UUID
var RootUUID = NilUUID

type LogicalPlanSpec struct {
	Procedures map[ProcedureID]*Procedure
}

type LogicalPlanner interface {
	Plan(*query.QuerySpec) (*LogicalPlanSpec, error)
}

type logicalPlanner struct {
	plan *LogicalPlanSpec
	q    *query.QuerySpec
}

func NewLogicalPlanner() LogicalPlanner {
	return new(logicalPlanner)
}

func (p *logicalPlanner) Plan(q *query.QuerySpec) (*LogicalPlanSpec, error) {
	p.q = q
	p.plan = &LogicalPlanSpec{
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

func (p *logicalPlanner) walkQuery(o *query.Operation) error {
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

func (p *logicalPlanner) createSpec(qk query.OperationKind) ProcedureSpec {
	k := opToProcedureKind[qk]
	typ := kindToGoType[k]
	return reflect.New(typ).Interface().(ProcedureSpec)
}
