package ast

import (
	"regexp"
	"time"
)

// Position represents a specific location in the source
type Position struct {
	Line   int // Line is the line in the source marked by this position
	Column int // Column is the column in the source marked by this position
}

// SourceLocation represents the location of a node in the AST
type SourceLocation struct {
	Start  Position // Start is the location in the source the node starts
	End    Position // End is the location in the source the node ends
	Source *string  // Source is optional raw source
}

// Node represents a node in the InfluxDB abstract syntax tree.
type Node interface {
	// node is unexported to ensure implementations of Node
	// can only originate in this package.
	node()
	Type() string // Type property is a string that contains the variant type of the node
	Location() *SourceLocation
}

func (*BaseNode) node()            {}
func (*ExpressionStatement) node() {}
func (*VariableDeclaration) node() {}
func (*VariableDeclarator) node()  {}

func (*FunctionExpression) node()    {}
func (*SequenceExpression) node()    {}
func (*BinaryExpression) node()      {}
func (*LogicalExpression) node()     {}
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
func (*FieldLiteral) node()    {}

// BaseNode holds the attributes every expression or statement should have
type BaseNode struct {
	Loc *SourceLocation
}

// Location is the source location of the Node
func (b *BaseNode) Location() *SourceLocation { return b.Loc }

// Program represents a complete program source tree
type Program struct {
	*BaseNode
	Body []Statement
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
	Expression Expression
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
	Declarations []VariableDeclarator
}

// Type is the abstract type
func (*VariableDeclaration) Type() string { return "VariableDeclaration" }

// VariableDeclarator represents the declaration of a variable
type VariableDeclarator struct {
	*BaseNode
	ID   Identifier
	Init Expression
}

// Type is the abstract type
func (*VariableDeclarator) Type() string { return "VariableDeclarator" }

// Expression represents an action that can be performed by InfluxDB that can be evaluated to a value.
type Expression interface {
	Node
	expression()
}

func (*FunctionExpression) expression()    {}
func (*SequenceExpression) expression()    {}
func (*BinaryExpression) expression()      {}
func (*LogicalExpression) expression()     {}
func (*ConditionalExpression) expression() {}
func (*ArrayExpression) expression()       {}

func (*StringLiteral) expression()   {}
func (*BooleanLiteral) expression()  {}
func (*NumberLiteral) expression()   {}
func (*RegexpLiteral) expression()   {}
func (*DurationLiteral) expression() {}
func (*DateTimeLiteral) expression() {}
func (*FieldLiteral) expression()    {}

// FunctionExpression represents a function call with an identifier and properties.
type FunctionExpression struct {
	*BaseNode
	ID     Identifier
	Params []Property
	Loc    *SourceLocation
	Chains []*FunctionExpression
}

// Type is the abstract type
func (*FunctionExpression) Type() string { return "FunctionExpression" }

// SequenceExpression uses comma operator to include multiple expressions
// in a location that requires a single expression.  Typically, multiple
// select statements on one line.
type SequenceExpression struct {
	*BaseNode
	Expressions []Expression
}

// Type is the abstract type
func (*SequenceExpression) Type() string { return "SequenceExpression" }

type OperatorKind int

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmatic operators
type BinaryExpression struct {
	*BaseNode
	Operator OperatorKind
	Left     Expression
	Right    Expression
}

// Type is the abstract type
func (*BinaryExpression) Type() string { return "BinaryExpression" }

// TODO Define logicaloperator kind
type LogicalOperatorKind int

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
type LogicalExpression struct {
	*BaseNode
	Operator LogicalOperatorKind
	Left     Expression
	Right    Expression
}

// Type is the abstract type
func (*LogicalExpression) Type() string { return "LogicalExpression" }

// ArrayExpression is used to create and directly specify the elements of an array object
type ArrayExpression struct {
	*BaseNode
	Elements []Expression
}

// Type is the abstract type
func (*ArrayExpression) Type() string { return "ArrayExpression" }

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
type ConditionalExpression struct {
	*BaseNode
	Test       Expression
	Alternate  Expression
	Consequent Expression
}

// Type is the abstract type
func (*ConditionalExpression) Type() string { return "ConditionalExpression" }

// Property is the value associated with a key
type Property struct {
	*BaseNode
	Key   interface{} // Literal or Identifier
	Value Expression
}

// Type is the abstract type
func (*Property) Type() string { return "Property" }

// Identifier represents a name that identifies a unique Node
type Identifier struct {
	*BaseNode
	Name string
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
func (*RegexpLiteral) literal()   {}
func (*DurationLiteral) literal() {}
func (*DateTimeLiteral) literal() {}
func (*FieldLiteral) literal()    {}

// StringLiteral expressions begin and end with double quote marks.
type StringLiteral struct {
	*BaseNode
	Value string
}

// BooleanLiteral represent boolean values
type BooleanLiteral struct {
	*BaseNode
	Value bool
}

// Type is the abstract type
func (*BooleanLiteral) Type() string { return "BooleanLiteral" }

// NumberLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
type NumberLiteral struct {
	*BaseNode
	Value float64
}

// Type is the abstract type
func (*NumberLiteral) Type() string { return "NumberLiteral" }

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
type RegexpLiteral struct {
	*BaseNode
	Value regexp.Regexp
}

// Type is the abstract type
func (*RegexpLiteral) Type() string { return "RegexpLiteral" }

// DurationLiteral represents the elapsed time between two instants as an int64 nanosecond count with syntax of golang's time.Duration
type DurationLiteral struct {
	*BaseNode
	Value time.Duration
}

// Type is the abstract type
func (*DurationLiteral) Type() string { return "DurationLiteral" }

// DateTimeLiteral represents an instant in time with nanosecond precision using the syntax of golang's RFC3339 Nanosecond variant
type DateTimeLiteral struct {
	*BaseNode
	Value time.Time
}

// Type is the abstract type
func (*DateTimeLiteral) Type() string { return "DateTimeLiteral" }

// FieldLiteral represents the point at a time and tagset with syntax `$`
type FieldLiteral struct {
	*BaseNode
	Value string
}

// Type is the abstract type
func (*FieldLiteral) Type() string { return "FieldLiteral" }

/*
enum BinaryOperator {
    "==" | "!=" | "===" | "!=="
         | "<" | "<=" | ">" | ">="
         | "<<" | ">>" | ">>>"
         | "+" | "-" | "*" | "/" | "%"
         | "|" | "^" | "&" | "in"
}

enum LogicalOperator {
    "||" | "&&"
}*/
