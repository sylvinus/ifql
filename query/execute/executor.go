package execute

import (
	"context"
	"fmt"
	"sync"
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

	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create physical plan")
	}

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

	results []Result
	sources []Source
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
		s := createS(pr.Spec, DatasetID(pr.ID), es.sr, es.p.Now)
		es.sources = append(es.sources, s)
		return s, nil
	}

	createT, ok := procedureToTransformation[pr.Spec.Kind()]

	if !ok {
		return nil, fmt.Errorf("unsupported procedure %v", pr.Spec.Kind())
	}
	t, ds, err := createT(DatasetID(pr.ID), AccumulatingMode, pr.Spec, es.p.Now)
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
		parent.addTransformation(t)
		parentIDs[i] = DatasetID(parentID)
	}
	t.SetParents(parentIDs)

	return ds, nil
}

func (es *executionState) do(ctx context.Context) {
	//TODO: pass through the context and design a concurrency system that works for any DAG,
	// this current implementation only works for linear DAGs.
	wg := new(sync.WaitGroup)
	wg.Add(len(es.sources))
	for _, s := range es.sources {
		s := s
		go func() {
			defer wg.Done()
			s.Run()
		}()
	}
	wg.Wait()
}
