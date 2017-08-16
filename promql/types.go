package promql

import (
	"strconv"
	"time"
)

type ArgKind int

const (
	IdentifierKind ArgKind = iota
	DurationKind
	ExprKind
	NumberKind
	StringKind
	SelectorKind
	NumKinds int = iota
)

type Arg interface {
	Type() ArgKind
	Value() interface{}
}

type Identifier struct {
	Name string `json:"name,omitempty"`
}

func (id *Identifier) Type() ArgKind {
	return IdentifierKind
}

func (id *Identifier) Value() interface{} {
	return id.Name
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

type Number struct {
	Val float64 `json:"val,omitempty"`
}

func (n *Number) Type() ArgKind {
	return NumberKind
}

func (n *Number) Value() interface{} {
	return n.Val
}

func NewNumber(val string) (*Number, error) {
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &Number{num}, nil
}

// MatchKind is an enum for label matching types.
type MatchKind int

// Possible MatchKinds.
const (
	Equal MatchKind = iota
	NotEqual
	RegexMatch
	RegexNoMatch
)

type LabelMatcher struct {
	Name  string
	Kind  MatchKind
	Value Arg
}

func NewLabelMatcher(ident *Identifier, kind MatchKind, value Arg) (*LabelMatcher, error) {
	return &LabelMatcher{
		Name:  ident.Name,
		Kind:  kind,
		Value: value,
	}, nil
}

func NewLabelMatches(first *LabelMatcher, rest interface{}) ([]*LabelMatcher, error) {
	matches := []*LabelMatcher{first}
	for _, m := range toIfaceSlice(rest) {
		if match, ok := m.(*LabelMatcher); ok {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

type Selector struct {
	Name          string
	Range         time.Duration
	Offset        time.Duration
	LabelMatchers []*LabelMatcher
}

func (s *Selector) Type() ArgKind {
	return SelectorKind
}

func (s *Selector) Value() interface{} {
	return s.Name // TODO: Change to AST
}

func NewSelector(metric *Identifier, block, rng, offset interface{}) (*Selector, error) {
	sel := &Selector{
		Name: metric.Name,
	}

	if rng != nil {
		sel.Range = rng.(time.Duration)
	}

	if offset != nil {
		sel.Offset = offset.(time.Duration)
	}

	return sel, nil
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
