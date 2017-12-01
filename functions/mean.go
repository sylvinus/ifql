package functions

import (
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
	ifql.RegisterMethod(MeanKind, createMeanOpSpec)
	query.RegisterOpSpec(MeanKind, newMeanOp)
	plan.RegisterProcedureSpec(MeanKind, newMeanProcedure, MeanKind)
	execute.RegisterTransformation(MeanKind, createMeanTransformation)
}
func createMeanOpSpec(args ifql.Arguments, ctx ifql.Context) (query.OperationSpec, error) {
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
func (s *MeanProcedureSpec) Copy() plan.ProcedureSpec {
	return new(MeanProcedureSpec)
}

type MeanAgg struct {
	count float64
	sum   float64
}

func createMeanTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(MeanAgg))
	return t, d, nil
}

func (a *MeanAgg) reset() {
	a.count = 0
	a.sum = 0
}
func (a *MeanAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *MeanAgg) NewIntAgg() execute.DoIntAgg {
	a.reset()
	return a
}

func (a *MeanAgg) NewUIntAgg() execute.DoUIntAgg {
	a.reset()
	return a
}

func (a *MeanAgg) NewFloatAgg() execute.DoFloatAgg {
	a.reset()
	return a
}

func (a *MeanAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

func (a *MeanAgg) DoInt(vs []int64) {
	a.count += float64(len(vs))
	for _, v := range vs {
		//TODO handle overflow
		a.sum += float64(v)
	}
}
func (a *MeanAgg) DoUInt(vs []uint64) {
	a.count += float64(len(vs))
	for _, v := range vs {
		//TODO handle overflow
		a.sum += float64(v)
	}
}
func (a *MeanAgg) DoFloat(vs []float64) {
	a.count += float64(len(vs))
	for _, v := range vs {
		a.sum += v
	}
}
func (a *MeanAgg) Type() execute.DataType {
	return execute.TFloat
}
func (a *MeanAgg) ValueFloat() float64 {
	if a.count < 1 {
		return math.NaN()
	}
	return a.sum / a.count
}
