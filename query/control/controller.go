package control

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type Controller struct {
	newQueries    chan *Query
	lastID        QueryID
	queriesMu     sync.RWMutex
	queries       map[QueryID]*Query
	queryDone     chan *Query
	cancelRequest chan QueryID

	lplanner plan.LogicalPlanner
	pplanner plan.Planner
	executor execute.Executor

	availableConcurrency int
	availableMemory      int64
}

type Config struct {
	ConcurrencyQuota int
	MemoryBytesQuota int64
	ExecutorConfig   execute.Config
}

type QueryID uint64

func New(c Config) *Controller {
	ctrl := &Controller{
		newQueries:           make(chan *Query),
		queries:              make(map[QueryID]*Query),
		queryDone:            make(chan *Query),
		cancelRequest:        make(chan QueryID),
		availableConcurrency: c.ConcurrencyQuota,
		availableMemory:      c.MemoryBytesQuota,
		lplanner:             plan.NewLogicalPlanner(),
		pplanner:             plan.NewPlanner(),
		executor:             execute.NewExecutor(c.ExecutorConfig),
	}
	go ctrl.run()
	return ctrl
}

// Query submits a query for execution returning immediately.
// The spec must not be modified while the query is still active.
func (c *Controller) Query(ctx context.Context, qSpec *query.QuerySpec) (*Query, error) {
	if err := qSpec.Validate(); err != nil {
		return nil, err
	}
	id := c.nextID()
	cctx, cancel := context.WithCancel(ctx)
	ready := make(chan []execute.Result, 1)
	q := &Query{
		id:     id,
		c:      c,
		Spec:   *qSpec,
		now:    time.Now().UTC(),
		ready:  ready,
		Ready:  ready,
		state:  Queueing,
		ctx:    cctx,
		cancel: cancel,
	}
	q.queueSpan, q.ctx = StartSpanFromContext(q.ctx, "queueing")
	queueingGauge.Inc()

	// Add query to the queue
	c.newQueries <- q
	return q, nil
}

func (c *Controller) nextID() QueryID {
	c.queriesMu.RLock()
	defer c.queriesMu.RUnlock()
	ok := true
	for ok {
		c.lastID++
		_, ok = c.queries[c.lastID]
	}
	return c.lastID
}

func (c *Controller) Queries() []*Query {
	c.queriesMu.RLock()
	defer c.queriesMu.RUnlock()
	queries := make([]*Query, 0, len(c.queries))
	for _, q := range c.queries {
		queries = append(queries, q)
	}
	return queries
}

func (c *Controller) run() {
	pq := newPriorityQueue()
	for {
		select {
		// Wait for resources to free
		case q := <-c.queryDone:
			c.free(q)
			c.queriesMu.Lock()
			delete(c.queries, q.id)
			c.queriesMu.Unlock()
		// Wait for new queries
		case q := <-c.newQueries:
			pq.Push(q)
			c.queriesMu.Lock()
			c.queries[q.id] = q
			c.queriesMu.Unlock()
		// Wait for cancel query requests
		case id := <-c.cancelRequest:
			c.queriesMu.RLock()
			q := c.queries[id]
			c.queriesMu.RUnlock()
			q.Cancel()
		}

		// Peek at head of priority queue
		q := pq.Peek()
		if q != nil {
			if q.tryPlan() {
				// Plan query to determine needed resources
				lp, err := c.lplanner.Plan(&q.Spec)
				if err != nil {
					q.setErr(errors.Wrap(err, "failed to create logical plan"))
					continue
				}

				p, err := c.pplanner.Plan(lp, nil, q.now)
				if err != nil {
					q.setErr(errors.Wrap(err, "failed to create physical plan"))
					continue
				}
				q.plan = p
				q.concurrency = p.Resources.ConcurrencyQuota
				q.memory = p.Resources.MemoryBytesQuota
			}

			// Check if we have enough resources
			if c.check(q) {
				// Update resource gauges
				c.consume(q)

				// Remove the query from the queue
				pq.Pop()

				// Execute query
				if q.tryExec() {
					r, err := c.executor.Execute(q.ctx, q.plan)
					if err != nil {
						q.setErr(errors.Wrap(err, "failed to execute query"))
						continue
					}
					q.setResults(r)
				}
			} else {
				// update state to queueing
				q.tryRequeue()
			}
		}
	}
}

func (c *Controller) check(q *Query) bool {
	return c.availableConcurrency >= q.concurrency && (q.memory == math.MaxInt64 || c.availableMemory >= q.memory)
}
func (c *Controller) consume(q *Query) {
	c.availableConcurrency -= q.concurrency

	if q.memory != math.MaxInt64 {
		c.availableMemory -= q.memory
	}
}

func (c *Controller) free(q *Query) {
	c.availableConcurrency += q.concurrency

	if q.memory != math.MaxInt64 {
		c.availableMemory += q.memory
	}
}

type Query struct {
	id QueryID
	c  *Controller

	Spec query.QuerySpec
	now  time.Time

	err error

	ready chan<- []execute.Result
	Ready <-chan []execute.Result

	ctx context.Context

	mu     sync.Mutex
	state  State
	cancel func()

	queueSpan,
	planSpan,
	requeueSpan,
	executeSpan *span

	plan *plan.PlanSpec

	concurrency int
	memory      int64
}

func (q *Query) ID() QueryID {
	return q.id
}
func (q *Query) Cancel() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cancel()
	if q.state != Errored {
		q.state = Canceled
	}
}

func (q *Query) Done() {
	q.mu.Lock()
	defer q.mu.Unlock()
	switch q.state {
	case Queueing:
		queueingGauge.Dec()
	case Planning:
		planningGauge.Dec()
	case Requeueing:
		requeueingGauge.Dec()
	case Executing:
		q.executeSpan.Finish()
		executingGauge.Dec()

		q.state = Finished
	case Canceled:
	}
	q.c.queryDone <- q
	close(q.ready)
	q.recordMetrics()
}

func (q *Query) recordMetrics() {
	queueingHist.Observe(q.queueSpan.Duration.Seconds())
	if q.requeueSpan != nil {
		requeueingHist.Observe(q.requeueSpan.Duration.Seconds())
	}
	planningHist.Observe(q.planSpan.Duration.Seconds())
	executingHist.Observe(q.executeSpan.Duration.Seconds())
}

func (q *Query) State() State {
	q.mu.Lock()
	s := q.state
	q.mu.Unlock()
	return s
}

func (q *Query) isOK() bool {
	q.mu.Lock()
	ok := q.state != Canceled && q.state != Errored
	q.mu.Unlock()
	return ok
}
func (q *Query) Err() error {
	q.mu.Lock()
	err := q.err
	q.mu.Unlock()
	return err
}
func (q *Query) setErr(err error) {
	q.mu.Lock()
	q.err = err
	q.state = Errored
	q.mu.Unlock()
}

func (q *Query) setResults(r []execute.Result) {
	q.mu.Lock()
	if q.state == Executing {
		q.ready <- r
	}
	q.mu.Unlock()
}

// tryRequeue attempts to transition the query into the Requeueing state.
func (q *Query) tryRequeue() bool {
	q.mu.Lock()
	if q.state == Planning {
		q.planSpan.Finish()
		planningGauge.Dec()

		q.requeueSpan, q.ctx = StartSpanFromContext(q.ctx, "requeueing")
		requeueingGauge.Inc()

		q.state = Requeueing
		q.mu.Unlock()
		return true
	}
	q.mu.Unlock()
	return false
}

// tryPlan attempts to transition the query into the Planning state.
func (q *Query) tryPlan() bool {
	q.mu.Lock()
	if q.state == Queueing {
		q.queueSpan.Finish()
		queueingGauge.Dec()

		q.planSpan, q.ctx = StartSpanFromContext(q.ctx, "planning")
		planningGauge.Inc()

		q.state = Planning
		q.mu.Unlock()
		return true
	}
	q.mu.Unlock()
	return false
}

// tryExec attempts to transition the query into the Executing state.
func (q *Query) tryExec() bool {
	q.mu.Lock()
	if q.state == Requeueing || q.state == Planning {
		switch q.state {
		case Requeueing:
			q.requeueSpan.Finish()
			requeueingGauge.Dec()
		case Planning:
			q.planSpan.Finish()
			planningGauge.Dec()
		}

		q.executeSpan, q.ctx = StartSpanFromContext(q.ctx, "executing")
		executingGauge.Inc()

		q.state = Executing
		q.mu.Unlock()
		return true
	}
	q.mu.Unlock()
	return false
}

type State int

const (
	Queueing State = iota
	Planning
	Requeueing
	Executing
	Errored
	Finished
	Canceled
)

func (s State) String() string {
	switch s {
	case Queueing:
		return "queueing"
	case Planning:
		return "planning"
	case Requeueing:
		return "requeing"
	case Executing:
		return "executing"
	case Errored:
		return "errored"
	case Finished:
		return "finished"
	case Canceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type span struct {
	s        opentracing.Span
	start    time.Time
	Duration time.Duration
}

func StartSpanFromContext(ctx context.Context, operationName string) (*span, context.Context) {
	start := time.Now()
	s, sctx := opentracing.StartSpanFromContext(ctx, operationName, opentracing.StartTime(start))
	return &span{
		s:     s,
		start: start,
	}, sctx
}

func (s *span) Finish() {
	finish := time.Now()
	s.Duration = finish.Sub(s.start)
	s.s.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: finish,
	})
}
