package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

func Execute(qSpec *query.QuerySpec) ([]Result, error) {
	lplanner := plan.NewLogicalPlanner()
	lp, err := lplanner.Plan(qSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logical plan")
	}
	//log.Println("logical plan", plan.Formatted(lp))

	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create physical plan")
	}
	//log.Println("plan", plan.Formatted(p))

	storage, err := NewStorageReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}

	e := NewExecutor(storage)
	r, err := e.Execute(context.Background(), p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	return r, nil
}

type Executor interface {
	Execute(context.Context, *plan.PlanSpec) ([]Result, error)
}

type executor struct {
	sr StorageReader
}

func NewExecutor(sr StorageReader) Executor {
	return &executor{
		sr: sr,
	}
}

type executionState struct {
	p  *plan.PlanSpec
	sr StorageReader

	bounds Bounds

	results []Result
	runners []Runner
}

func (e *executor) Execute(ctx context.Context, p *plan.PlanSpec) ([]Result, error) {
	es, err := e.createExecutionState(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize execute state")
	}
	es.do(ctx)
	return es.results, nil
}

func (e *executor) createExecutionState(p *plan.PlanSpec) (*executionState, error) {
	es := &executionState{
		p:       p,
		sr:      e.sr,
		results: make([]Result, len(p.Results)),
		bounds: Bounds{
			Start: Time(p.Bounds.Start.Time(p.Now).UnixNano()),
			Stop:  Time(p.Bounds.Stop.Time(p.Now).UnixNano()),
		},
	}
	for i, id := range p.Results {
		ds, err := es.createNode(p.Procedures[id])
		if err != nil {
			return nil, err
		}
		rs := newResultSink()
		ds.addTransformation(rs)
		es.results[i] = rs
	}
	return es, nil
}

// defaultTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var defaultTriggerSpec = query.AfterWatermarkTriggerSpec{}

type triggeringSpec interface {
	TriggerSpec() query.TriggerSpec
}

func (es *executionState) createNode(pr *plan.Procedure) (Node, error) {
	if createS, ok := procedureToSource[pr.Spec.Kind()]; ok {
		s := createS(pr.Spec, DatasetID(pr.ID), es.sr, es)
		es.runners = append(es.runners, s)
		return s, nil
	}

	createT, ok := procedureToTransformation[pr.Spec.Kind()]

	if !ok {
		return nil, fmt.Errorf("unsupported procedure %v", pr.Spec.Kind())
	}
	t, ds, err := createT(DatasetID(pr.ID), AccumulatingMode, pr.Spec, es)
	if err != nil {
		return nil, err
	}

	// Setup triggering
	var ts query.TriggerSpec = defaultTriggerSpec
	if t, ok := pr.Spec.(triggeringSpec); ok {
		ts = t.TriggerSpec()
	}
	ds.setTriggerSpec(ts)

	parentIDs := make([]DatasetID, len(pr.Parents))
	for i, parentID := range pr.Parents {
		parent, err := es.createNode(es.p.Procedures[parentID])
		if err != nil {
			return nil, err
		}
		transport := newTransformationTransport(t, nil)
		parent.addTransformation(transport)
		es.runners = append(es.runners, transport)
		parentIDs[i] = DatasetID(parentID)
	}
	t.SetParents(parentIDs)

	return ds, nil
}

func (es *executionState) do(ctx context.Context) {
	// TODO: plumb context.Context through the Runnables
	for _, r := range es.runners {
		r := r
		go r.Run()
	}
}

// Satisfy the ExecutionContext interface

func (es *executionState) ResolveTime(qt query.Time) Time {
	return Time(qt.Time(es.p.Now).UnixNano())
}
func (es *executionState) Bounds() Bounds {
	return es.bounds
}

type Runner interface {
	Run()
}
