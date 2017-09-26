package functions

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SumKind = "sum"

type SumOpSpec struct {
}

func init() {
	ifql.RegisterFunction(SumKind, createSumOpSpec)
	query.RegisterOpSpec(SumKind, newSumOp)
	plan.RegisterProcedureSpec(SumKind, newSumProcedure, SumKind)
	execute.RegisterTransformation(SumKind, createSumTransformation)
}

func createSumOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`sum function requires no arguments`)
	}

	return new(SumOpSpec), nil
}

func newSumOp() query.OperationSpec {
	return new(SumOpSpec)
}

func (s *SumOpSpec) Kind() query.OperationKind {
	return SumKind
}

type SumProcedureSpec struct {
}

func newSumProcedure(query.OperationSpec) (plan.ProcedureSpec, error) {
	return new(SumProcedureSpec), nil
}

func (s *SumProcedureSpec) Kind() plan.ProcedureKind {
	return SumKind
}

type SumAgg struct {
	sum float64
}

func createSumTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, now time.Time) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := execute.NewAggregateTransformation(d, cache, new(SumAgg))
	return t, d, nil
}

func (a *SumAgg) Do(vs []float64) {
	for _, v := range vs {
		a.sum += v
	}
}
func (a *SumAgg) Value() float64 {
	return a.sum
}
func (a *SumAgg) Reset() {
	a.sum = 0
}
