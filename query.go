/*
The IFQL package contains the parser, query engine, query functions
and a basic server and HTTP client for the IFQL query language and
engine.
*/
package ifql

import (
	"context"

	_ "github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query/execute"
	"github.com/pkg/errors"
)

func Query(ctx context.Context, queryStr string, verbose, trace bool) ([]execute.Result, error) {
	qSpec, err := ifql.NewQuery(queryStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
	}

	var opts []execute.Option
	if verbose {
		opts = append(opts, execute.Verbose())
	}
	if trace {
		opts = append(opts, execute.Trace())
	}
	return execute.Execute(ctx, qSpec, opts...)
}
