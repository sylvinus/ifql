package main

import (
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
	"github.com/influxdata/ifql/semantic"
)

var scope = interpreter.NewScope()

func completer(d prompt.Document) []prompt.Suggest {
	names := scope.Names()
	s := make([]prompt.Suggest, len(names))
	for i, n := range scope.Names() {
		s[i] = prompt.Suggest{
			Text: n,
		}
	}
	return s
}

func input(t string) {
	if t == "" {
		return
	}
	astProg, err := parser.NewAST(t)
	if err != nil {
		fmt.Println(err)
		return
	}

	semProg, err := semantic.New(astProg)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := interpreter.Eval(semProg, scope, nil); err != nil {
		fmt.Println(err)
		return
	}

	v := scope.Return()
	if v.Type() != semantic.KInvalid {
		fmt.Println(v.Value())
	}
}

func main() {
	scope = interpreter.NewScope()
	p := prompt.New(
		input,
		completer,
		prompt.OptionPrefix("> "),
		prompt.OptionTitle("ifql"),
	)
	p.Run()
}
