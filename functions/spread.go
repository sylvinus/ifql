package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

// SpreadKind is the registration name for ifql, query, plan, and execution.
const SpreadKind = "spread"

func init() {
	ifql.RegisterFunction(SpreadKind, createSpreadOpSpec)
	query.RegisterOpSpec(SpreadKind, newSpreadOp)
	plan.RegisterProcedureSpec(SpreadKind, newSpreadProcedure, SpreadKind)
	execute.RegisterTransformation(SpreadKind, createSpreadTransformation)
}

func createSpreadOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	return new(SpreadOpSpec), nil
}

func newSpreadOp() query.OperationSpec {
	return new(SpreadOpSpec)
}

// SpreadOpSpec defines the required arguments for IFQL.  Currently,
// spread takes no arguments.
type SpreadOpSpec struct{}

// Kind is used to lookup createSpreadOpSpec producing SpreadOpSpec
func (s *SpreadOpSpec) Kind() query.OperationKind {
	return SpreadKind
}

func newSpreadProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	_, ok := qs.(*SpreadOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &SpreadProcedureSpec{}, nil
}

// SpreadProcedureSpec is created when mapping from SpreadOpSpec.Kind
// to a CreateProcedureSpec.
type SpreadProcedureSpec struct{}

// Kind is used to lookup CreateTransformation producing SpreadAgg
func (s *SpreadProcedureSpec) Kind() plan.ProcedureKind {
	return SpreadKind
}

func createSpreadTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, ctx.Bounds(), new(SpreadAgg))
	return t, d, nil
}

// SpreadAgg finds the difference between the max and min values a block
type SpreadAgg struct {
	min float64
	max float64
}

// Do searches for the min and max value of the array and caches them in the aggregate
func (a *SpreadAgg) Do(vs []float64) {
	var (
		minSet bool
		maxSet bool
	)
	for _, v := range vs {
		if !minSet || v < a.min {
			minSet = true
			a.min = v
		}
		if !maxSet || v > a.max {
			maxSet = true
			a.max = v
		}
	}
}

// Value returns the difference between max and min
func (a *SpreadAgg) Value() float64 {
	return a.max - a.min
}

// Reset clears the values of min and max
func (a *SpreadAgg) Reset() {
	a.min = 0
	a.max = 0
}
