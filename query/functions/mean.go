package functions

import (
	"fmt"
	"math"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const MeanKind = "mean"

type MeanOpSpec struct {
}

func init() {
	ifql.RegisterFunction(MeanKind, createMeanOpSpec)
	query.RegisterOpSpec(MeanKind, newMeanOp)
	plan.RegisterProcedureSpec(MeanKind, newMeanProcedure, MeanKind)
	execute.RegisterTransformation(MeanKind, createMeanTransformation)
}
func createMeanOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`mean function requires no arguments`)
	}

	return new(MeanOpSpec), nil
}

func newMeanOp() query.OperationSpec {
	return new(MeanOpSpec)
}

func (s *MeanOpSpec) Kind() query.OperationKind {
	return MeanKind
}

type MeanProcedureSpec struct {
}

func newMeanProcedure(query.OperationSpec) (plan.ProcedureSpec, error) {
	return new(MeanProcedureSpec), nil
}

func (s *MeanProcedureSpec) Kind() plan.ProcedureKind {
	return MeanKind
}

type MeanAgg struct {
	count float64
	sum   float64
}

func createMeanTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(MeanAgg))
	return t, d, nil
}

func (a *MeanAgg) Do(vs []float64) {
	a.count += float64(len(vs))
	for _, v := range vs {
		a.sum += v
	}
}
func (a *MeanAgg) Value() float64 {
	if a.count < 1 {
		return math.NaN()
	}
	return a.sum / a.count
}
func (a *MeanAgg) Reset() {
	a.sum = 0
	a.count = 0
}
