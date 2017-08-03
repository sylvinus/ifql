package plan

import (
	"fmt"

	"github.com/influxdata/ifql/query"
)

type AbstractPlan interface {
	Operations() []AbstractOperation
	Datasets() []AbstractDataset
}

type AbstractDataset interface {
	// Bounded reports whether the dataset is bounded.
	Bounded() bool
	Bounds() Bounds
	Windowed() bool
	Window() Window
	setBounds(Bounds)
	Source() AbstractSource
	setSource(AbstractSource)

	MakeNarrowChild() AbstractDataset
}

type absDataset struct {
	source AbstractSource
	bounds Bounds
}

func (d *absDataset) Bounded() bool {
	return d.bounds != nil
}
func (d *absDataset) Bounds() Bounds {
	return d.bounds
}
func (d *absDataset) setBounds(b Bounds) {
	d.bounds = b
}
func (d *absDataset) Source() AbstractSource {
	return d.source
}
func (d *absDataset) setSource(s AbstractSource) {
	d.source = s
}

func (d *absDataset) MakeNarrowChild() AbstractDataset {
	c := &absDataset{
		bounds: d.bounds,
	}
	return c
}

type Bounds interface {
	Start() query.Time
	Stop() query.Time
}

type bounds struct {
	start query.Time
	stop  query.Time
}

func (b *bounds) Start() query.Time {
	return b.start
}
func (b *bounds) Stop() query.Time {
	return b.stop
}

type Window interface {
	Every() query.Duration
	Period() query.Duration
	Round() query.Duration
	Start() query.Time
}

type window struct {
	every  query.Duration
	period query.Duration
	round  query.Duration
	start  query.Time
}

func (w *window) Every() query.Duration {
	return w.every
}
func (w *window) Period() query.Duration {
	return w.period
}
func (w *window) Round() query.Duration {
	return w.round
}
func (w *window) Start() query.Time {
	return w.start
}

type absPlan struct {
	operations []AbstractOperation
	datasets   []AbstractDataset
}

func (p *absPlan) Operations() []AbstractOperation {
	return p.operations
}
func (p *absPlan) Datasets() []AbstractDataset {
	return p.datasets
}

type AbstractSource interface{}

type AbstractPlanner interface {
	Plan(*query.QuerySpec) (AbstractPlan, error)
}

type absPlanner struct {
	plan          *absPlan
	q             *query.QuerySpec
	datasetLookup map[query.OperationID][]AbstractDataset
}

func NewAbstractPlanner() AbstractPlanner {
	return new(absPlanner)
}

func (p *absPlanner) Plan(q *query.QuerySpec) (AbstractPlan, error) {
	p.q = q
	p.plan = new(absPlan)
	p.datasetLookup = make(map[query.OperationID][]AbstractDataset)
	err := q.Walk(p.walk)
	if err != nil {
		return nil, err
	}
	return p.plan, nil
}

func (p *absPlanner) walk(o *query.Operation) error {
	switch spec := o.Spec.(type) {
	case *query.SelectOpSpec:
		return p.doSelect(o, spec)
	case *query.RangeOpSpec:
		return p.doRange(o, spec)
	case *query.ClearOpSpec:
		return p.doClear(o, spec)
	default:
		return fmt.Errorf("unsupported query operation %v", o.Spec.Kind())
	}
}

func (p *absPlanner) doSelect(o *query.Operation, spec *query.SelectOpSpec) error {
	ds := new(absDataset)
	ds.setSource(o)
	p.plan.datasets = append(p.plan.datasets, ds)
	p.datasetLookup[o.ID] = append(p.datasetLookup[o.ID], ds)
	return nil
}

func (p *absPlanner) doRange(o *query.Operation, spec *query.RangeOpSpec) error {
	b := &bounds{
		start: spec.Start,
		stop:  spec.Stop,
	}

	parents := p.q.Parents(o.ID)
	for _, parent := range parents {
		datasets := p.datasetLookup[parent.ID]
		for _, ds := range datasets {
			ds.setBounds(b)
		}
		// Re-index the parent datasets under the range operation ID
		p.datasetLookup[o.ID] = append(p.datasetLookup[o.ID], datasets)
	}
	return nil
}

func (p *absPlanner) doClear(o *query.Operation, spec *query.ClearOpSpec) error {
	parents := p.q.Parents(o.ID)
	for _, parent := range parents {
		parentDS := p.datasetLookup[parent.ID]
		childDS := make([]AbstractDataset, len(parentDS))
		for i := range parentDS {
			childDS[i] = parentDS[i].MakeNarrowChild()
		}

		co := &clearOperation{
			operation: operation{
				parents:   parentDS,
				children:  childDS,
				operation: o,
			},
		}
		p.operations = append(p.operations, co)
		p.datasetLookup[o.ID] = append(p.datasetLookup[o.ID], childDS)
	}
	return nil
}
