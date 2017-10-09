package expression

import (
	"fmt"
	"time"
)

type Type int

const (
	Binary Type = iota
	Unary
	StringLiteral
	IntegerLiteral
	BooleanLiteral
	FloatLiteral
	DurationLiteral
	TimeLiteral
	RegexpLiteral
	Reference
)

func (t Type) String() string {
	switch t {
	case Binary:
		return "binary"
	case Unary:
		return "unary"
	case StringLiteral:
		return "stringLiteral"
	case IntegerLiteral:
		return "integerLiteral"
	case BooleanLiteral:
		return "booleanLiteral"
	case FloatLiteral:
		return "floatLiteral"
	case DurationLiteral:
		return "durationLiteral"
	case TimeLiteral:
		return "timeLiteral"
	case RegexpLiteral:
		return "regexpLiteral"
	case Reference:
		return "reference"
	default:
		return fmt.Sprintf("unknown type %d", int(t))
	}
}

type Node interface {
	Type() Type
}

type BinaryNode struct {
	Operator Operator `json:"operator"`
	Left     Node     `json:"left"`
	Right    Node     `json:"right"`
}

func (*BinaryNode) Type() Type {
	return Binary
}

type UnaryNode struct {
	Operator Operator `json:"operator"`
	Node     Node     `json:"node"`
}

func (*UnaryNode) Type() Type {
	return Unary
}

type StringLiteralNode struct {
	Value string `json:"value"`
}

func (*StringLiteralNode) Type() Type {
	return StringLiteral
}

type IntegerLiteralNode struct {
	Value int64 `json:"value"`
}

func (*IntegerLiteralNode) Type() Type {
	return IntegerLiteral
}

type BooleanLiteralNode struct {
	Value bool `json:"value"`
}

func (*BooleanLiteralNode) Type() Type {
	return BooleanLiteral
}

type FloatLiteralNode struct {
	Value float64 `json:"value"`
}

func (*FloatLiteralNode) Type() Type {
	return FloatLiteral
}

type DurationLiteralNode struct {
	Value time.Duration `json:"value"`
}

func (*DurationLiteralNode) Type() Type {
	return DurationLiteral
}

type TimeLiteralNode struct {
	Value time.Time `json:"value"`
}

func (*TimeLiteralNode) Type() Type {
	return TimeLiteral
}

type RegexpLiteralNode struct {
	Value string `json:"value"`
}

func (*RegexpLiteralNode) Type() Type {
	return RegexpLiteral
}

type ReferenceNode struct {
	// Name is the name of the item being referenced
	Name string `json:"name"`
	// Kind is any kind for the reference, can be used to indicate type information about the item being referenced.
	Kind string `json:"kind"`
}

func (*ReferenceNode) Type() Type {
	return Reference
}

type Operator int

const (
	MultiplicationOperator Operator = iota
	DivisionOperator
	AdditionOperator
	SubtractionOperator
	LessThanEqualOperator
	LessThanOperator
	GreaterThanEqualOperator
	GreaterThanOperator
	StartsWithOperator
	InOperator
	NotEmptyOperator
	EmptyOperator
	EqualOperator
	NotEqualOperator
	RegexpMatchOperator
	RegexpNotMatchOperator

	AndOperator
	OrOperator
	NotOperator
)

func (o Operator) String() string {
	switch o {
	case MultiplicationOperator:
		return "multiplicationOperator"
	case DivisionOperator:
		return "divisionOperator"
	case AdditionOperator:
		return "additionOperator"
	case SubtractionOperator:
		return "subtractionOperator"
	case LessThanEqualOperator:
		return "lessThanEqualOperator"
	case LessThanOperator:
		return "lessThanOperator"
	case GreaterThanEqualOperator:
		return "greaterThanEqualOperator"
	case GreaterThanOperator:
		return "greaterThanOperator"
	case StartsWithOperator:
		return "startsWithOperator"
	case InOperator:
		return "inOperator"
	case NotEmptyOperator:
		return "notEmptyOperator"
	case EmptyOperator:
		return "emptyOperator"
	case EqualOperator:
		return "equalOperator"
	case NotEqualOperator:
		return "notEqualOperator"
	case RegexpMatchOperator:
		return "regexpMatchOperator"
	case RegexpNotMatchOperator:
		return "regexpNotMatchOperator"
	case AndOperator:
		return "andOperator"
	case OrOperator:
		return "orOperator"
	case NotOperator:
		return "notOperator"
	default:
		return fmt.Sprintf("unknown operator %d", int(o))
	}
}

func Walk(n Node, f func(Node) error) error {
	if err := f(n); err != nil {
		return err
	}
	switch node := n.(type) {
	case *BinaryNode:
		if err := Walk(node.Left, f); err != nil {
			return err
		}
		if err := Walk(node.Right, f); err != nil {
			return err
		}
	case *UnaryNode:
		if err := Walk(node.Node, f); err != nil {
			return err
		}
	}
	return nil
}
