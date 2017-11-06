package functions

import (
	"fmt"

	"math/rand"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SampleKind = "sample"

type SampleOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
	N          int  `json:"n"`
	Pos        int  `json:"pos"`
}

func init() {
	ifql.RegisterFunction(SampleKind, createSampleOpSpec)
	query.RegisterOpSpec(SampleKind, newSampleOp)
	plan.RegisterProcedureSpec(SampleKind, newSampleProcedure, SampleKind)
	execute.RegisterTransformation(SampleKind, createSampleTransformation)
}

func createSampleOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(SampleOpSpec)
	if value, ok := args["useRowTime"]; ok {
		if value.Type != ifql.TBool {
			return nil, fmt.Errorf(`sample useRowTime argument must be a boolean`)
		}
		spec.UseRowTime = value.Value.(bool)
	}

	if value, ok := args["n"]; ok {
		if value.Type != ifql.TInt {
			return nil, fmt.Errorf(`sample n argument must be an integer`)
		}
		spec.N = value.Value.(int)
	} else {
		return nil, fmt.Errorf(`n argument is required for sample function`)
	}

	if value, ok := args["pos"]; ok {
		if value.Type != ifql.TInt {
			return nil, fmt.Errorf(`sample pos argument must be an integer`)
		}
		spec.Pos = value.Value.(int)
	} else {
		spec.Pos = -1
	}

	return spec, nil
}

func newSampleOp() query.OperationSpec {
	return new(SampleOpSpec)
}

func (s *SampleOpSpec) Kind() query.OperationKind {
	return SampleKind
}

type SampleProcedureSpec struct {
	UseRowTime bool
	N          int
	Pos        int
}

func newSampleProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SampleOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &SampleProcedureSpec{
		UseRowTime: spec.UseRowTime,
		N:          spec.N,
		Pos:        spec.Pos,
	}, nil
}

func (s *SampleProcedureSpec) Kind() plan.ProcedureKind {
	return SampleKind
}

type SampleSelector struct {
	N   int
	Pos int

	offset   int
	selected []int
}

func createSampleTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*SampleProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}

	ss := &SampleSelector{
		N:   ps.N,
		Pos: ps.Pos,
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, ctx.Bounds(), ss, ps.UseRowTime)
	return t, d, nil
}

func (s *SampleSelector) reset() {
	pos := s.Pos
	if pos < 0 {
		pos = rand.Intn(s.N)
	}
	s.offset = pos
}

func (s *SampleSelector) NewBoolSelector() execute.DoBoolIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewIntSelector() execute.DoIntIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewUIntSelector() execute.DoUIntIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewFloatSelector() execute.DoFloatIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewStringSelector() execute.DoStringIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) selectSample(l int) []int {
	var i int
	s.selected = s.selected[0:0]
	for i = s.offset; i < l; i += s.N {
		s.selected = append(s.selected, i)
	}
	s.offset = i - l
	return s.selected
}

func (s *SampleSelector) DoBool(vs []bool) []int {
	return s.selectSample(len(vs))
}
func (s *SampleSelector) DoInt(vs []int64) []int {
	return s.selectSample(len(vs))
}
func (s *SampleSelector) DoUInt(vs []uint64) []int {
	return s.selectSample(len(vs))
}
func (s *SampleSelector) DoFloat(vs []float64) []int {
	return s.selectSample(len(vs))
}
func (s *SampleSelector) DoString(vs []string) []int {
	return s.selectSample(len(vs))
}
