package plan

import (
	"reflect"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

var NilUUID uuid.UUID
var RootUUID = NilUUID

type AbstractPlanSpec struct {
	Procedures []*Procedure
	Datasets   []*Dataset
}

type AbstractPlanner interface {
	Plan(*query.QuerySpec) (*AbstractPlanSpec, error)
}

type abstractPlanner struct {
	plan            *AbstractPlanSpec
	q               *query.QuerySpec
	procedureLookup map[ProcedureID]*Procedure
	datasetLookup   map[DatasetID]*Dataset
}

func NewAbstractPlanner() AbstractPlanner {
	return new(abstractPlanner)
}

func (p *abstractPlanner) Plan(q *query.QuerySpec) (*AbstractPlanSpec, error) {
	p.q = q
	p.plan = new(AbstractPlanSpec)
	p.procedureLookup = make(map[ProcedureID]*Procedure)
	p.datasetLookup = make(map[DatasetID]*Dataset)
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
	p.procedureLookup[pr.ID] = pr
	p.plan.Procedures = append(p.plan.Procedures, pr)

	// Create child dataset
	childDataset := &Dataset{
		ID:     CreateDatasetID(pr.ID),
		Source: pr.ID,
	}
	pr.Child = childDataset.ID
	p.datasetLookup[childDataset.ID] = childDataset
	p.plan.Datasets = append(p.plan.Datasets, childDataset)

	// Link parents destination procedures
	parentOps := p.q.Parents(o.ID)
	for _, parentOp := range parentOps {
		parentPID := ProcedureIDFromOperationID(parentOp.ID)
		parentP := p.procedureLookup[parentPID]
		parentDS := p.datasetLookup[parentP.Child]
		parentDS.Destinations = append(parentDS.Destinations, pr.ID)
		pr.Parents = append(pr.Parents, parentDS.ID)
	}

	return nil
}

func (p *abstractPlanner) createSpec(qk query.OperationKind) ProcedureSpec {
	k := opToProcedureKind[qk]
	typ := kindToGoType[k]
	return reflect.New(typ).Interface().(ProcedureSpec)
}

func CreateDatasetID(pid ProcedureID) DatasetID {
	return DatasetID(pid)
}
