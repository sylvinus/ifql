package execute

import (
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Process interface {
	Do(src DataFrame) (DataFrame, bool)
	Spec() plan.ProcedureSpec
}

func processFromProcedureSpec(spec plan.ProcedureSpec, now time.Time) Process {
	switch s := spec.(type) {
	case *plan.SumProcedureSpec:
		return sumProc{
			spec: s,
		}
	case *plan.RangeProcedureSpec:
		return rangeProc{
			spec: s,
			stop: Time(s.Bounds.Stop.Time(now).UnixNano()),
		}
	case *plan.WhereProcedureSpec:
		return passthroughProc{
			spec: s,
		}
	default:
		return nil
	}
}

type passthroughProc struct {
	spec plan.ProcedureSpec
}

func (o passthroughProc) Spec() plan.ProcedureSpec {
	return o.spec
}

func (passthroughProc) Do(src DataFrame) (DataFrame, bool) {
	return src, true
}

type sumProc struct {
	spec *plan.SumProcedureSpec
}

func (sumProc) Do(src DataFrame) (DataFrame, bool) {
	dst := NewWriteDataFrame(src.NRows(), 1, src.Bounds())
	stop := src.Bounds().Stop()
	dst.SetColTime(0, stop)
	n := src.NRows()
	for i := 0; i < n; i++ {
		row, tags := src.RowSlice(i)
		dst.SetRowTags(i, tags)
		sum := 0.0
		for _, v := range row {
			sum += v
		}
		dst.Set(i, 0, sum)
	}
	return dst, true
}

func (o sumProc) Spec() plan.ProcedureSpec {
	return o.spec
}

type rangeProc struct {
	spec *plan.RangeProcedureSpec
	stop Time
}

func (o rangeProc) Do(src DataFrame) (DataFrame, bool) {
	if src.Bounds().Stop() > o.stop {
		return nil, false
	}
	return src, true
}

func (o rangeProc) Spec() plan.ProcedureSpec {
	return o.spec
}
