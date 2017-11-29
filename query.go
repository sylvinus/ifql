/*
Package ifql contains the parser, query engine, query functions
and a basic server and HTTP client for the IFQL query language and
engine.
*/
package ifql

import (
	"context"
	"log"
	"time"

	// Import functions
	_ "github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query/control"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

// Options define the query execution options
type Options struct {
	Verbose bool
	Trace   bool
	Hosts   []string
}

var emptyOptions = new(Options)

// ExecuteQuery parses the queryStr according to opts, plans, and returns results and the query specification
func ExecuteQuery(ctx context.Context, queryStr string, opts *Options) ([]execute.Result, *query.QuerySpec, error) {
	if opts == nil {
		opts = new(Options)
	}

	qSpec, err := QuerySpec(ctx, queryStr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse query")
	}

	return QueryWithSpec(ctx, qSpec, opts)
}

func QuerySpec(ctx context.Context, queryStr string) (*query.QuerySpec, error) {
	return ifql.NewQuery(queryStr)
}

// QueryWithSpec unmarshals the JSON plan, returns results and the query specification
func QueryWithSpec(ctx context.Context, qSpec *query.QuerySpec, opts *Options) ([]execute.Result, *query.QuerySpec, error) {
	lplanner := plan.NewLogicalPlanner()
	lp, err := lplanner.Plan(qSpec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create logical plan")
	}
	if opts.Verbose {
		log.Println("logical plan", plan.Formatted(lp))
	}

	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil, time.Now())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create physical plan")
	}
	if opts.Verbose {
		log.Println("physical plan", plan.Formatted(p))
	}

	var c execute.Config
	s, err := execute.NewStorageReader(opts.Hosts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create storage reader")
	}
	c.StorageReader = s

	e := execute.NewExecutor(c)
	r, err := e.Execute(ctx, p)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to execute query")
	}
	return r, qSpec, nil
}

type Controller = control.Controller
type Query = control.Query

func NewController(opts Options) (*Controller, error) {
	s, err := execute.NewStorageReader(opts.Hosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}
	c := control.Config{
		ConcurrencyQuota: 20,
		MemoryBytesQuota: 1e6,
		ExecutorConfig: execute.Config{
			StorageReader: s,
		},
	}
	return control.New(c), nil
}
