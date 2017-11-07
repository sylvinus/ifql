package functions

import (
	"errors"
	"fmt"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

const LimitKind = "limit"

type LimitOpSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func init() {
	ifql.RegisterFunction(LimitKind, createLimitOpSpec)
	query.RegisterOpSpec(LimitKind, newLimitOp)
	plan.RegisterProcedureSpec(LimitKind, newLimitProcedure, LimitKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	//execute.RegisterTransformation(LimitKind, createLimitTransformation)
}

func createLimitOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(LimitOpSpec)

	limitValue, ok := args["limit"]
	if !ok {
		return nil, errors.New(`limit function requires argument "limit"`)
	}

	if limitValue.Type != ifql.TInt {
		return nil, fmt.Errorf(`limit argument "limit" must be an integer, got %v`, limitValue.Type)
	}
	spec.Limit = limitValue.Value.(int64)

	if offsetValue, ok := args["offset"]; ok {
		if offsetValue.Type != ifql.TInt {
			return nil, fmt.Errorf(`limit argument "offset" must be an integer, got %v`, offsetValue.Type)
		}
		spec.Offset = offsetValue.Value.(int64)
	}

	return spec, nil
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
func (s *LimitProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LimitProcedureSpec)
	ns.Limit = s.Limit
	ns.Offset = s.Offset
	return ns
}

func (s *LimitProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: []plan.ProcedureKind{RangeKind, WhereKind},
	}
}
func (s *LimitProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.LimitSet {
		root = dup()
		selectSpec = root.Spec.(*SelectProcedureSpec)
		selectSpec.LimitSet = false
		selectSpec.Limit = 0
		selectSpec.Offset = 0
		return
	}
	selectSpec.LimitSet = true
	selectSpec.Limit = s.Limit
	selectSpec.Offset = s.Offset
}
