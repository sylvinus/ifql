package execute

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

type Option func(*Options)
type Options struct {
	Verbose bool
	Trace   bool
}

func Verbose() Option {
	return func(o *Options) { o.Verbose = true }
}
func Trace() Option {
	return func(o *Options) { o.Trace = true }
}

func Execute(ctx context.Context, qSpec *query.QuerySpec, os ...Option) ([]Result, error) {
	var opts Options
	for _, o := range os {
		o(&opts)
	}
	if opts.Verbose {
		log.Println("query", query.Formatted(qSpec))
	}

	lplanner := plan.NewLogicalPlanner()
	lp, err := lplanner.Plan(qSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logical plan")
	}
	if opts.Verbose {
		log.Println("logical plan", plan.Formatted(lp))
	}

	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create physical plan")
	}
	if opts.Verbose {
		log.Println("physical plan", plan.Formatted(p))
	}

	storage, err := NewStorageReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}

	e := NewExecutor(storage, &opts)
	r, err := e.Execute(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	return r, nil
}

type Executor interface {
	Execute(context.Context, *plan.PlanSpec) ([]Result, error)
}

type executor struct {
	sr   StorageReader
	opts Options
}

func NewExecutor(sr StorageReader, opts *Options) Executor {
	e := &executor{
		sr: sr,
	}
	if opts != nil {
		e.opts = *opts
	}
	return e
}

type executionState struct {
	p    *plan.PlanSpec
	sr   StorageReader
	opts Options

	bounds Bounds

	results []Result
	runners []Runner
}

func (e *executor) Execute(ctx context.Context, p *plan.PlanSpec) ([]Result, error) {
	es, err := e.createExecutionState(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize execute state")
	}
	es.do(ctx)
	return es.results, nil
}

func (e *executor) createExecutionState(ctx context.Context, p *plan.PlanSpec) (*executionState, error) {
	es := &executionState{
		p:       p,
		sr:      e.sr,
		opts:    e.opts,
		results: make([]Result, len(p.Results)),
		bounds: Bounds{
			Start: Time(p.Bounds.Start.Time(p.Now).UnixNano()),
			Stop:  Time(p.Bounds.Stop.Time(p.Now).UnixNano()),
		},
	}
	for i, id := range p.Results {
		ds, err := es.createNode(ctx, p.Procedures[id])
		if err != nil {
			return nil, err
		}
		rs := newResultSink()
		ds.AddTransformation(rs)
		es.results[i] = rs
	}
	return es, nil
}

// DefaultTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var DefaultTriggerSpec = query.AfterWatermarkTriggerSpec{}

type triggeringSpec interface {
	TriggerSpec() query.TriggerSpec
}

func (es *executionState) createNode(ctx context.Context, pr *plan.Procedure) (Node, error) {
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
	var ts query.TriggerSpec = DefaultTriggerSpec
	if t, ok := pr.Spec.(triggeringSpec); ok {
		ts = t.TriggerSpec()
	}
	ds.SetTriggerSpec(ts)

	parentIDs := make([]DatasetID, len(pr.Parents))
	for i, parentID := range pr.Parents {
		parent, err := es.createNode(ctx, es.p.Procedures[parentID])
		if err != nil {
			return nil, err
		}
		transport := newTransformationTransport(t)
		parent.AddTransformation(transport)
		es.runners = append(es.runners, transport)
		parentIDs[i] = DatasetID(parentID)
	}
	t.SetParents(parentIDs)

	return ds, nil
}

func (es *executionState) abort(err error) {
	for _, r := range es.results {
		r.abort(err)
	}
}

type Runner interface {
	Run(ctx context.Context)
}

func (es *executionState) do(ctx context.Context) {
	for _, r := range es.runners {
		go func(r Runner) {
			defer func() {
				if e := recover(); e != nil {
					// We had a panic, abort the entire execution.
					//TODO(nathanielc): Only abort results that were effected by the panic?
					// This requires tracing the Runner through the execution DAG.
					var err error
					switch e := e.(type) {
					case error:
						err = e
					default:
						err = fmt.Errorf("%v", e)
					}
					if es.opts.Trace {
						es.abort(fmt.Errorf("panic: %v\n%s", err, debug.Stack()))
					} else {
						es.abort(errors.Wrap(err, "panic"))
					}
				}
			}()
			r.Run(ctx)
		}(r)
	}
}

// Satisfy the ExecutionContext interface

func (es *executionState) ResolveTime(qt query.Time) Time {
	return Time(qt.Time(es.p.Now).UnixNano())
}
func (es *executionState) Bounds() Bounds {
	return es.bounds
}
