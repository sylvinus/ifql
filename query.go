/*
The IFQL package contains the parser, query engine, query functions
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

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

type Options struct {
	Verbose bool
	Trace   bool
	Hosts   []string
}

var emptyOptions = new(Options)

func Query(ctx context.Context, queryStr string, opts *Options) ([]execute.Result, error) {
	if opts == nil {
		opts = new(Options)
	}

	qSpec, err := ifql.NewQuery(queryStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
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

	var c execute.Config
	c.Trace = opts.Trace

	s, err := execute.NewStorageReader(opts.Hosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}
	c.StorageReader = s

	e := execute.NewExecutor(c)
	r, err := e.Execute(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	return r, nil
}
