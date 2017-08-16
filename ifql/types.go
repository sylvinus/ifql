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
	FieldKind
	NumKinds int = iota
)

type Function struct {
	Name     string         `json:"name,omitempty"`
	Args     []*FunctionArg `json:"args,omitempty"`
	Children []*Function    `json:"children,omitempty"`
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
	Name string `json:"name,omitempty"`
	Arg  Arg    `json:"arg,omitempty"`
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
	String string `json:"string,omitempty"`
}

func (s *StringLiteral) Type() ArgKind {
	return StringKind
}

func (s *StringLiteral) Value() interface{} {
	return s.String
}

type Duration struct {
	Dur time.Duration `json:"dur,omitempty"`
}

func (d *Duration) Type() ArgKind {
	return DurationKind
}

func (d *Duration) Value() interface{} {
	return d.Dur
}

type DateTime struct {
	Date time.Time `json:"date,omitempty"`
}

func (d *DateTime) Type() ArgKind {
	return DateTimeKind
}

func (d *DateTime) Value() interface{} {
	return d.Date
}

type Number struct {
	Val float64 `json:"val,omitempty"`
}

func (n *Number) Type() ArgKind {
	return NumberKind
}

func (n *Number) Value() interface{} {
	return n.Val
}

type WhereExpr struct {
	Node *storage.Node `json:"node,omitempty"`
}

func (w *WhereExpr) Type() ArgKind {
	return ExprKind
}

func (w *WhereExpr) Value() interface{} {
	return w.Node
}

// Field represents a value associated with a series
type Field struct {
}

func (f *Field) Type() ArgKind {
	return FieldKind
}

func (f *Field) Value() interface{} {
	return "_field"
}
