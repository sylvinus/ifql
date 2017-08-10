package ifql

import (
	"time"

	"github.com/influxdata/ifql/query/execute/storage"
)

type Arg interface {
	Type() ArgKind
	Value() interface{}
}

type ArgKind int

const (
	DateTimeKind ArgKind = iota
	DurationKind
	ExprKind
	NumberKind
	StringKind
	NumKinds int = iota
)

type Function struct {
	Name     string
	Args     []*FunctionArg
	Children []*Function
}

func NewFunction(name, args, children interface{}) (*Function, error) {
	chain := []*Function{}
	if children != nil {
		for _, child := range toIfaceSlice(children) {
			chain = append(chain, child.(*Function))
		}
	}
	funcArgs := []*FunctionArg{}
	if args != nil {
		for _, arg := range toIfaceSlice(args) {
			funcArgs = append(funcArgs, arg.([]*FunctionArg)...)
		}
	}
	return &Function{
		Name:     name.(string),
		Args:     funcArgs,
		Children: chain,
	}, nil
}

type FunctionArg struct {
	Name string
	Arg  Arg
}

func NewFunctionArgs(first, rest interface{}) ([]*FunctionArg, error) {
	args := []*FunctionArg{first.(*FunctionArg)}
	if rest != nil {
		for _, arg := range toIfaceSlice(rest) {
			args = append(args, arg.(*FunctionArg))
		}
	}
	return args, nil
}

type StringLiteral struct {
	String string
}

func (s *StringLiteral) Type() ArgKind {
	return StringKind
}

func (s *StringLiteral) Value() interface{} {
	return s.String
}

type Duration struct {
	Dur time.Duration
}

func (d *Duration) Type() ArgKind {
	return DurationKind
}

func (d *Duration) Value() interface{} {
	return d.Dur
}

type DateTime struct {
	Date time.Time
}

func (d *DateTime) Type() ArgKind {
	return DateTimeKind
}

func (d *DateTime) Value() interface{} {
	return d.Date
}

type Number struct {
	Val float64
}

func (n *Number) Type() ArgKind {
	return NumberKind
}

func (n *Number) Value() interface{} {
	return n.Val
}

type WhereExpr struct {
	node *storage.Node
}

func (w *WhereExpr) Type() ArgKind {
	return ExprKind
}

func (w *WhereExpr) Value() interface{} {
	return w.node
}
