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
	Type() ArgType
	Value() interface{}
}
