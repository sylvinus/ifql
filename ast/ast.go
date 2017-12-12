package ast

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Position represents a specific location in the source
type Position struct {
	Line   int `json:"line"`   // Line is the line in the source marked by this position
	Column int `json:"column"` // Column is the column in the source marked by this position
}

// SourceLocation represents the location of a node in the AST
type SourceLocation struct {
	Start  Position `json:"start"`            // Start is the location in the source the node starts
	End    Position `json:"end"`              // End is the location in the source the node ends
	Source *string  `json:"source,omitempty"` // Source is optional raw source
}

// Node represents a node in the InfluxDB abstract syntax tree.
type Node interface {
	node()
	Type() string // Type property is a string that contains the variant type of the node
	Location() *SourceLocation
}

func (*BaseNode) node() {}

func (*BlockStatement) node()      {}
func (*ExpressionStatement) node() {}
func (*VariableDeclaration) node() {}
func (*VariableDeclarator) node()  {}
func (*ReturnStatement) node()     {}

func (*CallExpression) node()        {}
func (*MemberExpression) node()      {}
func (*BinaryExpression) node()      {}
func (*UnaryExpression) node()       {}
func (*LogicalExpression) node()     {}
func (*ObjectExpression) node()      {}
func (*ConditionalExpression) node() {}
func (*ArrayExpression) node()       {}

func (*Property) node()   {}
func (*Identifier) node() {}

func (*StringLiteral) node()   {}
func (*BooleanLiteral) node()  {}
func (*NumberLiteral) node()   {}
func (*RegexpLiteral) node()   {}
func (*DurationLiteral) node() {}
func (*DateTimeLiteral) node() {}

// BaseNode holds the attributes every expression or statement should have
type BaseNode struct {
	Loc *SourceLocation `json:"location,omitempty"`
}

// Location is the source location of the Node
func (b *BaseNode) Location() *SourceLocation { return b.Loc }

// Program represents a complete program source tree
type Program struct {
	*BaseNode
	Body []Statement `json:"body"`
}

// Type is the abstract type
func (*Program) Type() string { return "Program" }

func (p *Program) MarshalJSON() ([]byte, error) {
	type Alias Program
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  p.Type(),
		Alias: (*Alias)(p),
	}
	return json.Marshal(raw)
}

func (p *Program) UnmarshalJSON(data []byte) error {
	type Alias Program
	raw := struct {
		*Alias
		Body []json.RawMessage `json:"body"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*p = *(*Program)(raw.Alias)
	}

	p.Body = make([]Statement, len(raw.Body))
	for i, r := range raw.Body {
		s, err := unmarshalStatement(r)
		if err != nil {
			return err
		}
		p.Body[i] = s
	}
	return nil
}

// Statement Perhaps we don't even want statements nor expression statements
type Statement interface {
	Node
	stmt()
}

func (*BlockStatement) stmt()      {}
func (*ExpressionStatement) stmt() {}
func (*ReturnStatement) stmt()     {}
func (*VariableDeclaration) stmt() {}

// BlockStatement is a set of statements
type BlockStatement struct {
	*BaseNode
	Body []Statement `json:"body"`
}

// Type is the abstract type
func (*BlockStatement) Type() string { return "BlockStatement" }

func (s *BlockStatement) MarshalJSON() ([]byte, error) {
	type Alias BlockStatement
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  s.Type(),
		Alias: (*Alias)(s),
	}
	return json.Marshal(raw)
}
func (s *BlockStatement) UnmarshalJSON(data []byte) error {
	type Alias BlockStatement
	raw := struct {
		*Alias
		Body []json.RawMessage `json:"body"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*s = *(*BlockStatement)(raw.Alias)
	}

	s.Body = make([]Statement, len(raw.Body))
	for i, r := range raw.Body {
		stmt, err := unmarshalStatement(r)
		if err != nil {
			return err
		}
		s.Body[i] = stmt
	}
	return nil
}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
type ExpressionStatement struct {
	*BaseNode
	Expression Expression `json:"expression"`
}

// Type is the abstract type
func (*ExpressionStatement) Type() string { return "ExpressionStatement" }

func (s *ExpressionStatement) MarshalJSON() ([]byte, error) {
	type Alias ExpressionStatement
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  s.Type(),
		Alias: (*Alias)(s),
	}
	return json.Marshal(raw)
}
func (s *ExpressionStatement) UnmarshalJSON(data []byte) error {
	type Alias ExpressionStatement
	raw := struct {
		*Alias
		Expression json.RawMessage `json:"expression"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*s = *(*ExpressionStatement)(raw.Alias)
	}

	e, err := unmarshalExpression(raw.Expression)
	if err != nil {
		return err
	}
	s.Expression = e
	return nil
}

// ReturnStatement defines an Expression to return
type ReturnStatement struct {
	*BaseNode
	Argument Expression `json:"argument"`
}

// Type is the abstract type
func (*ReturnStatement) Type() string { return "ReturnStatement" }

func (s *ReturnStatement) MarshalJSON() ([]byte, error) {
	type Alias ReturnStatement
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  s.Type(),
		Alias: (*Alias)(s),
	}
	return json.Marshal(raw)
}
func (s *ReturnStatement) UnmarshalJSON(data []byte) error {
	type Alias ReturnStatement
	raw := struct {
		*Alias
		Argument json.RawMessage `json:"argument"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*s = *(*ReturnStatement)(raw.Alias)
	}

	e, err := unmarshalExpression(raw.Argument)
	if err != nil {
		return err
	}
	s.Argument = e
	return nil
}

// Declaration statements are used to declare the type of one or more local variables.
type Declaration interface {
	Statement
	declaration()
}

func (*VariableDeclaration) declaration() {}

// VariableDeclaration declares one or more variables using assignment
type VariableDeclaration struct {
	*BaseNode
	Declarations []*VariableDeclarator `json:"declarations"`
}

// Type is the abstract type
func (*VariableDeclaration) Type() string { return "VariableDeclaration" }

func (d *VariableDeclaration) MarshalJSON() ([]byte, error) {
	type Alias VariableDeclaration
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  d.Type(),
		Alias: (*Alias)(d),
	}
	return json.Marshal(raw)
}

// VariableDeclarator represents the declaration of a variable
type VariableDeclarator struct {
	*BaseNode
	ID   *Identifier `json:"id"`
	Init Expression  `json:"init"`
}

// Type is the abstract type
func (*VariableDeclarator) Type() string { return "VariableDeclarator" }

func (d *VariableDeclarator) MarshalJSON() ([]byte, error) {
	type Alias VariableDeclarator
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  d.Type(),
		Alias: (*Alias)(d),
	}
	return json.Marshal(raw)
}
func (d *VariableDeclarator) UnmarshalJSON(data []byte) error {
	type Alias VariableDeclarator
	raw := struct {
		*Alias
		Init json.RawMessage `json:"init"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*d = *(*VariableDeclarator)(raw.Alias)
	}

	e, err := unmarshalExpression(raw.Init)
	if err != nil {
		return err
	}
	d.Init = e
	return nil
}

// Expression represents an action that can be performed by InfluxDB that can be evaluated to a value.
type Expression interface {
	Node
	expression()
}

func (*CallExpression) expression()          {}
func (*MemberExpression) expression()        {}
func (*BinaryExpression) expression()        {}
func (*UnaryExpression) expression()         {}
func (*LogicalExpression) expression()       {}
func (*ObjectExpression) expression()        {}
func (*ConditionalExpression) expression()   {}
func (*ArrayExpression) expression()         {}
func (*Identifier) expression()              {}
func (*StringLiteral) expression()           {}
func (*BooleanLiteral) expression()          {}
func (*NumberLiteral) expression()           {}
func (*IntegerLiteral) expression()          {}
func (*RegexpLiteral) expression()           {}
func (*DurationLiteral) expression()         {}
func (*DateTimeLiteral) expression()         {}
func (*ArrowFunctionExpression) expression() {}

// CallExpression represents a function all whose callee may be an Identifier or MemberExpression
type CallExpression struct {
	*BaseNode
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments,omitempty"`
}

// Type is the abstract type
func (*CallExpression) Type() string { return "CallExpression" }

func (e *CallExpression) MarshalJSON() ([]byte, error) {
	type Alias CallExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *CallExpression) UnmarshalJSON(data []byte) error {
	type Alias CallExpression
	raw := struct {
		*Alias
		Callee    json.RawMessage   `json:"callee"`
		Arguments []json.RawMessage `json:"arguments"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*CallExpression)(raw.Alias)
	}

	callee, err := unmarshalExpression(raw.Callee)
	if err != nil {
		return err
	}
	e.Callee = callee

	e.Arguments = make([]Expression, len(raw.Arguments))
	for i, r := range raw.Arguments {
		expr, err := unmarshalExpression(r)
		if err != nil {
			return err
		}
		e.Arguments[i] = expr
	}
	return nil
}

// MemberExpression represents calling a property of a CallExpression
type MemberExpression struct {
	*BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
}

// Type is the abstract type
func (*MemberExpression) Type() string { return "MemberExpression" }

func (e *MemberExpression) MarshalJSON() ([]byte, error) {
	type Alias MemberExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *MemberExpression) UnmarshalJSON(data []byte) error {
	type Alias MemberExpression
	raw := struct {
		*Alias
		Object   json.RawMessage `json:"object"`
		Property json.RawMessage `json:"property"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*MemberExpression)(raw.Alias)
	}

	object, err := unmarshalExpression(raw.Object)
	if err != nil {
		return err
	}
	e.Object = object

	property, err := unmarshalExpression(raw.Property)
	if err != nil {
		return err
	}
	e.Property = property

	return nil
}

type ArrowFunctionExpression struct {
	*BaseNode
	Params []*Identifier `json:"params"`
	Body   Node          `json:"body"`
}

// Type is the abstract type
func (*ArrowFunctionExpression) Type() string { return "ArrowFunctionExpression" }

func (e *ArrowFunctionExpression) MarshalJSON() ([]byte, error) {
	type Alias ArrowFunctionExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *ArrowFunctionExpression) UnmarshalJSON(data []byte) error {
	type Alias ArrowFunctionExpression
	raw := struct {
		*Alias
		Body json.RawMessage `json:"body"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*ArrowFunctionExpression)(raw.Alias)
	}

	body, err := unmarshalNode(raw.Body)
	if err != nil {
		return err
	}
	e.Body = body
	return nil
}

// OperatorKind are Equality and Arithmatic operators.
// Result of evaluating an equality operator is always of type Boolean based on whether the
// comparison is true
// Arithmetic operators take numerical values (either literals or variables) as their operands
//  and return a single numerical value.
type OperatorKind int

const (
	opBegin OperatorKind = iota
	MultiplicationOperator
	DivisionOperator
	AdditionOperator
	SubtractionOperator
	LessThanEqualOperator
	LessThanOperator
	GreaterThanEqualOperator
	GreaterThanOperator
	StartsWithOperator
	InOperator
	NotOperator
	NotEmptyOperator
	EmptyOperator
	EqualOperator
	NotEqualOperator
	opEnd
)

func (o OperatorKind) String() string {
	return OperatorTokens[o]
}

// OperatorLookup converts the operators to OperatorKind
func OperatorLookup(op string) OperatorKind {
	return operators[op]
}

func (o OperatorKind) MarshalText() ([]byte, error) {
	text, ok := OperatorTokens[o]
	if !ok {
		return nil, fmt.Errorf("unknown operator %d", int(o))
	}
	return []byte(text), nil
}
func (o *OperatorKind) UnmarshalText(data []byte) error {
	var ok bool
	*o, ok = operators[string(data)]
	if !ok {
		return fmt.Errorf("unknown operator %q", string(data))
	}
	return nil
}

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmatic operators
type BinaryExpression struct {
	*BaseNode
	Operator OperatorKind `json:"operator"`
	Left     Expression   `json:"left"`
	Right    Expression   `json:"right"`
}

// Type is the abstract type
func (*BinaryExpression) Type() string { return "BinaryExpression" }

func (e *BinaryExpression) MarshalJSON() ([]byte, error) {
	type Alias BinaryExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *BinaryExpression) UnmarshalJSON(data []byte) error {
	type Alias BinaryExpression
	raw := struct {
		*Alias
		Left  json.RawMessage `json:"left"`
		Right json.RawMessage `json:"right"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*BinaryExpression)(raw.Alias)
	}

	l, err := unmarshalExpression(raw.Left)
	if err != nil {
		return err
	}
	e.Left = l

	r, err := unmarshalExpression(raw.Right)
	if err != nil {
		return err
	}
	e.Right = r
	return nil
}

// UnaryExpression use operators act on a single operand in an expression.
type UnaryExpression struct {
	*BaseNode
	Operator OperatorKind `json:"operator"`
	Argument Expression   `json:"argument"`
}

// Type is the abstract type
func (*UnaryExpression) Type() string { return "UnaryExpression" }

func (e *UnaryExpression) MarshalJSON() ([]byte, error) {
	type Alias UnaryExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *UnaryExpression) UnmarshalJSON(data []byte) error {
	type Alias UnaryExpression
	raw := struct {
		*Alias
		Argument json.RawMessage `json:"argument"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*UnaryExpression)(raw.Alias)
	}

	argument, err := unmarshalExpression(raw.Argument)
	if err != nil {
		return err
	}
	e.Argument = argument

	return nil
}

// LogicalOperatorKind are used with boolean (logical) values
type LogicalOperatorKind int

const (
	logOpBegin LogicalOperatorKind = iota
	AndOperator
	OrOperator
	logOpEnd
)

func (o LogicalOperatorKind) String() string {
	return LogicalOperatorTokens[o]
}

// LogicalOperatorLookup converts the operators to LogicalOperatorKind
func LogicalOperatorLookup(op string) LogicalOperatorKind {
	return logOperators[op]
}

func (o LogicalOperatorKind) MarshalText() ([]byte, error) {
	text, ok := LogicalOperatorTokens[o]
	if !ok {
		return nil, fmt.Errorf("unknown logical operator %d", int(o))
	}
	return []byte(text), nil
}
func (o *LogicalOperatorKind) UnmarshalText(data []byte) error {
	var ok bool
	*o, ok = logOperators[string(data)]
	if !ok {
		return fmt.Errorf("unknown logical operator %q", string(data))
	}
	return nil
}

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
type LogicalExpression struct {
	*BaseNode
	Operator LogicalOperatorKind `json:"operator"`
	Left     Expression          `json:"left"`
	Right    Expression          `json:"right"`
}

// Type is the abstract type
func (*LogicalExpression) Type() string { return "LogicalExpression" }

func (e *LogicalExpression) MarshalJSON() ([]byte, error) {
	type Alias LogicalExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *LogicalExpression) UnmarshalJSON(data []byte) error {
	type Alias LogicalExpression
	raw := struct {
		*Alias
		Left  json.RawMessage `json:"left"`
		Right json.RawMessage `json:"right"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*LogicalExpression)(raw.Alias)
	}

	l, err := unmarshalExpression(raw.Left)
	if err != nil {
		return err
	}
	e.Left = l

	r, err := unmarshalExpression(raw.Right)
	if err != nil {
		return err
	}
	e.Right = r
	return nil
}

// ArrayExpression is used to create and directly specify the elements of an array object
type ArrayExpression struct {
	*BaseNode
	Elements []Expression `json:"elements"`
}

// Type is the abstract type
func (*ArrayExpression) Type() string { return "ArrayExpression" }

func (e *ArrayExpression) MarshalJSON() ([]byte, error) {
	type Alias ArrayExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}
func (e *ArrayExpression) UnmarshalJSON(data []byte) error {
	type Alias ArrayExpression
	raw := struct {
		*Alias
		Elements []json.RawMessage `json:"elements"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*ArrayExpression)(raw.Alias)
	}

	e.Elements = make([]Expression, len(raw.Elements))
	for i, r := range raw.Elements {
		expr, err := unmarshalExpression(r)
		if err != nil {
			return err
		}
		e.Elements[i] = expr
	}
	return nil
}

// ObjectExpression allows the declaration of an anonymous object within a declaration.
type ObjectExpression struct {
	*BaseNode
	Properties []*Property `json:"properties"`
}

// Type is the abstract type
func (*ObjectExpression) Type() string { return "ObjectExpression" }

func (e *ObjectExpression) MarshalJSON() ([]byte, error) {
	type Alias ObjectExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
type ConditionalExpression struct {
	*BaseNode
	Test       Expression `json:"test"`
	Alternate  Expression `json:"alternate"`
	Consequent Expression `json:"consequent"`
}

// Type is the abstract type
func (*ConditionalExpression) Type() string { return "ConditionalExpression" }

func (e *ConditionalExpression) MarshalJSON() ([]byte, error) {
	type Alias ConditionalExpression
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type(),
		Alias: (*Alias)(e),
	}
	return json.Marshal(raw)
}

func (e *ConditionalExpression) UnmarshalJSON(data []byte) error {
	type Alias ConditionalExpression
	raw := struct {
		*Alias
		Test       json.RawMessage `json:"test"`
		Alternate  json.RawMessage `json:"alternate"`
		Consequent json.RawMessage `json:"consequent"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*e = *(*ConditionalExpression)(raw.Alias)
	}

	test, err := unmarshalExpression(raw.Test)
	if err != nil {
		return err
	}
	e.Test = test

	alternate, err := unmarshalExpression(raw.Alternate)
	if err != nil {
		return err
	}
	e.Alternate = alternate

	consequent, err := unmarshalExpression(raw.Consequent)
	if err != nil {
		return err
	}
	e.Consequent = consequent
	return nil
}

// Property is the value associated with a key
type Property struct {
	*BaseNode
	Key   *Identifier `json:"key"`
	Value Expression  `json:"value"`
}

func (p *Property) MarshalJSON() ([]byte, error) {
	type Alias Property
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  p.Type(),
		Alias: (*Alias)(p),
	}
	return json.Marshal(raw)
}

func (p *Property) UnmarshalJSON(data []byte) error {
	type Alias Property
	raw := struct {
		*Alias
		Value json.RawMessage `json:"value"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Alias != nil {
		*p = *(*Property)(raw.Alias)
	}

	value, err := unmarshalExpression(raw.Value)
	if err != nil {
		return err
	}
	p.Value = value
	return nil
}

// Type is the abstract type
func (*Property) Type() string { return "Property" }

// Identifier represents a name that identifies a unique Node
type Identifier struct {
	*BaseNode
	Name string `json:"name"`
}

// Type is the abstract type
func (*Identifier) Type() string { return "Identifier" }

func (i *Identifier) MarshalJSON() ([]byte, error) {
	type Alias Identifier
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  i.Type(),
		Alias: (*Alias)(i),
	}
	return json.Marshal(raw)
}

// Literal are thelexical forms for literal expressions which define
// boolean, string, integer, number, duration, datetime and field values.
// Literals must be coerced explicitly.
type Literal interface {
	Expression
	literal()
}

func (*StringLiteral) literal()   {}
func (*BooleanLiteral) literal()  {}
func (*NumberLiteral) literal()   {}
func (*IntegerLiteral) literal()  {}
func (*RegexpLiteral) literal()   {}
func (*DurationLiteral) literal() {}
func (*DateTimeLiteral) literal() {}

// StringLiteral expressions begin and end with double quote marks.
type StringLiteral struct {
	*BaseNode
	Value string `json:"value"`
}

func (*StringLiteral) Type() string { return "StringLiteral" }

func (l *StringLiteral) MarshalJSON() ([]byte, error) {
	type Alias StringLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
	}
	return json.Marshal(raw)
}

// BooleanLiteral represent boolean values
type BooleanLiteral struct {
	*BaseNode
	Value bool `json:"value"`
}

// Type is the abstract type
func (*BooleanLiteral) Type() string { return "BooleanLiteral" }

func (l *BooleanLiteral) MarshalJSON() ([]byte, error) {
	type Alias BooleanLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
	}
	return json.Marshal(raw)
}

// NumberLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
type NumberLiteral struct {
	*BaseNode
	Value float64 `json:"value"`
}

// Type is the abstract type
func (*NumberLiteral) Type() string { return "NumberLiteral" }

func (l *NumberLiteral) MarshalJSON() ([]byte, error) {
	type Alias NumberLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
	}
	return json.Marshal(raw)
}

// IntegerLiteral represent integer numbers.
type IntegerLiteral struct {
	*BaseNode
	Value int64 `json:"value"`
}

// Type is the abstract type
func (*IntegerLiteral) Type() string { return "IntegerLiteral" }

func (l *IntegerLiteral) MarshalJSON() ([]byte, error) {
	type Alias IntegerLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
		Value string `json:"value"`
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
		Value: strconv.FormatInt(l.Value, 10),
	}
	return json.Marshal(raw)
}
func (l *IntegerLiteral) UnmarshalJSON(data []byte) error {
	type Alias IntegerLiteral
	raw := struct {
		*Alias
		Value string `json:"value"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Alias != nil {
		*l = *(*IntegerLiteral)(raw.Alias)
	}

	value, err := strconv.ParseInt(raw.Value, 10, 64)
	if err != nil {
		return err
	}
	l.Value = value
	return nil
}

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
type RegexpLiteral struct {
	*BaseNode
	Value *regexp.Regexp `json:"value"`
}

// Type is the abstract type
func (*RegexpLiteral) Type() string { return "RegexpLiteral" }

func (l *RegexpLiteral) MarshalJSON() ([]byte, error) {
	type Alias RegexpLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
		Value string `json:"value"`
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
		Value: l.Value.String(),
	}
	return json.Marshal(raw)
}
func (l *RegexpLiteral) UnmarshalJSON(data []byte) error {
	type Alias RegexpLiteral
	raw := struct {
		*Alias
		Value string `json:"value"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Alias != nil {
		*l = *(*RegexpLiteral)(raw.Alias)
	}

	value, err := regexp.Compile(raw.Value)
	if err != nil {
		return err
	}
	l.Value = value
	return nil
}

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
type DurationLiteral struct {
	*BaseNode
	Value time.Duration `json:"value"`
}

// Type is the abstract type
func (*DurationLiteral) Type() string { return "DurationLiteral" }

func (l *DurationLiteral) MarshalJSON() ([]byte, error) {
	type Alias DurationLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
		Value string `json:"value"`
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
		Value: l.Value.String(),
	}
	return json.Marshal(raw)
}
func (l *DurationLiteral) UnmarshalJSON(data []byte) error {
	type Alias DurationLiteral
	raw := struct {
		*Alias
		Value string `json:"value"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Alias != nil {
		*l = *(*DurationLiteral)(raw.Alias)
	}

	value, err := time.ParseDuration(raw.Value)
	if err != nil {
		return err
	}
	l.Value = value
	return nil
}

// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
type DateTimeLiteral struct {
	*BaseNode
	Value time.Time `json:"value"`
}

// Type is the abstract type
func (*DateTimeLiteral) Type() string { return "DateTimeLiteral" }

func (l *DateTimeLiteral) MarshalJSON() ([]byte, error) {
	type Alias DateTimeLiteral
	raw := struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  l.Type(),
		Alias: (*Alias)(l),
	}
	return json.Marshal(raw)
}

// OperatorTokens converts OperatorKind to string
var OperatorTokens = map[OperatorKind]string{
	MultiplicationOperator:   "*",
	DivisionOperator:         "/",
	AdditionOperator:         "+",
	SubtractionOperator:      "-",
	LessThanEqualOperator:    "<=",
	LessThanOperator:         "<",
	GreaterThanOperator:      ">",
	GreaterThanEqualOperator: ">=",
	InOperator:               "in",
	NotOperator:              "not",
	NotEmptyOperator:         "not empty",
	EmptyOperator:            "empty",
	StartsWithOperator:       "startswith",
	EqualOperator:            "==",
	NotEqualOperator:         "!=",
}

// LogicalOperatorTokens converts LogicalOperatorKind to string
var LogicalOperatorTokens = map[LogicalOperatorKind]string{
	AndOperator: "and",
	OrOperator:  "or",
}

var operators map[string]OperatorKind
var logOperators map[string]LogicalOperatorKind

func init() {
	operators = make(map[string]OperatorKind)
	for op := opBegin + 1; op < opEnd; op++ {
		operators[OperatorTokens[op]] = op
	}

	logOperators = make(map[string]LogicalOperatorKind)
	for op := logOpBegin + 1; op < logOpEnd; op++ {
		logOperators[LogicalOperatorTokens[op]] = op
	}
}

func unmarshalStatement(msg json.RawMessage) (Statement, error) {
	if len(msg) == 0 {
		return nil, nil
	}
	n, err := unmarshalNode(msg)
	if err != nil {
		return nil, err
	}
	s, ok := n.(Statement)
	if !ok {
		return nil, fmt.Errorf("node %q is not a statement", n.Type())
	}
	return s, nil
}
func unmarshalExpression(msg json.RawMessage) (Expression, error) {
	if len(msg) == 0 {
		return nil, nil
	}
	n, err := unmarshalNode(msg)
	if err != nil {
		return nil, err
	}
	e, ok := n.(Expression)
	if !ok {
		return nil, fmt.Errorf("node %q is not an expression", n.Type())
	}
	return e, nil
}
func unmarshalNode(msg json.RawMessage) (Node, error) {
	if len(msg) == 0 {
		return nil, nil
	}

	type typeRawMessage struct {
		Type string `json:"type"`
	}

	typ := typeRawMessage{}
	if err := json.Unmarshal(msg, &typ); err != nil {
		return nil, err
	}

	var node Node
	switch typ.Type {
	case "Program":
		node = new(Program)
	case "BlockStatement":
		node = new(BlockStatement)
	case "ExpressionStatement":
		node = new(ExpressionStatement)
	case "ReturnStatement":
		node = new(ReturnStatement)
	case "VariableDeclaration":
		node = new(VariableDeclaration)
	case "VariableDeclarator":
		node = new(VariableDeclarator)
	case "CallExpression":
		node = new(CallExpression)
	case "MemberExpression":
		node = new(MemberExpression)
	case "BinaryExpression":
		node = new(BinaryExpression)
	case "UnaryExpression":
		node = new(UnaryExpression)
	case "LogicalExpression":
		node = new(LogicalExpression)
	case "ObjectExpression":
		node = new(ObjectExpression)
	case "ConditionalExpression":
		node = new(ConditionalExpression)
	case "ArrayExpression":
		node = new(ArrayExpression)
	case "Identifier":
		node = new(Identifier)
	case "StringLiteral":
		node = new(StringLiteral)
	case "BooleanLiteral":
		node = new(BooleanLiteral)
	case "NumberLiteral":
		node = new(NumberLiteral)
	case "IntegerLiteral":
		node = new(IntegerLiteral)
	case "RegexpLiteral":
		node = new(RegexpLiteral)
	case "DurationLiteral":
		node = new(DurationLiteral)
	case "DateTimeLiteral":
		node = new(DateTimeLiteral)
	case "ArrowFunctionExpression":
		node = new(ArrowFunctionExpression)
	case "Property":
		node = new(Property)
	default:
		return nil, fmt.Errorf("unknown type %q", typ.Type)
	}

	if err := json.Unmarshal(msg, node); err != nil {
		return nil, err
	}
	return node, nil
}
func UnmarshalNode(data []byte) (Node, error) {
	return unmarshalNode((json.RawMessage)(data))
}
