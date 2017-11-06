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
	rows []execute.Row
}

func createMinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*MinProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(MinSelector), ps.UseRowTime)
	return t, d, nil
}

type MinIntSelector struct {
	MinSelector
	min int64
}
type MinUIntSelector struct {
	MinSelector
	min uint64
}
type MinFloatSelector struct {
	MinSelector
	min float64
}

func (s *MinSelector) NewBoolSelector() execute.DoBoolRowSelector {
	return nil
}

func (s *MinSelector) NewIntSelector() execute.DoIntRowSelector {
	return new(MinIntSelector)
}

func (s *MinSelector) NewUIntSelector() execute.DoUIntRowSelector {
	return new(MinUIntSelector)
}

func (s *MinSelector) NewFloatSelector() execute.DoFloatRowSelector {
	return new(MinFloatSelector)
}

func (s *MinSelector) NewStringSelector() execute.DoStringRowSelector {
	return nil
}

func (s *MinSelector) Rows() []execute.Row {
	if !s.set {
		return nil
	}
	return s.rows
}

func (s *MinSelector) selectRow(idx int, rr execute.RowReader) {
	// Capture row
	if idx >= 0 {
		s.rows = []execute.Row{execute.ReadRow(idx, rr)}
	}
}

func (s *MinIntSelector) DoInt(vs []int64, rr execute.RowReader) {
	minIdx := -1
	for i, v := range vs {
		if !s.set || v < s.min {
			s.set = true
			s.min = v
			minIdx = i
		}
	}
	s.selectRow(minIdx, rr)
}
func (s *MinUIntSelector) DoUInt(vs []uint64, rr execute.RowReader) {
	minIdx := -1
	for i, v := range vs {
		if !s.set || v < s.min {
			s.set = true
			s.min = v
			minIdx = i
		}
	}
	s.selectRow(minIdx, rr)
}
func (s *MinFloatSelector) DoFloat(vs []float64, rr execute.RowReader) {
	minIdx := -1
	for i, v := range vs {
		if !s.set || v < s.min {
			s.set = true
			s.min = v
			minIdx = i
		}
	}
	s.selectRow(minIdx, rr)
}
