package functions

import (
	"errors"
	"fmt"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const WhereKind = "where"

type WhereOpSpec struct {
	Exp expression.Node `json:"exp"`
}

func init() {
	ifql.RegisterFunction(WhereKind, createWhereOpSpec)
	query.RegisterOpSpec(WhereKind, newWhereOp)
	plan.RegisterProcedureSpec(WhereKind, newWhereProcedure, WhereKind)
	// TODO register a where transformation. Currently where is only supported if it is pushed down into a select procedure.
	//execute.RegisterTransformation(WhereKind, createWhereTransformation)
}

func createWhereOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	expValue, ok := args["exp"]
	if !ok {
		return nil, errors.New(`where function requires an argument "exp"`)
	}
	if expValue.Type != ifql.TExpression {
		return nil, fmt.Errorf(`where function argument "exp" must be an expression, got %v`, expValue.Type)
	}

	return &WhereOpSpec{
		Exp: expValue.Value.(expression.Node),
	}, nil
}
func newWhereOp() query.OperationSpec {
	return new(WhereOpSpec)
}

func (s *WhereOpSpec) Kind() query.OperationKind {
	return WhereKind
}

type WhereProcedureSpec struct {
	Exp expression.Node
}

func newWhereProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*WhereOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &WhereProcedureSpec{
		Exp: spec.Exp,
	}, nil
}

func (s *WhereProcedureSpec) Kind() plan.ProcedureKind {
	return WhereKind
}

func (s *WhereProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: []plan.ProcedureKind{LimitKind, RangeKind},
	}
}
func (s *WhereProcedureSpec) PushDown(root *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.WhereSet {
		// TODO: create copy of select spec and set new where expression
	}
	selectSpec.WhereSet = true
	p, err := execute.ExpressionToStoragePredicate(s.Exp)
	if err != nil {
		//TODO(nathanielc): Handle this error
		panic(err)
	}
	selectSpec.Where = p
}
