package expression

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	MemberReference
)

func (t Type) String() string {
	s, err := t.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(s)
}
func (t Type) MarshalText() ([]byte, error) {
	switch t {
	case Binary:
		return []byte("binary"), nil
	case Unary:
		return []byte("unary"), nil
	case StringLiteral:
		return []byte("stringLiteral"), nil
	case IntegerLiteral:
		return []byte("integerLiteral"), nil
	case BooleanLiteral:
		return []byte("booleanLiteral"), nil
	case FloatLiteral:
		return []byte("floatLiteral"), nil
	case DurationLiteral:
		return []byte("durationLiteral"), nil
	case TimeLiteral:
		return []byte("timeLiteral"), nil
	case RegexpLiteral:
		return []byte("regexpLiteral"), nil
	case Reference:
		return []byte("reference"), nil
	case MemberReference:
		return []byte("memberReference"), nil
	default:
		return nil, fmt.Errorf("unknown type %d", int(t))
	}
}

func (t *Type) UnmarshalText(data []byte) error {
	switch string(data) {
	case "binary":
		*t = Binary
	case "unary":
		*t = Unary
	case "stringLiteral":
		*t = StringLiteral
	case "integerLiteral":
		*t = IntegerLiteral
	case "booleanLiteral":
		*t = BooleanLiteral
	case "floatLiteral":
		*t = FloatLiteral
	case "durationLiteral":
		*t = DurationLiteral
	case "timeLiteral":
		*t = TimeLiteral
	case "regexpLiteral":
		*t = RegexpLiteral
	case "reference":
		*t = Reference
	case "memberReference":
		*t = MemberReference
	default:
		return fmt.Errorf("unknown type %q", string(data))
	}
	return nil
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

func (n *BinaryNode) MarshalJSON() ([]byte, error) {
	type Alias BinaryNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

func (n *BinaryNode) UnmarshalJSON(data []byte) error {
	type binaryNode struct {
		Operator Operator         `json:"operator"`
		Left     *json.RawMessage `json:"left"`
		Right    *json.RawMessage `json:"right"`
	}
	bn := binaryNode{}
	if err := json.Unmarshal(data, &bn); err != nil {
		return err
	}
	n.Operator = bn.Operator
	left, err := unmarshalNode(bn.Left)
	if err != nil {
		return err
	}
	n.Left = left
	right, err := unmarshalNode(bn.Right)
	if err != nil {
		return err
	}
	n.Right = right
	return nil
}

type UnaryNode struct {
	Operator Operator `json:"operator"`
	Node     Node     `json:"node"`
}

func (*UnaryNode) Type() Type {
	return Unary
}

func (n *UnaryNode) MarshalJSON() ([]byte, error) {
	type Alias UnaryNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

func (n *UnaryNode) UnmarshalJSON(data []byte) error {
	type unaryNode struct {
		Operator Operator         `json:"operator"`
		Node     *json.RawMessage `json:"node"`
	}
	un := unaryNode{}
	if err := json.Unmarshal(data, &un); err != nil {
		return err
	}
	n.Operator = un.Operator
	node, err := unmarshalNode(un.Node)
	if err != nil {
		return err
	}
	n.Node = node
	return nil
}

type StringLiteralNode struct {
	Value string `json:"value"`
}

func (*StringLiteralNode) Type() Type {
	return StringLiteral
}

func (n *StringLiteralNode) MarshalJSON() ([]byte, error) {
	type Alias StringLiteralNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type IntegerLiteralNode struct {
	Value int64 `json:"value"`
}

func (*IntegerLiteralNode) Type() Type {
	return IntegerLiteral
}

type integerLiteralNodeJSON struct {
	Type  Type   `json:"type"`
	Value string `json:"value"`
}

func (n *IntegerLiteralNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(integerLiteralNodeJSON{
		Type:  n.Type(),
		Value: strconv.FormatInt(n.Value, 10),
	})
}

func (n *IntegerLiteralNode) UnmarshalJSON(data []byte) error {
	in := integerLiteralNodeJSON{}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	i, err := strconv.ParseInt(in.Value, 10, 64)
	if err != nil {
		return err
	}
	n.Value = i
	return nil
}

type BooleanLiteralNode struct {
	Value bool `json:"value"`
}

func (*BooleanLiteralNode) Type() Type {
	return BooleanLiteral
}

func (n *BooleanLiteralNode) MarshalJSON() ([]byte, error) {
	type Alias BooleanLiteralNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type FloatLiteralNode struct {
	Value float64 `json:"value"`
}

func (*FloatLiteralNode) Type() Type {
	return FloatLiteral
}

func (n *FloatLiteralNode) MarshalJSON() ([]byte, error) {
	type Alias FloatLiteralNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type DurationLiteralNode struct {
	Value time.Duration `json:"value"`
}

func (*DurationLiteralNode) Type() Type {
	return DurationLiteral
}

type durationLiteralNodeJSON struct {
	Type  Type   `json:"type"`
	Value string `json:"value"`
}

func (n *DurationLiteralNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(durationLiteralNodeJSON{
		Type:  n.Type(),
		Value: n.Value.String(),
	})
}
func (n *DurationLiteralNode) UnmarshalJSON(data []byte) error {
	dn := durationLiteralNodeJSON{}
	if err := json.Unmarshal(data, &dn); err != nil {
		return err
	}
	d, err := time.ParseDuration(dn.Value)
	if err != nil {
		return err
	}
	n.Value = d
	return nil
}

type TimeLiteralNode struct {
	Value time.Time `json:"value"`
}

func (*TimeLiteralNode) Type() Type {
	return TimeLiteral
}

func (n *TimeLiteralNode) MarshalJSON() ([]byte, error) {
	type Alias TimeLiteralNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type RegexpLiteralNode struct {
	Value string `json:"value"`
}

func (*RegexpLiteralNode) Type() Type {
	return RegexpLiteral
}

func (n *RegexpLiteralNode) MarshalJSON() ([]byte, error) {
	type Alias RegexpLiteralNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type ReferenceNode struct {
	// Name is the name of the object being referenced
	Name string `json:"name"`
}

func (*ReferenceNode) Type() Type {
	return Reference
}

func (n *ReferenceNode) MarshalJSON() ([]byte, error) {
	type Alias ReferenceNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

type MemberReferenceNode struct {
	// Object is the object whose property is being referenced.
	Object Node `json:"object"`
	// Property is the name of the property of object being referenced
	Property string `json:"property"`
}

func (*MemberReferenceNode) Type() Type {
	return MemberReference
}

func (n *MemberReferenceNode) MarshalJSON() ([]byte, error) {
	type Alias MemberReferenceNode
	var raw = &struct {
		Type Type `json:"type"`
		*Alias
	}{
		Type:  n.Type(),
		Alias: (*Alias)(n),
	}
	return json.Marshal(raw)
}

func (n *MemberReferenceNode) UnmarshalJSON(data []byte) error {
	type memberReferenceNode struct {
		Object   *json.RawMessage `json:"object"`
		Property string           `json:"property"`
	}
	mn := memberReferenceNode{}
	if err := json.Unmarshal(data, &mn); err != nil {
		return err
	}
	n.Property = mn.Property

	obj, err := unmarshalNode(mn.Object)
	if err != nil {
		return err
	}
	n.Object = obj
	return nil
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
	s, err := o.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(s)
}
func (o Operator) MarshalText() ([]byte, error) {
	switch o {
	case MultiplicationOperator:
		return []byte("*"), nil
	case DivisionOperator:
		return []byte("/"), nil
	case AdditionOperator:
		return []byte("+"), nil
	case SubtractionOperator:
		return []byte("-"), nil
	case LessThanEqualOperator:
		return []byte("<="), nil
	case LessThanOperator:
		return []byte("<"), nil
	case GreaterThanEqualOperator:
		return []byte(">="), nil
	case GreaterThanOperator:
		return []byte(">"), nil
	case StartsWithOperator:
		return []byte("startsWith"), nil
	case InOperator:
		return []byte("in"), nil
	case NotEmptyOperator:
		return []byte("notEmpty"), nil
	case EmptyOperator:
		return []byte("empty"), nil
	case NotOperator:
		return []byte("!"), nil
	case EqualOperator:
		return []byte("=="), nil
	case NotEqualOperator:
		return []byte("!="), nil
	case RegexpMatchOperator:
		return []byte("regexpMatch"), nil
	case RegexpNotMatchOperator:
		return []byte("regexpNotMatch"), nil
	case AndOperator:
		return []byte("and"), nil
	case OrOperator:
		return []byte("or"), nil
	default:
		return nil, fmt.Errorf("unknown operator %d", int(o))
	}
}
func (o *Operator) UnmarshalText(data []byte) error {
	switch string(data) {
	case "*":
		*o = MultiplicationOperator
	case "/":
		*o = DivisionOperator
	case "+":
		*o = AdditionOperator
	case "-":
		*o = SubtractionOperator
	case "<=":
		*o = LessThanEqualOperator
	case "<":
		*o = LessThanOperator
	case ">=":
		*o = GreaterThanEqualOperator
	case ">":
		*o = GreaterThanOperator
	case "startsWith":
		*o = StartsWithOperator
	case "in":
		*o = InOperator
	case "notEmpty":
		*o = NotEmptyOperator
	case "empty":
		*o = EmptyOperator
	case "!":
		*o = NotOperator
	case "==":
		*o = EqualOperator
	case "!=":
		*o = NotEqualOperator
	case "regexpMatch":
		*o = RegexpMatchOperator
	case "regexpNotMatch":
		*o = RegexpNotMatchOperator
	case "and":
		*o = AndOperator
	case "or":
		*o = OrOperator
	default:
		return fmt.Errorf("unknown operator %q", string(data))
	}
	return nil
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
	case *MemberReferenceNode:
		if err := Walk(node.Object, f); err != nil {
			return err
		}
	}
	return nil
}

type Expression struct {
	Root Node `json:"root"`
	// Params is a list of names that are provided as parameters to the expression.
	Params []string `json:"params"`
}

func (e *Expression) UnmarshalJSON(data []byte) error {
	type expression struct {
		Root   *json.RawMessage `json:"root"`
		Params []string         `json:"params"`
	}
	expr := expression{}
	if err := json.Unmarshal(data, &expr); err != nil {
		return err
	}
	e.Params = expr.Params

	node, err := unmarshalNode(expr.Root)
	if err != nil {
		return err
	}
	e.Root = node
	return nil
}

func unmarshalNode(msg *json.RawMessage) (Node, error) {
	type typeRawMessage struct {
		Type Type `json:"type"`
	}
	if msg == nil {
		return nil, nil
	}

	typ := typeRawMessage{}
	if err := json.Unmarshal(*msg, &typ); err != nil {
		return nil, err
	}

	var node Node
	switch typ.Type {
	case Binary:
		node = new(BinaryNode)
	case Unary:
		node = new(UnaryNode)
	case StringLiteral:
		node = new(StringLiteralNode)
	case IntegerLiteral:
		node = new(IntegerLiteralNode)
	case BooleanLiteral:
		node = new(BooleanLiteralNode)
	case FloatLiteral:
		node = new(FloatLiteralNode)
	case DurationLiteral:
		node = new(DurationLiteralNode)
	case TimeLiteral:
		node = new(TimeLiteralNode)
	case RegexpLiteral:
		node = new(RegexpLiteralNode)
	case Reference:
		node = new(ReferenceNode)
	case MemberReference:
		node = new(MemberReferenceNode)
	default:
		return nil, fmt.Errorf("unknown type %v", typ.Type)
	}

	if err := json.Unmarshal(*msg, node); err != nil {
		return nil, err
	}
	return node, nil
}
