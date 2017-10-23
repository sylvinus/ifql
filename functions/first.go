package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const FirstKind = "first"

type FirstOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
}

func init() {
	ifql.RegisterFunction(FirstKind, createFirstOpSpec)
	query.RegisterOpSpec(FirstKind, newFirstOp)
	plan.RegisterProcedureSpec(FirstKind, newFirstProcedure, FirstKind)
	execute.RegisterTransformation(FirstKind, createFirstTransformation)
}

func createFirstOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(FirstOpSpec)
	if value, ok := args["useRowTime"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf(`first useRowTime argument must be a boolean`)
		}
		spec.UseRowTime = value.Value.(bool)
	}

	return spec, nil
}

func newFirstOp() query.OperationSpec {
	return new(FirstOpSpec)
}

func (s *FirstOpSpec) Kind() query.OperationKind {
	return FirstKind
}

type FirstProcedureSpec struct {
	UseRowTime bool
}

func newFirstProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FirstOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &FirstProcedureSpec{
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *FirstProcedureSpec) Kind() plan.ProcedureKind {
	return FirstKind
}

type FirstSelector struct {
	rows []execute.Row
}

func createFirstTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*FirstProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(FirstSelector), ps.UseRowTime)
	return t, d, nil
}

func (s *FirstSelector) Do(vs []float64, rr execute.RowReader) {
	if s.rows == nil && len(vs) > 0 {
		s.rows = []execute.Row{execute.ReadRow(0, rr)}
	}
}
func (s *FirstSelector) Rows() []execute.Row {
	return s.rows
}
func (s *FirstSelector) Reset() {
	s.rows = nil
}
