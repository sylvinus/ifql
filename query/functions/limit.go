package functions

import (
	"fmt"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

const LimitKind = "limit"

type LimitOpSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func init() {
	query.RegisterOpSpec(LimitKind, newLimitOp)
	plan.RegisterProcedureSpec(LimitKind, newLimitProcedure, LimitKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	//execute.RegisterTransformation(LimitKind, createLimitTransformation)
}

func newLimitOp() query.OperationSpec {
	return new(LimitOpSpec)
}

func (s *LimitOpSpec) Kind() query.OperationKind {
	return LimitKind
}

type LimitProcedureSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func newLimitProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LimitOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &LimitProcedureSpec{
		Limit:  spec.Limit,
		Offset: spec.Offset,
	}, nil
}

func (s *LimitProcedureSpec) Kind() plan.ProcedureKind {
	return LimitKind
}

func (s *LimitProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: []plan.ProcedureKind{RangeKind, WhereKind},
	}
}
func (s *LimitProcedureSpec) PushDown(root *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.LimitSet {
		// TODO: create copy of select spec and set new limit
	}
	selectSpec.LimitSet = true
	selectSpec.Limit = s.Limit
	selectSpec.Offset = s.Offset
}
