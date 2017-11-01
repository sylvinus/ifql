package influxql

import (
	"log"
	"time"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/influxql"
)

type transformer struct {
	stmt *influxql.SelectStatement
	cond influxql.Expr
	tr   influxql.TimeRange

	// output
	ops    []*query.Operation
	edges  []query.Edge
	parent query.OperationID
	err    error
}

// transform takes a validated SelectStatement and produces a QuerySpec
func (t *transformer) transform() (*query.QuerySpec, error) {
	t.build()

	t.addSelect()
	t.addRange()

	return &query.QuerySpec{
		Operations: t.ops,
		Edges:      t.edges,
	}, t.err
}

func (t *transformer) addSelect() {
	t.ops = []*query.Operation{
		{
			ID: "select",
			Spec: &functions.SelectOpSpec{
				Database: t.database(),
			},
		},
	}
	t.parent = "select"
}

func (t *transformer) addRange() {
	var ro functions.RangeOpSpec
	if !t.tr.Min.IsZero() {
		ro.Start = query.Time{Absolute: t.tr.Min}
	}

	if !t.tr.Max.IsZero() {
		ro.Stop = query.Time{Absolute: t.tr.Max}
	}

	id := query.OperationID("range")

	t.ops = append(t.ops, &query.Operation{ID: id, Spec: &ro})
	t.edges = append(t.edges, query.Edge{
		Parent: t.parent,
		Child:  id,
	})
	t.parent = id
}

func (t *transformer) build() {
	t.cond, t.tr, t.err = influxql.ConditionExpr(t.stmt.Condition, &influxql.NowValuer{Now: time.Now()})
	if t.err != nil {
		return
	}

	log.Println(t.cond, t.tr.MinTime(), t.tr.MaxTime())
}

func (t *transformer) database() string {
	return t.stmt.Sources.Measurements()[0].Database
}
