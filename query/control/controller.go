package control

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
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
		state:  Queued,
		ctx:    cctx,
		cancel: cancel,
	}

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
		log.Println("Controller", c.availableConcurrency, c.availableMemory)
		select {
		// Wait for resources to free
		case q := <-c.queryDone:
			c.availableConcurrency += q.concurrency
			c.availableMemory += q.memory
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
			if !q.advanceState(Planning) {
				continue
			}
			// Plan query to determine needed resources
			lp, err := c.lplanner.Plan(&q.Spec)
			if err != nil {
				q.SetErr(errors.Wrap(err, "failed to create logical plan"))
				continue
			}

			p, err := c.pplanner.Plan(lp, nil, q.now)
			if err != nil {
				q.SetErr(errors.Wrap(err, "failed to create physical plan"))
				continue
			}
			q.concurrency = p.Resources.ConcurrencyQuota
			q.memory = p.Resources.MemoryBytesQuota

			// Check if we have enough resources
			if c.availableConcurrency >= q.concurrency && c.availableMemory >= q.memory {
				// Mark resources as consumed
				c.availableConcurrency -= q.concurrency
				c.availableMemory -= q.memory
				pq.Pop()

				// Execute query
				if !q.advanceState(Executing) {
					continue
				}
				r, err := c.executor.Execute(q.ctx, p)
				if err != nil {
					q.SetErr(errors.Wrap(err, "failed to execute query"))
					continue
				}
				q.setResults(r)
			} // else go back to waiting for new queries or new resources
		}
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
	q.c.queryDone <- q
	close(q.ready)
}

func (q *Query) Done() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.state != Canceled && q.state != Errored {
		q.state = Finished
		q.c.queryDone <- q
		close(q.ready)
	}
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
func (q *Query) SetErr(err error) {
	q.mu.Lock()
	q.err = err
	q.state = Errored
	q.mu.Unlock()
}

func (q *Query) setResults(r []execute.Result) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.state == Executing {
		q.ready <- r
	}
}
func (q *Query) setState(s State) {
	q.mu.Lock()
	q.state = s
	q.mu.Unlock()
}
func (q *Query) advanceState(s State) (ok bool) {
	q.mu.Lock()
	ok = q.state != Canceled && q.state != Errored
	if ok {
		q.state = s
	}
	q.mu.Unlock()
	return
}

type State int

const (
	Queued State = iota
	Planning
	Executing
	Errored
	Finished
	Canceled
)

func (s State) String() string {
	switch s {
	case Queued:
		return "queued"
	case Planning:
		return "planning"
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
