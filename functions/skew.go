package functions

import (
	"fmt"
	"math"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SkewKind = "skew"

type SkewOpSpec struct {
}

func init() {
	ifql.RegisterFunction(SkewKind, createSkewOpSpec)
	query.RegisterOpSpec(SkewKind, newSkewOp)
	plan.RegisterProcedureSpec(SkewKind, newSkewProcedure, SkewKind)
	execute.RegisterTransformation(SkewKind, createSkewTransformation)
}
func createSkewOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`skew function requires no arguments`)
	}

	return new(SkewOpSpec), nil
}

func newSkewOp() query.OperationSpec {
	return new(SkewOpSpec)
}

func (s *SkewOpSpec) Kind() query.OperationKind {
	return SkewKind
}

type SkewProcedureSpec struct {
}

func newSkewProcedure(query.OperationSpec) (plan.ProcedureSpec, error) {
	return new(SkewProcedureSpec), nil
}

func (s *SkewProcedureSpec) Kind() plan.ProcedureKind {
	return SkewKind
}

type SkewAgg struct {
	n, m1, m2, m3 float64
}

func createSkewTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(SkewAgg))
	return t, d, nil
}

func (a *SkewAgg) Do(vs []float64) {
	for _, v := range vs {
		n0 := a.n
		a.n++
		delta := v - a.m1
		deltaN := delta / a.n
		t := delta * deltaN * n0
		a.m3 += t*deltaN*(a.n-2) - 3*deltaN*a.m2
		a.m2 += t
		a.m1 += deltaN
	}
}
func (a *SkewAgg) Value() float64 {
	if a.n < 2 {
		return math.NaN()
	}
	return math.Sqrt(a.n) * a.m3 / math.Pow(a.m2, 1.5)
}
func (a *SkewAgg) Reset() {
	a.n = 0
	a.m1 = 0
	a.m2 = 0
	a.m3 = 0
}
