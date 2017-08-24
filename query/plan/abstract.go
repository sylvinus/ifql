package plan

import (
	"fmt"
	"reflect"
	"strconv"

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

	parents := p.q.Parents(o.ID)

	switch spec := spec.(type) {
	case NarrowProcedureSpec:
		p.doNarrow(parents, pr, spec)
	case WideProcedureSpec:
		p.doWide(pr, spec)
	default:
		return fmt.Errorf("operation must be be either narrow or wide: %v", spec.Kind())
	}
	return nil
}

func (p *abstractPlanner) createSpec(qk query.OperationKind) ProcedureSpec {
	k := opToProcedureKind[qk]
	typ := kindToGoType[k]
	return reflect.New(typ).Interface().(ProcedureSpec)
}

func (p *abstractPlanner) doNarrow(parents []*query.Operation, o *Procedure, spec NarrowProcedureSpec) {
	for _, parent := range parents {
		parentOpID := ProcedureIDFromOperationID(parent.ID)
		parentOp := p.procedureLookup[parentOpID]
		parentDatasetIDs := parentOp.Children
		childrenDatasets := make([]*Dataset, len(parentDatasetIDs))
		childrenDatasetIDs := make([]DatasetID, len(parentDatasetIDs))
		for i, pid := range parentDatasetIDs {
			parentDataset := p.datasetLookup[pid]
			childDataset := parentDataset.MakeNarrowChild(o.ID, strconv.Itoa(i))
			childDataset.Source = o.ID
			spec.NewChild(childDataset)
			childrenDatasets[i] = childDataset
			childrenDatasetIDs[i] = childDataset.ID
			p.datasetLookup[childDataset.ID] = childDataset
		}
		p.plan.Datasets = append(p.plan.Datasets, childrenDatasets...)
		o.Children = append(o.Children, childrenDatasetIDs...)
		o.Parents = append(o.Parents, parentDatasetIDs...)
	}
}

func CreateDatasetID(pid ProcedureID, name string) DatasetID {
	return DatasetID(uuid.NewV5(uuid.UUID(pid), name))
}

func (p *abstractPlanner) doWide(o *Procedure, spec WideProcedureSpec) {
	children := spec.DetermineChildren()
	childIDs := make([]DatasetID, len(children))

	for i, c := range children {
		// TODO figure out how the dataset ID hierarchy should work
		// This currently can create duplicate IDs
		c.ID = CreateDatasetID(o.ID, strconv.Itoa(i))
		c.Source = o.ID
		childIDs[i] = c.ID
		p.datasetLookup[c.ID] = c
	}

	p.plan.Datasets = append(p.plan.Datasets, children...)
	o.Children = childIDs
}
