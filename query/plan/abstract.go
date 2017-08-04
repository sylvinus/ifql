package plan

import (
	"fmt"
	"reflect"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

var NilUUID uuid.UUID
var RootUUID = NilUUID

type AbstractPlanSpec struct {
	Operations []*Operation
	Datasets   []*Dataset
}

type AbstractPlanner interface {
	Plan(*query.QuerySpec) (*AbstractPlanSpec, error)
}

type absPlanner struct {
	plan            *AbstractPlanSpec
	q               *query.QuerySpec
	operationLookup map[OperationID]*Operation
	datasetLookup   map[DatasetID]*Dataset
}

func NewAbstractPlanner() AbstractPlanner {
	return new(absPlanner)
}

func (p *absPlanner) Plan(q *query.QuerySpec) (*AbstractPlanSpec, error) {
	p.q = q
	p.plan = new(AbstractPlanSpec)
	p.operationLookup = make(map[OperationID]*Operation)
	p.datasetLookup = make(map[DatasetID]*Dataset)
	err := q.Walk(p.walkQuery)
	if err != nil {
		return nil, err
	}
	return p.plan, nil
}

func OpIDFromQueryOpID(id query.OperationID) OperationID {
	return OperationID(uuid.NewV5(RootUUID, string(id)))
}

func (p *absPlanner) walkQuery(o *query.Operation) error {
	opSpec := p.createSpec(o.Spec.Kind())
	if err := opSpec.SetSpec(o.Spec); err != nil {
		return err
	}

	op := &Operation{
		ID:   OpIDFromQueryOpID(o.ID),
		Spec: opSpec,
	}
	p.operationLookup[op.ID] = op
	p.plan.Operations = append(p.plan.Operations, op)

	parents := p.q.Parents(o.ID)
	switch spec := opSpec.(type) {
	case NarrowOperationSpec:
		p.doNarrow(parents, op, spec)
	case WideOperationSpec:
		p.doWide(op, spec)
	default:
		return fmt.Errorf("operation must be be either narrow or wide: %v", spec.Kind())
	}
	return nil
}

func (p *absPlanner) createSpec(qk query.OperationKind) OperationSpec {
	k := queryOpToOpKind[qk]
	typ := kindToGoType[k]
	return reflect.New(typ).Interface().(OperationSpec)
}

func (p *absPlanner) doNarrow(parents []*query.Operation, o *Operation, spec NarrowOperationSpec) {
	for _, parent := range parents {
		parentOpID := OpIDFromQueryOpID(parent.ID)
		parentOp := p.operationLookup[parentOpID]
		parentDatasetIDs := parentOp.Children
		childrenDatasets := make([]*Dataset, len(parentDatasetIDs))
		childrenDatasetIDs := make([]DatasetID, len(parentDatasetIDs))
		for i, pid := range parentDatasetIDs {
			parentDataset := p.datasetLookup[pid]
			childDataset := parentDataset.MakeNarrowChild(o.ID)
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

func CreateDatasetID(root DatasetID, oid OperationID) DatasetID {
	return DatasetID(uuid.NewV5(uuid.UUID(root), oid.String()))
}

func (p *absPlanner) doWide(o *Operation, spec WideOperationSpec) {
	children := spec.DetermineChildren()
	childIDs := make([]DatasetID, len(children))

	for i, c := range children {
		c.ID = CreateDatasetID(InvalidDatasetID, o.ID)
		c.Source = o.ID
		childIDs[i] = c.ID
		p.datasetLookup[c.ID] = c
	}

	p.plan.Datasets = append(p.plan.Datasets, children...)
	o.Children = childIDs
}
