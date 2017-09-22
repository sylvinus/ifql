package ifql

import (
	"github.com/influxdata/ifql/query"
)

// NewQuery parses IFQL into an AST; validates and checks the AST; and produces a query.QuerySpec.
func NewQuery(ifql string, opts ...Option) (*query.QuerySpec, error) {
	program, err := NewAST(ifql, opts...)
	if err != nil {
		return nil, err
	}
	return Evaluate(program)
}
