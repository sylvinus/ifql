package influxql_test

import (
	"errors"
	"testing"

	"github.com/influxdata/ifql/influxql"
)

var (
	errStatement    = errors.New("parse error: unsupported statement, expected SELECT")
	errMultiple     = errors.New("parse error: multiple InfluxQL statements not supported")
	errInto         = errors.New("parse error: SELECT INTO not supported")
	errSubquery     = errors.New("parse error: subqueries not supported")
	errDatabaseName = errors.New("parse error: database name required in FROM")
	errMultipleDBs  = errors.New("parse error: multiple databases found in FROM")
	errTime         = errors.New("parse error: time cannot be combined with OR operator")
)

func errEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}

	return a.Error() == b.Error()
}

func TestParseQuery_Unsupported(t *testing.T) {
	cases := []struct {
		sql string
		err error
	}{
		{`DROP MEASUREMENT m0`, errStatement},
		{`DROP SERIES FROM m0`, errStatement},
		{`DELETE FROM m0`, errStatement},
		{`SELECT v0 FROM m0; SELECT v1 FROM m0`, errMultiple},
		{`SELECT v0 INTO m1 FROM m0`, errInto},
		{`SELECT v0 FROM (SELECT v0 FROM m0)`, errSubquery},
		{`SELECT v0 FROM (SELECT v0 FROM (SELECT v0 FROM m0))`, errSubquery},
		{`SELECT v0 FROM m0`, errDatabaseName},
		{`SELECT v0 FROM rp.m0`, errDatabaseName},

		// --
		// queries that aren't supported for IFQL

		// InfluxDB 1.4: invalid, no results; IFQL: not supported
		{`SELECT v0 FROM db0..m0 WHERE (time >= 0 AND time < 10) OR (time >= 20 AND time < 30)`, errTime},
		// InfluxDB 1.4: invalid, no results; IFQL: not supported
		{`SELECT v0 FROM db0..m0 WHERE time = 0 OR time = 1 `, errTime},

		// --
		// queries that probably should be supported but do not currently have an analog in IFQL

		// TODO(sgc): no way to return results from multiple databases in sequence
		{`SELECT v0 FROM db0..m0, db1..m1`, errMultipleDBs},
	}

	for _, tc := range cases {
		t.Run(tc.sql, func(t *testing.T) {
			if _, err := influxql.ParseQuery(tc.sql); !errEqual(err, tc.err) {
				t.Errorf("unexpected error; got=%v, exp=%s", err, tc.err)
			}
		})
	}
}
