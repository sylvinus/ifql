package execute

import (
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Operation interface {
	Do(src DataFrame) (DataFrame, bool)
	Spec() plan.OperationSpec
}

func operationFromSpec(spec plan.OperationSpec, now time.Time) Operation {
	switch s := spec.(type) {
	case *plan.SumOpSpec:
		return sumOp{
			spec: s,
		}
	case *plan.RangeOpSpec:
		return rangeOp{
			spec: s,
			stop: Time(s.Bounds.Stop.Time(now).UnixNano()),
		}
	default:
		return nil
	}
}

type sumOp struct {
	spec *plan.SumOpSpec
}

func (sumOp) Do(src DataFrame) (DataFrame, bool) {
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

func (o sumOp) Spec() plan.OperationSpec {
	return o.spec
}

type rangeOp struct {
	spec *plan.RangeOpSpec
	stop Time
}

func (o rangeOp) Do(src DataFrame) (DataFrame, bool) {
	//if src.Bounds().Stop() > o.stop {
	//	return nil, false
	//}
	return src, true
}

func (o rangeOp) Spec() plan.OperationSpec {
	return o.spec
}
