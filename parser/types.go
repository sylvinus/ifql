package parser

import "time"

type StringLiteral struct {
	value string
}

func (s *StringLiteral) Type() ArgType {
	return STRING
}

func (s *StringLiteral) Value() interface{} {
	return n.value
}

type Duration struct {
	value time.Duration
}

func (d *Duration) Type() ArgType {
	return DURATION
}

func (d *Duration) Value() interface{} {
	return n.value
}

type DateTime struct {
	value time.Time
}

func (d *DateTime) Type() ArgType {
	return DATETIME
}

func (d *DateTime) Value() interface{} {
	return n.value
}

type Number struct {
	value float64
}

func (n *Number) Type() ArgType {
	return NUMBER
}

func (n *Number) Value() interface{} {
	return n.value
}

type WhereExpr struct {
	node Node
}

func (w *WhereExpr) Type() ArgType {
	return EXPR
}

func (w *WhereExpr) Value() interface{} {
	return w.Node
}

type Node struct {
	Type       NodeType
	Comparison ComparisonType
	Logical    LogicalType
	Children   []Node
	Value      Arg
}
