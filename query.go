/*
Package ifql contains the parser, query engine, query functions
and a basic server and HTTP client for the IFQL query language and
engine.
*/
package ifql

import (

	// Import functions

	_ "github.com/influxdata/ifql/functions"

	"github.com/influxdata/ifql/query/control"
	"github.com/influxdata/ifql/query/execute"
	"github.com/pkg/errors"
)

type Config struct {
	Hosts []string

	ConcurrencyQuota int
	MemoryBytesQuota int
}

// Use type aliases to expose simple API for entire project

// Controller provides a central location to manage all incoming queries.
// The controller is responsible for queueing, planning, and executing queries.
type Controller = control.Controller

// Query represents a single request.
type Query = control.Query

func NewController(conf Config) (*Controller, error) {
	s, err := execute.NewStorageReader(conf.Hosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}
	c := control.Config{
		ConcurrencyQuota: conf.ConcurrencyQuota,
		MemoryBytesQuota: int64(conf.MemoryBytesQuota),
		ExecutorConfig: execute.Config{
			StorageReader: s,
		},
	}
	return control.New(c), nil
}
