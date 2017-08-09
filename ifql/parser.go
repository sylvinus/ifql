package parser

type Function struct {
	Name string
	Args
}

type FunctionChain struct {
	Name string
	Args
}

type FunctionArg struct {
	Name string
	Arg  Arg
}

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
	NumKinds int = iota
)
