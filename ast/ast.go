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

func (*ExpressionStatement) node() {}
func (*VariableDeclaration) node() {}
func (*VariableDeclarator) node()  {}
func (*FunctionExpression) node()  {}
func (*SequenceExpression) node()  {}
func (*BinaryExpression) node()    {}
func (*LogicalExpression) node()   {}
func (*ArrayExpression) node()     {}
func (*Property) node()            {}
func (*Identifier) node()          {}
func (*Literal) node()             {}

// Program represents a complete program source tree
type Program struct {
	Body []Statement
	Loc  *SourceLocation
}

func (*Program) Type() string                { return "Program" }
func (p *Program) Location() *SourceLocation { return p.Loc }

// Perhaps we don't even want statements nor expression statements
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
	Declarations []VariableDeclarator
	Loc          *SourceLocation
}

func (*VariableDeclaration) Type() string                { return "VariableDeclaration" }
func (v *VariableDeclaration) Location() *SourceLocation { return v.Loc }

type VariableDeclarator struct {
	ID   Identifier
	Init Expression
	Loc  *SourceLocation
}

func (*VariableDeclarator) Type() string                { return "VariableDeclarator" }
func (v *VariableDeclarator) Location() *SourceLocation { return v.Loc }

// Expression represents an action that can be performed by InfluxDB that can be evaluated to a value.
type Expression interface {
	Node
	expression()
}

func (*FunctionExpression) expression() {}
func (*SequenceExpression) expression() {}
func (*BinaryExpression) expression()   {}
func (*LogicalExpression) expression()  {}
func (*ArrayExpression) expression()    {}

type FunctionExpression struct {
	ID     Identifier
	Params []Property
	Loc    *SourceLocation
	Chains []*FunctionExpression
}

func (*FunctionExpression) Type() string                { return "FunctionExpression" }
func (f *FunctionExpression) Location() *SourceLocation { return f.Loc }

type SequenceExpression struct {
	Expressions []Expression
	Loc         *SourceLocation
}

func (*SequenceExpression) Type() string                { return "SequenceExpression" }
func (s *SequenceExpression) Location() *SourceLocation { return s.Loc }

type OperatorKind int

// TODO: fill out operatorkind
type BinaryExpression struct {
	Operator OperatorKind
	Left     Expression
	Right    Expression
	Loc      *SourceLocation
}

func (*BinaryExpression) Type() string                { return "BinaryExpression" }
func (b *BinaryExpression) Location() *SourceLocation { return b.Loc }

type LogicalOperatorKind int

// TODO Define logicaloperator kind
type LogicalExpression struct {
	Operator LogicalOperatorKind
	Left     Expression
	Right    Expression
	Loc      *SourceLocation
}

func (*LogicalExpression) Type() string                { return "LogicalExpression" }
func (l *LogicalExpression) Location() *SourceLocation { return l.Loc }

type ArrayExpression struct {
	Elements []Expression
	Loc      *SourceLocation
}

func (*ArrayExpression) Type() string                { return "ArrayExpression" }
func (a *ArrayExpression) Location() *SourceLocation { return a.Loc }

type Property struct {
	Key   interface{} // Literal or Identifier
	Value Expression
	Loc   *SourceLocation
}

func (*Property) Type() string                { return "Property" }
func (p *Property) Location() *SourceLocation { return p.Loc }

// Identifier represents a name that identifies a unique Node
type Identifier struct {
	Name string
	Loc  *SourceLocation
}

// Type of an identifier
func (*Identifier) Type() string { return "Identifier" }

// Location is the optional location of the Identifier
func (i *Identifier) Location() *SourceLocation { return i.Loc }

// TODO: Should this be an interface?
type Literal struct {
	Loc *SourceLocation
}

func (*Literal) Type() string                { return "Literal" }
func (l *Literal) Location() *SourceLocation { return l.Loc }

type StringLiteral struct {
	Literal
	Value string
}

type BooleanLiteral struct {
	Literal
	Value bool
}

type NumberLiteral struct {
	Literal
	Value float64
}

type RegExpLiteral struct {
	Literal
	Value regexp.Regexp
}

type DurationLiteral struct {
	Literal
	Value time.Duration
}

type DateTimeLiteral struct {
	Literal
	Value time.Time
}

type FieldLiteral struct {
	Literal
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
