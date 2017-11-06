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
func (s *FirstProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FirstProcedureSpec)
	*ns = *s
	return ns
}

type FirstSelector struct {
	selected bool
}

func createFirstTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*FirstProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(FirstSelector), ps.UseRowTime)
	return t, d, nil
}

func (s *FirstSelector) reset() {
	s.selected = false
}

func (s *FirstSelector) NewBoolSelector() execute.DoBoolIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewIntSelector() execute.DoIntIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewUIntSelector() execute.DoUIntIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewFloatSelector() execute.DoFloatIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewStringSelector() execute.DoStringIndexSelector {
	s.reset()
	return s
}

func (s *FirstSelector) selectFirst(l int) []int {
	if !s.selected && l > 0 {
		s.selected = true
		return []int{0}
	}
	return nil
}
func (s *FirstSelector) DoBool(vs []bool) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoInt(vs []int64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoUInt(vs []uint64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoFloat(vs []float64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoString(vs []string) []int {
	return s.selectFirst(len(vs))
}
