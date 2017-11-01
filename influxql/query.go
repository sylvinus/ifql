package influxql

import (
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/influxql"
)

func ParseQuery(s string) (*query.QuerySpec, error) {
	var (
		stmt *influxql.SelectStatement
		err  error
	)

	if stmt, err = parseAndValidateSelect(s); err != nil {
		return nil, err
	}

	t := transformer{stmt: stmt}
	return t.transform()
}

func parseAndValidateSelect(s string) (*influxql.SelectStatement, error) {
	q, err := influxql.ParseQuery(s)
	if err != nil {
		return nil, err
	}

	var v validateVisitor
	influxql.Walk(&v, q)

	return v.s, v.err
}
