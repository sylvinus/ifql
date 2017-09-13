package execute

import (
	"context"
	"sync"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

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
		return nil, err
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
	for i, r := range p.Results {
		ds := es.createNode(p.Datasets[r])
		rs := newResultSink()
		ds.addTransformation(rs)
		es.results[i] = rs
	}
	return es, nil
}

// nonWindowTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var nonWindowTriggerSpec = query.AfterWatermarkTriggerSpec{}

func (es *executionState) createNode(d *plan.Dataset) Node {
	src := es.p.Procedures[d.Source]
	if src.Spec.Kind() == plan.SelectKind {
		spec := src.Spec.(*plan.SelectProcedureSpec)
		s := newStorageSource(DatasetID(d.ID), es.sr, spec, es.p.Now)
		es.sources = append(es.sources, s)
		return s
	}

	ds := newDataset(DatasetID(d.ID), AccumulatingMode)

	// Setup triggering
	if src.Spec.Kind() == plan.WindowKind {
		w := src.Spec.(*plan.WindowProcedureSpec)
		triggerSpec := w.Triggering
		ds.setTriggerSpec(triggerSpec)
	} else {
		ds.setTriggerSpec(nonWindowTriggerSpec)
	}

	t := transformationFromProcedureSpec(ds, src.Spec, es.p.Now)
	for _, parentDS := range src.Parents {
		parent := es.createNode(es.p.Datasets[parentDS])
		parent.addTransformation(t)
	}

	return ds
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
