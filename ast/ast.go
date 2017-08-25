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

func (*BaseNode) node()              {}
func (*ExpressionStatement) node()   {}
func (*VariableDeclaration) node()   {}
func (*VariableDeclarator) node()    {}
func (*FunctionExpression) node()    {}
func (*SequenceExpression) node()    {}
func (*BinaryExpression) node()      {}
func (*LogicalExpression) node()     {}
func (*ConditionalExpression) node() {}
func (*ArrayExpression) node()       {}
func (*Property) node()              {}
func (*Identifier) node()            {}
func (*Literal) node()               {}

// BaseNode holds the attributes every expression or statement should have
type BaseNode struct {
	loc *SourceLocation
}

// Location is the source location of the Node
func (b *BaseNode) Location() *SourceLocation { return b.loc }

// Program represents a complete program source tree
type Program struct {
	*BaseNode
	Body []Statement
}

func (*Program) Type() string { return "Program" }

// Statement Perhaps we don't even want statements nor expression statements
type Statement interface {
	Node
	stmt()
}

func (*ExpressionStatement) stmt() {}
func (*VariableDeclaration) stmt() {}

type ExpressionStatement struct {
	Expression Expression
}

func (*ExpressionStatement) Type() string { return "ExpressionStatement" }

type Declaration interface {
	Statement
	declaration()
}

func (*VariableDeclaration) declaration() {}

type VariableDeclaration struct {
	*BaseNode
	Declarations []VariableDeclarator
}

func (*VariableDeclaration) Type() string { return "VariableDeclaration" }

type VariableDeclarator struct {
	*BaseNode
	ID   Identifier
	Init Expression
}

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

type FunctionExpression struct {
	*BaseNode
	ID     Identifier
	Params []Property
	Loc    *SourceLocation
	Chains []*FunctionExpression
}

func (*FunctionExpression) Type() string { return "FunctionExpression" }

type SequenceExpression struct {
	*BaseNode
	Expressions []Expression
}

func (*SequenceExpression) Type() string { return "SequenceExpression" }

type OperatorKind int

// TODO: fill out operatorkind
type BinaryExpression struct {
	*BaseNode
	Operator OperatorKind
	Left     Expression
	Right    Expression
}

func (*BinaryExpression) Type() string { return "BinaryExpression" }

type LogicalOperatorKind int

// TODO Define logicaloperator kind
type LogicalExpression struct {
	*BaseNode
	Operator LogicalOperatorKind
	Left     Expression
	Right    Expression
}

func (*LogicalExpression) Type() string { return "LogicalExpression" }

type ArrayExpression struct {
	*BaseNode
	Elements []Expression
}

func (*ArrayExpression) Type() string { return "ArrayExpression" }

type ConditionalExpression struct {
	*BaseNode
	Test       Expression
	Alternate  Expression
	Consequent Expression
}

func (*ConditionalExpression) Type() string { return "ConditionalExpression" }

type Property struct {
	*BaseNode
	Key   interface{} // Literal or Identifier
	Value Expression
}

func (*Property) Type() string { return "Property" }

// Identifier represents a name that identifies a unique Node
type Identifier struct {
	*BaseNode
	Name string
}

// Type of an identifier
func (*Identifier) Type() string { return "Identifier" }

// Location is the optional location of the Identifier

// TODO: Should this be an interface?
type Literal struct {
	*BaseNode
}

func (*Literal) Type() string { return "Literal" }

type StringLiteral struct {
	*Literal
	Value string
}

type BooleanLiteral struct {
	*Literal
	Value bool
}

type NumberLiteral struct {
	*Literal
	Value float64
}

type RegExpLiteral struct {
	*Literal
	Value regexp.Regexp
}

type DurationLiteral struct {
	*Literal
	Value time.Duration
}

type DateTimeLiteral struct {
	*Literal
	Value time.Time
}

type FieldLiteral struct {
	*Literal
	Value string
}

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
