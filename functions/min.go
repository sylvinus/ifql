package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const MinKind = "min"

type MinOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
}

func init() {
	ifql.RegisterFunction(MinKind, createMinOpSpec)
	query.RegisterOpSpec(MinKind, newMinOp)
	plan.RegisterProcedureSpec(MinKind, newMinProcedure, MinKind)
	execute.RegisterTransformation(MinKind, createMinTransformation)
}

func createMinOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(MinOpSpec)
	if value, ok := args["useRowTime"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf(`min useRowTime argument must be a boolean`)
		}
		spec.UseRowTime = value.Value.(bool)
	}

	return spec, nil
}

func newMinOp() query.OperationSpec {
	return new(MinOpSpec)
}

func (s *MinOpSpec) Kind() query.OperationKind {
	return MinKind
}

type MinProcedureSpec struct {
	UseRowTime bool
}

func newMinProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MinOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &MinProcedureSpec{
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *MinProcedureSpec) Kind() plan.ProcedureKind {
	return MinKind
}

type MinSelector struct {
	set  bool
	min  float64
	rows []execute.Row
}

func createMinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*MinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(MinSelector), ps.UseRowTime)
	return t, d, nil
}

func (s *MinSelector) Do(vs []float64, rr execute.RowReader) {
	minIdx := -1
	for i, v := range vs {
		if !s.set || v < s.min {
			s.set = true
			s.min = v
			minIdx = i
		}
	}
	// Capture minimum row
	if minIdx >= 0 {
		s.rows = []execute.Row{execute.ReadRow(minIdx, rr)}
	}
}
func (s *MinSelector) Rows() []execute.Row {
	if !s.set {
		return nil
	}
	return s.rows
}
func (s *MinSelector) Reset() {
	s.set = false
	s.rows = nil
}
