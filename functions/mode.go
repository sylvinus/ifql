package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const ModeKind = "mode"

type ModeOpSpec struct {
}

func init() {
	ifql.RegisterFunction(ModeKind, createModeOpSpec)
	query.RegisterOpSpec(ModeKind, newModeOp)
	plan.RegisterProcedureSpec(ModeKind, newModeProcedure, ModeKind)
	execute.RegisterTransformation(ModeKind, createModeTransformation)
}
func createModeOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(`mode function requires no arguments`)
	}

	return new(ModeOpSpec), nil
}

func newModeOp() query.OperationSpec {
	return new(ModeOpSpec)
}

func (s *ModeOpSpec) Kind() query.OperationKind {
	return ModeKind
}

type ModeProcedureSpec struct {
}

func newModeProcedure(query.OperationSpec) (plan.ProcedureSpec, error) {
	return new(ModeProcedureSpec), nil
}

func (s *ModeProcedureSpec) Kind() plan.ProcedureKind {
	return ModeKind
}
func (s *ModeProcedureSpec) Copy() plan.ProcedureSpec {
	return new(ModeProcedureSpec)
}

type ModeAgg struct {
	Float64  map[float64]uint64
	Bool     map[bool]uint64
	Uint64   map[uint64]uint64
	Int64    map[int64]uint64
	String   map[string]uint64
	DataType execute.DataType
}

func createModeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(ModeAgg))
	return t, d, nil
}

func (a *ModeAgg) NewBoolAgg() execute.DoBoolAgg {
	a.Bool = make(map[bool]uint64)
	a.DataType = execute.TBool
	return a
}

func (a *ModeAgg) NewIntAgg() execute.DoIntAgg {
	a.Int64 = make(map[int64]uint64)
	a.DataType = execute.TInt
	return a
}

func (a *ModeAgg) NewUIntAgg() execute.DoUIntAgg {
	a.Uint64 = make(map[uint64]uint64)
	a.DataType = execute.TUInt
	return a
}

func (a *ModeAgg) NewFloatAgg() execute.DoFloatAgg {
	a.Float64 = make(map[float64]uint64)
	a.DataType = execute.TFloat
	return a
}

func (a *ModeAgg) NewStringAgg() execute.DoStringAgg {
	a.String = make(map[string]uint64)
	a.DataType = execute.TString
	return a
}

func (a *ModeAgg) DoBool(vs []bool) {
	for _, v := range vs {
		a.Bool[v]++
	}
}

func (a *ModeAgg) DoInt(vs []int64) {
	for _, v := range vs {
		a.Int64[v]++
	}
}

func (a *ModeAgg) DoUInt(vs []uint64) {
	for _, v := range vs {
		a.Uint64[v]++
	}
}

func (a *ModeAgg) DoFloat(vs []float64) {
	for _, v := range vs {
		a.Float64[v]++
	}
}

func (a *ModeAgg) DoString(vs []string) {
	for _, v := range vs {
		a.String[v]++
	}
}

func (a *ModeAgg) Type() execute.DataType {
	return a.DataType
}

func (a *ModeAgg) ValueInt() int64 {
	var (
		maxCount uint64
		max      int64
	)
	for val, count := range a.Int64 {
		if count > maxCount {
			max = val
			maxCount = count
		}
	}
	return max
}

func (a *ModeAgg) ValueBool() bool {
	var (
		maxCount uint64
		max      bool
	)
	for val, count := range a.Bool {
		if count > maxCount {
			max = val
			maxCount = count
		}
	}
	return max
}

func (a *ModeAgg) ValueString() string {
	var (
		maxCount uint64
		max      string
	)
	for val, count := range a.String {
		if count > maxCount {
			max = val
			maxCount = count
		}
	}
	return max
}

func (a *ModeAgg) ValueFloat() float64 {
	var (
		maxCount uint64
		max      float64
	)
	for val, count := range a.Float64 {
		if count > maxCount {
			max = val
			maxCount = count
		}
	}
	return max
}

func (a *ModeAgg) ValueUInt() uint64 {
	var (
		maxCount uint64
		max      uint64
	)
	for val, count := range a.Uint64 {
		if count > maxCount {
			max = val
			maxCount = count
		}
	}
	return max
}
