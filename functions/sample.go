package functions

import (
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"math/rand"
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
	N      int
	offset int
	rows   []execute.Row
	Pos    int
}

func createSampleTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*SampleProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}

	usePos := ps.Pos
	if usePos < 0 {
		usePos = rand.Intn(ps.N)
	}

	ss := SampleSelector{
		N:   ps.N,
		Pos: usePos,
	}

	t, d := execute.NewSelectorTransformationAndDataset(id, mode, ctx.Bounds(), &ss, ps.UseRowTime)
	return t, d, nil
}

func (s *SampleSelector) Do(vs []float64, rr execute.RowReader) {
	var i int
	for i = s.offset + s.Pos; i < len(vs); i += s.N {
		s.rows = append(s.rows, execute.ReadRow(i, rr))
	}

	if s.Pos > 0 {
		s.Pos = 0
	}
	s.offset = i - len(vs)
}

func (s *SampleSelector) Rows() []execute.Row {
	return s.rows
}

func (s *SampleSelector) Reset() {
	s.offset = 0
	s.Pos = rand.Intn(s.N)
	s.rows = nil
}
