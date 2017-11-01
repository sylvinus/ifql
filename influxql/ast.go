package influxql

import (
	"errors"
	"strings"

	"github.com/influxdata/influxql"
)

func isConjunction(op influxql.Token) bool {
	return op == influxql.AND || op == influxql.OR
}

func isLogical(op influxql.Token) bool {
	return op >= influxql.EQ && op <= influxql.GTE
}

func isValidTimeLogical(op influxql.Token) bool {
	return op == influxql.EQ ||
		op == influxql.LT ||
		op == influxql.LTE ||
		op == influxql.GT ||
		op == influxql.GTE
}

type timeRangeVisitor struct {
	ops   []influxql.Token
	hasOR bool
	err   error
}

func (v *timeRangeVisitor) push(t influxql.Token) {
	v.ops = append(v.ops, t)
}

func (v *timeRangeVisitor) pop(t influxql.Token) {
	v.ops = append(v.ops, t)
}

func (v *timeRangeVisitor) Visit(node influxql.Node) influxql.Visitor {
	switch n := node.(type) {
	case *influxql.BinaryExpr:
		if n.Op == influxql.OR {
			v.push(n.Op)
			var old bool
			old, v.hasOR = v.hasOR, true
			influxql.Walk(v, n.LHS)
			influxql.Walk(v, n.RHS)
			v.hasOR = old
			if v.err != nil {
				return nil
			}
		} else if isLogical(n.Op) {
			if ref, ok := n.LHS.(*influxql.VarRef); ok && strings.ToLower(ref.Val) == "time" && v.hasOR {
				v.err = errors.New("time cannot be combined with OR operator")
				return nil
			}
		}
	}

	return v
}

func ValidateTimeRangeExpr(expr influxql.Expr) error {
	var v timeRangeVisitor
	influxql.Walk(&v, expr)
	return v.err
}
