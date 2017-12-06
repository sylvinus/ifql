package ast

import (
	"regexp"
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

func (*ExpressionStatement) node() {}
func (*VariableDeclaration) node() {}
func (*VariableDeclarator) node()  {}

func (*CallExpression) node()        {}
func (*MemberExpression) node()      {}
func (*SequenceExpression) node()    {}
func (*BinaryExpression) node()      {}
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

// Statement Perhaps we don't even want statements nor expression statements
type Statement interface {
	Node
	stmt()
}

func (*ExpressionStatement) stmt() {}
func (*VariableDeclaration) stmt() {}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
type ExpressionStatement struct {
	*BaseNode
	Expression Expression `json:"expression"`
}

// Type is the abstract type
func (*ExpressionStatement) Type() string { return "ExpressionStatement" }

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

// VariableDeclarator represents the declaration of a variable
type VariableDeclarator struct {
	*BaseNode
	ID   *Identifier `json:"id"`
	Init Expression  `json:"init"`
}

// Type is the abstract type
func (*VariableDeclarator) Type() string { return "VariableDeclarator" }

// Expression represents an action that can be performed by InfluxDB that can be evaluated to a value.
type Expression interface {
	Node
	expression()
}

func (*CallExpression) expression()          {}
func (*MemberExpression) expression()        {}
func (*SequenceExpression) expression()      {}
func (*BinaryExpression) expression()        {}
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

// MemberExpression represents calling a property of a CallExpression
type MemberExpression struct {
	*BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
}

// Type is the abstract type
func (*MemberExpression) Type() string { return "MemberExpression" }

// SequenceExpression uses comma operator to include multiple expressions
// in a location that requires a single expression.  Typically, multiple
// select statements on one line.
type SequenceExpression struct {
	*BaseNode
	Expressions []Expression `json:"expressions"`
}

// Type is the abstract type
func (*SequenceExpression) Type() string { return "SequenceExpression" }

type ArrowFunctionExpression struct {
	*BaseNode
	Params []*Identifier `json:"params"`
	Body   Expression    `json:"body"`
}

// Type is the abstract type
func (*ArrowFunctionExpression) Type() string { return "ArrowFunctionExpression" }

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

// ArrayExpression is used to create and directly specify the elements of an array object
type ArrayExpression struct {
	*BaseNode
	Elements []Expression `json:"elements"`
}

// Type is the abstract type
func (*ArrayExpression) Type() string { return "ArrayExpression" }

// ObjectExpression allows the declaration of an anonymous object within a declaration.
type ObjectExpression struct {
	*BaseNode
	Properties []*Property `json:"properties"`
}

// Type is the abstract type
func (*ObjectExpression) Type() string { return "ObjectExpression" }

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

// Property is the value associated with a key
type Property struct {
	*BaseNode
	Key   *Identifier `json:"key"`
	Value Expression  `json:"value"`
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

// BooleanLiteral represent boolean values
type BooleanLiteral struct {
	*BaseNode
	Value bool `json:"value"`
}

// Type is the abstract type
func (*BooleanLiteral) Type() string { return "BooleanLiteral" }

// NumberLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
type NumberLiteral struct {
	*BaseNode
	Value float64 `json:"value"`
}

// Type is the abstract type
func (*NumberLiteral) Type() string { return "NumberLiteral" }

// IntegerLiteral represent integer numbers.
type IntegerLiteral struct {
	*BaseNode
	Value int64 `json:"value"`
}

// Type is the abstract type
func (*IntegerLiteral) Type() string { return "IntegerLiteral" }

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
type RegexpLiteral struct {
	*BaseNode
	Value *regexp.Regexp `json:"value"`
}

// Type is the abstract type
func (*RegexpLiteral) Type() string { return "RegexpLiteral" }

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
type DurationLiteral struct {
	*BaseNode
	Value time.Duration `json:"value"`
}

// Type is the abstract type
func (*DurationLiteral) Type() string { return "DurationLiteral" }

// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
type DateTimeLiteral struct {
	*BaseNode
	Value time.Time `json:"value"`
}

// Type is the abstract type
func (*DateTimeLiteral) Type() string { return "DateTimeLiteral" }

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
