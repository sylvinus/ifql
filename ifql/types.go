package ifql

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/ifql/ast"
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

func howdy(name string, args, children interface{}) {

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

func call(name, args, children interface{}, text []byte, pos position) (*ast.CallExpression, error) {
	if children == nil {
		return &ast.CallExpression{
			Callee:    name.(*ast.Identifier),
			Arguments: []ast.Expression{args.(*ast.ObjectExpression)},
		}, nil
	}
	/*
		for _, child := range toIfaceSlice(children) {
			chain = append(chain, child.(*Function))
		}
	*/
	return nil, nil
}

func object(first, rest interface{}, text []byte, pos position) (*ast.ObjectExpression, error) {
	props := []*ast.Property{first.(*ast.Property)}
	if rest != nil {
		for _, prop := range toIfaceSlice(rest) {
			props = append(props, prop.(*ast.Property))
		}
	}
	return &ast.ObjectExpression{
		Properties: props,
		BaseNode:   base(text, pos),
	}, nil
}

func property(key, value interface{}, text []byte, pos position) (*ast.Property, error) {
	return &ast.Property{
		Key:      key.(*ast.Identifier),
		Value:    value.(ast.Expression),
		BaseNode: base(text, pos),
	}, nil
}

func identifier(text []byte, pos position) (*ast.Identifier, error) {
	return &ast.Identifier{
		Name:     string(text),
		BaseNode: base(text, pos),
	}, nil
}

func logicalExpression(head, tails interface{}) (ast.Expression, error) {
	res := head.(ast.Expression)
	for _, tail := range toIfaceSlice(tails) {
		right := toIfaceSlice(tail)
		res = &ast.LogicalExpression{
			Left:     res,
			Right:    right[3].(ast.Expression),
			Operator: right[1].(ast.LogicalOperatorKind),
		}
	}
	return res, nil
}

func logicalOp(text []byte) (ast.LogicalOperatorKind, error) {
	return ast.LogicalOperatorLookup(strings.ToLower(string(text))), nil
}

func binaryExpression(head, tails interface{}) (ast.Expression, error) {
	res := head.(ast.Expression)
	for _, tail := range toIfaceSlice(tails) {
		right := toIfaceSlice(tail)
		res = &ast.BinaryExpression{
			Left:     res,
			Right:    right[3].(ast.Expression),
			Operator: right[1].(ast.OperatorKind),
		}
	}
	return res, nil
}

func binaryOp(text []byte) (ast.OperatorKind, error) {
	return ast.OperatorLookup(strings.ToLower(string(text))), nil
}

func stringLiteral(text []byte, pos position) (*ast.StringLiteral, error) {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		return nil, err
	}
	return &ast.StringLiteral{
		BaseNode: base(text, pos),
		Value:    s,
	}, nil
}

func numberLiteral(text []byte, pos position) (*ast.NumberLiteral, error) {
	n, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return nil, err
	}
	return &ast.NumberLiteral{
		BaseNode: base(text, pos),
		Value:    n,
	}, nil
}

func fieldLiteral(text []byte, pos position) (*ast.FieldLiteral, error) {
	return &ast.FieldLiteral{
		BaseNode: base(text, pos),
	}, nil
}

func regexLiteral(chars interface{}, text []byte, pos position) (*ast.RegexpLiteral, error) {
	var regex string
	for _, char := range toIfaceSlice(chars) {
		regex += char.(string)
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return &ast.RegexpLiteral{
		BaseNode: base(text, pos),
		Value:    r,
	}, nil
}

func durationLiteral(text []byte, pos position) (*ast.DurationLiteral, error) {
	d, err := time.ParseDuration(string(text))
	if err != nil {
		return nil, err
	}
	return &ast.DurationLiteral{
		BaseNode: base(text, pos),
		Value:    d,
	}, nil
}

func datetime(text []byte, pos position) (*ast.DateTimeLiteral, error) {
	t, err := time.Parse(time.RFC3339Nano, string(text))
	if err != nil {
		return nil, err
	}
	return &ast.DateTimeLiteral{
		BaseNode: base(text, pos),
		Value:    t,
	}, nil
}

func base(text []byte, pos position) *ast.BaseNode {
	return &ast.BaseNode{
		Loc: &ast.SourceLocation{
			Start: ast.Position{
				Line:   pos.line,
				Column: pos.col,
			},
			End: ast.Position{
				Line:   pos.line,
				Column: pos.col + len(text),
			},
			Source: source(text),
		},
	}
}

func source(text []byte) *string {
	str := string(text)
	return &str
}
