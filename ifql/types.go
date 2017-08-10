package ifql

import (
	"fmt"
	"strings"
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

// *github.com/influxdata/ifql/query/execute/storage.Predicate
type WhereExpr struct {
	node Node
}

func (w *WhereExpr) Type() ArgKind {
	return ExprKind
}

func (w *WhereExpr) Value() interface{} {
	return w.node
}

type NodeKind int
type ComparisonKind int
type LogicalKind int

type Node struct {
	Kind       NodeKind
	Comparison ComparisonKind
	Logical    LogicalKind
	Children   []Node
	Value      Arg
}

func NewComparisonOperator(text []byte) (storage.Node_Comparison, error) {
	op := strings.ToLower(string(text))
	// "<=" / "<" / ">=" / ">" / "=" / "!=" / "startsWith"i / "in"i / "not empty"i / "empty"i
	switch op {
	case "=":
		return storage.ComparisonEqual, nil
	case "!=":
		return storage.ComparisonNotEqual, nil
	case "startswith":
		return storage.ComparisonStartsWith, nil
	case "<=", "<", ">=", ">", "in", "not empty", "empty":
		return 0, fmt.Errorf("Unimplemented comparison operator %s", op)
	default:
		return 0, fmt.Errorf("Unknown comparison operator %s", op)
	}
}

func NewLogicalOperator(text []byte) (storage.Node_Logical, error) {
	op := strings.ToLower(string(text))
	if op == "and" {
		return storage.LogicalAnd, nil
	} else if op == "or" {
		return storage.LogicalOr, nil
	}
	return 0, fmt.Errorf("Unknown logical operator %s", op)
}
