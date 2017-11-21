package functions

import (
	"errors"
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

const RangeKind = "range"

type RangeOpSpec struct {
	Start query.Time `json:"start"`
	Stop  query.Time `json:"stop"`
}

func init() {
	ifql.RegisterMethod(RangeKind, createRangeOpSpec)
	query.RegisterOpSpec(RangeKind, newRangeOp)
	plan.RegisterProcedureSpec(RangeKind, newRangeProcedure, RangeKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	//execute.RegisterTransformation(RangeKind, createRangeTransformation)
}

func createRangeOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	startValue, ok := args["start"]
	if !ok {
		return nil, errors.New(`range function requires argument "start"`)
	}

	spec := new(RangeOpSpec)
	start, err := ifql.ToQueryTime(startValue)
	if err != nil {
		return nil, err
	}
	spec.Start = start

	if stopValue, ok := args["stop"]; ok {
		stop, err := ifql.ToQueryTime(stopValue)
		if err != nil {
			return nil, err
		}
		spec.Stop = stop
	} else {
		// Make stop time implicit "now"
		spec.Stop.IsRelative = true
	}

	return spec, nil
}

func newRangeOp() query.OperationSpec {
	return new(RangeOpSpec)
}

func (s *RangeOpSpec) Kind() query.OperationKind {
	return RangeKind
}

type RangeProcedureSpec struct {
	Bounds plan.BoundsSpec
}

func newRangeProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*RangeOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &RangeProcedureSpec{
		Bounds: plan.BoundsSpec{
			Start: spec.Start,
			Stop:  spec.Stop,
		},
	}, nil
}

func (s *RangeProcedureSpec) Kind() plan.ProcedureKind {
	return RangeKind
}
func (s *RangeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RangeProcedureSpec)
	ns.Bounds = s.Bounds
	return ns
}

func (s *RangeProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    FromKind,
		Through: []plan.ProcedureKind{GroupKind, LimitKind, FilterKind},
	}
}
func (s *RangeProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*FromProcedureSpec)
	if selectSpec.BoundsSet {
		// Example case where this matters
		//    var data = select(database: "mydb")
		//    var past = data.range(start:-2d,stop:-1d)
		//    var current = data.range(start:-1d,stop:now)
		root = dup()
		selectSpec = root.Spec.(*FromProcedureSpec)
		selectSpec.BoundsSet = false
		selectSpec.Bounds = plan.BoundsSpec{}
		return
	}
	selectSpec.BoundsSet = true
	selectSpec.Bounds = s.Bounds
}

func (s *RangeProcedureSpec) TimeBounds() plan.BoundsSpec {
	return s.Bounds
}
