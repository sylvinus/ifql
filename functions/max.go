package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const MaxKind = "max"

type MaxOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
}

func init() {
	ifql.RegisterFunction(MaxKind, createMaxOpSpec)
	query.RegisterOpSpec(MaxKind, newMaxOp)
	plan.RegisterProcedureSpec(MaxKind, newMaxProcedure, MaxKind)
	execute.RegisterTransformation(MaxKind, createMaxTransformation)
}

func createMaxOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(MaxOpSpec)
	if value, ok := args["useRowTime"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf(`max useRowTime argument must be a boolean`)
		}
		spec.UseRowTime = value.Value.(bool)
	}

	return spec, nil
}

func newMaxOp() query.OperationSpec {
	return new(MaxOpSpec)
}

func (s *MaxOpSpec) Kind() query.OperationKind {
	return MaxKind
}

type MaxProcedureSpec struct {
	UseRowTime bool
}

func newMaxProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MaxOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &MaxProcedureSpec{
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *MaxProcedureSpec) Kind() plan.ProcedureKind {
	return MaxKind
}

type MaxSelector struct {
	set  bool
	max  float64
	rows []execute.Row
}

func createMaxTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*MaxProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(MaxSelector), ps.UseRowTime)
	return t, d, nil
}

func (s *MaxSelector) Do(vs []float64, rr execute.RowReader) {
	maxIdx := -1
	for i, v := range vs {
		if !s.set || v > s.max {
			s.set = true
			s.max = v
			maxIdx = i
		}
	}
	// Capture maximum row
	if maxIdx >= 0 {
		s.rows = []execute.Row{execute.ReadRow(maxIdx, rr)}
	}
}
func (s *MaxSelector) Rows() []execute.Row {
	if !s.set {
		return nil
	}
	return s.rows
}
func (s *MaxSelector) Reset() {
	s.set = false
	s.rows = nil
}
