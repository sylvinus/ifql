package execute

import (
	"context"
	"log"

	"github.com/influxdata/ifql/query/plan"
)

type Executor interface {
	Execute(context.Context, *plan.PlanSpec) ([]DataFrame, error)
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
	// list of result datasets
	results []Dataset

	// list of actual result data frames
	resultFrames []DataFrame
}

func (e *executor) Execute(ctx context.Context, p *plan.PlanSpec) ([]DataFrame, error) {
	es, err := e.createExecutionState(p)
	if err != nil {
		return nil, err
	}
	es.do(ctx)
	return es.resultFrames, nil
}

func (e *executor) createExecutionState(p *plan.PlanSpec) (*executionState, error) {
	es := &executionState{
		p:       p,
		sr:      e.sr,
		results: make([]Dataset, len(p.Results)),
	}
	for i, r := range p.Results {
		ds := es.createDataset(p.Datasets[r])
		es.results[i] = ds
	}
	return es, nil
}

func (es *executionState) createDataset(d *plan.Dataset) Dataset {
	src := es.p.Operations[d.Source]
	if src.Spec.Kind() == plan.SelectKind {
		return &readDataset{
			reader: es.sr,
			spec:   src.Spec.(*plan.SelectOpSpec),
		}
	}
	ds := new(dataset)
	ds.op = operationFromSpec(src.Spec, es.p.Now)
	// TODO implement more than one parent
	ds.parent = es.createDataset(es.p.Datasets[src.Parents[0]]).Frames()
	return ds
}

func (es *executionState) do(ctx context.Context) {
	for _, r := range es.results {
		frames := r.Frames()
		log.Println("do", frames)
		for f, ok := frames.NextFrame(); ok; f, ok = frames.NextFrame() {
			log.Println("frame", f)
			es.resultFrames = append(es.resultFrames, f)
		}
	}
}
