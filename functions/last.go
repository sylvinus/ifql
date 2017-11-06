package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const LastKind = "last"

type LastOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
}

func init() {
	ifql.RegisterFunction(LastKind, createLastOpSpec)
	query.RegisterOpSpec(LastKind, newLastOp)
	plan.RegisterProcedureSpec(LastKind, newLastProcedure, LastKind)
	execute.RegisterTransformation(LastKind, createLastTransformation)
}

func createLastOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(LastOpSpec)
	if value, ok := args["useRowTime"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf(`last useRowTime argument must be a boolean`)
		}
		spec.UseRowTime = value.Value.(bool)
	}

	return spec, nil
}

func newLastOp() query.OperationSpec {
	return new(LastOpSpec)
}

func (s *LastOpSpec) Kind() query.OperationKind {
	return LastKind
}

type LastProcedureSpec struct {
	UseRowTime bool
}

func newLastProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LastOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &LastProcedureSpec{
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *LastProcedureSpec) Kind() plan.ProcedureKind {
	return LastKind
}
func (s *LastProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LastProcedureSpec)
	ns.UseRowTime = s.UseRowTime
	return ns
}

type LastSelector struct {
	rows []execute.Row
}

func createLastTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*LastProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(LastSelector), ps.UseRowTime)
	return t, d, nil
}

func (s *LastSelector) reset() {
	s.rows = nil
}
func (s *LastSelector) NewBoolSelector() execute.DoBoolRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewIntSelector() execute.DoIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewUIntSelector() execute.DoUIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewFloatSelector() execute.DoFloatRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewStringSelector() execute.DoStringRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) Rows() []execute.Row {
	return s.rows
}

func (s *LastSelector) selectLast(l int, rr execute.RowReader) {
	if l > 0 {
		s.rows = []execute.Row{execute.ReadRow(l-1, rr)}
	}
}

func (s *LastSelector) DoBool(vs []bool, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoInt(vs []int64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoUInt(vs []uint64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoFloat(vs []float64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoString(vs []string, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
