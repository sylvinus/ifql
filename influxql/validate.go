package influxql

import (
	"errors"

	"github.com/influxdata/influxql"
)

var (
	errStatement    = errors.New("parse error: unsupported statement, expected SELECT")
	errMultiple     = errors.New("parse error: multiple InfluxQL statements not supported")
	errInto         = errors.New("parse error: SELECT INTO not supported")
	errSubquery     = errors.New("parse error: subqueries not supported")
	errDatabaseName = errors.New("parse error: database name required in FROM")
	errMultipleDBs  = errors.New("parse error: multiple databases found in FROM")
)

// The validateVisitor type ensures an *influxql.Query meets the requirements
// for transformation to a query.QuerySpec.
type validateVisitor struct {
	s   *influxql.SelectStatement
	err error
}

func (v *validateVisitor) Visit(node influxql.Node) influxql.Visitor {
	switch n := node.(type) {
	case influxql.Statements:
		if len(n) != 1 {
			v.err = errMultiple
			return nil
		}

	case *influxql.SubQuery:
		v.err = errSubquery
		return nil

	case *influxql.SelectStatement:
		v.s = n
		if v.s.Target != nil {
			v.err = errInto
			return nil
		}

		if influxql.HasTimeExpr(n.Condition) {
			if err := ValidateTimeRangeExpr(n.Condition); err != nil {
				v.err = errors.New("parse error: " + err.Error())
				return nil
			}
		}

	case influxql.Statement:
		v.err = errStatement
		return nil

	case influxql.Sources:
		// validate each source first
		for _, s := range n {
			influxql.Walk(v, s)
		}

		if v.err != nil {
			return nil
		}

		// all sources are a *Measurement
		ms := n.Measurements()
		db := ms[0].Database
		if len(ms) > 1 {
			for i := 1; i < len(ms); i++ {
				if db != ms[i].Database {
					v.err = errMultipleDBs
					return nil
				}
			}
		}

		if db == "" {
			v.err = errDatabaseName
			return nil
		}

	}

	return v
}
