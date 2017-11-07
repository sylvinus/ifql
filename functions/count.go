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

func (s *CountProcedureSpec) Copy() plan.ProcedureSpec {
	return new(CountProcedureSpec)
}

func (s *CountProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: nil,
	}
}
func (s *CountProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.AggregateSet {
		root = dup()
		selectSpec = root.Spec.(*SelectProcedureSpec)
		selectSpec.AggregateSet = false
		selectSpec.AggregateType = ""
		return
	}
	selectSpec.AggregateSet = true
	selectSpec.AggregateType = CountKind
}

type CountAgg struct {
	count int64
}

func createCountTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(CountAgg))
	return t, d, nil
}

func (a *CountAgg) NewBoolAgg() execute.DoBoolAgg {
	a.count = 0
	return a
}
func (a *CountAgg) NewIntAgg() execute.DoIntAgg {
	a.count = 0
	return a
}
func (a *CountAgg) NewUIntAgg() execute.DoUIntAgg {
	a.count = 0
	return a
}
func (a *CountAgg) NewFloatAgg() execute.DoFloatAgg {
	a.count = 0
	return a
}
func (a *CountAgg) NewStringAgg() execute.DoStringAgg {
	a.count = 0
	return a
}

func (a *CountAgg) DoBool(vs []bool) {
	a.count += int64(len(vs))
}
func (a *CountAgg) DoUInt(vs []uint64) {
	a.count += int64(len(vs))
}
func (a *CountAgg) DoInt(vs []int64) {
	a.count += int64(len(vs))
}
func (a *CountAgg) DoFloat(vs []float64) {
	a.count += int64(len(vs))
}
func (a *CountAgg) DoString(vs []string) {
	a.count += int64(len(vs))
}

func (a *CountAgg) Type() execute.DataType {
	return execute.TInt
}
func (a *CountAgg) ValueInt() int64 {
	return a.count
}
