package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const CountKind = "count"

type CountOpSpec struct {
}

func init() {
	ifql.RegisterFunction(CountKind, createCountOpSpec)
	query.RegisterOpSpec(CountKind, newCountOp)
	plan.RegisterProcedureSpec(CountKind, newCountProcedure, CountKind)
	execute.RegisterTransformation(CountKind, createCountTransformation)
}
func createCountOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`count function requires no arguments`)
	}

	return new(CountOpSpec), nil
}

func newCountOp() query.OperationSpec {
	return new(CountOpSpec)
}

func (s *CountOpSpec) Kind() query.OperationKind {
	return CountKind
}

type CountProcedureSpec struct {
}

func newCountProcedure(query.OperationSpec) (plan.ProcedureSpec, error) {
	return new(CountProcedureSpec), nil
}

func (s *CountProcedureSpec) Kind() plan.ProcedureKind {
	return CountKind
}

type CountAgg struct {
	count float64
}

func createCountTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.ExecutionContext) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformation(id, mode, ctx.Bounds(), new(CountAgg))
	return t, d, nil
}

func (a *CountAgg) Do(vs []float64) {
	a.count += float64(len(vs))
}
func (a *CountAgg) Value() float64 {
	return a.count
}
func (a *CountAgg) Reset() {
	a.count = 0
}
