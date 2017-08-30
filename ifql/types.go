package ifql

import (
	"regexp"
	"time"
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
	RegexKind
	NumKinds int = iota
)

// TODO: Convert to QuerySpec
type Function struct {
	Name     string         `json:"name,omitempty"`
	Args     []*FunctionArg `json:"args,omitempty"`
	Children []*Function    `json:"children,omitempty"`
}

func NewFunction(name string, args, children interface{}) (*Function, error) {
	chain := []*Function{}
	if children != nil {
		for _, child := range toIfaceSlice(children) {
			chain = append(chain, child.(*Function))
		}
	}
	funcArgs := []*FunctionArg{}
	if args != nil {
		funcArgs = args.([]*FunctionArg)
	}
	return &Function{
		Name:     name,
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

// Regex represents a regular expression function argument
type Regex struct {
	Regexp *regexp.Regexp
}

// NewRegex compiles the regular expression and returns the regex
func NewRegex(chars interface{}) (*Regex, error) {
	var regex string
	for _, char := range toIfaceSlice(chars) {
		regex += char.(string)
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return &Regex{r}, nil
}

func (r *Regex) Type() ArgKind {
	return RegexKind
}

func (r *Regex) Value() interface{} {
	return r.Regexp
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
	Expr *BinaryExpression `json:"expr,omitempty"`
}

func (w *WhereExpr) Type() ArgKind {
	return ExprKind
}

func (w *WhereExpr) Value() interface{} {
	return w.Expr
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

// Update left and right to be expr interfaces... add expr interface to all the correct types
type BinaryExpression struct {
	Left     interface{} `json:"left,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Right    interface{} `json:"right,omitempty"`
}

func NewBinaryExpression(head, tails interface{}) (interface{}, error) {
	res := head
	for _, tail := range toIfaceSlice(tails) {
		right := toIfaceSlice(tail)
		res = &BinaryExpression{
			Left:     res,
			Right:    right[3],
			Operator: right[1].(string),
		}
	}
	return res, nil
}

/*
func (b *BinaryExpression) String() string {
	res := ""
	switch l := b.Left.(type) {
	case *BinaryExpression:
		res += "(" + l.String() + ")"
	case *StringLiteral:
		res += `"` + l.String + `"`
	case *Number:
		res += fmt.Sprintf("%f", l.Val)
	case *Field:
		res += "$"
	}
	res += " " + b.Operator.(string) + " "

	switch r := b.Right.(type) {
	case *BinaryExpression:
		res += "(" + r.String() + ")"
	case *StringLiteral:
		res += `"` + r.String + `"`
	case *Number:
		res += fmt.Sprintf("%f", r.Val)
	case *Field:
		res += "$"
	}
	return res
}*/
