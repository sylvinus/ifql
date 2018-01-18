package main

import (
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
	"github.com/influxdata/ifql/semantic"
)

func completer(d prompt.Document) []prompt.Suggest {
	return nil
}

func main() {
	fmt.Println("Please select table.")
	for {
		t := prompt.Input("> ", completer)
		if t == "exit" {
			break
		}
		astProg, err := parser.NewAST(t)
		if err != nil {
			fmt.Println(err)
			continue
		}

		semProg, err := semantic.New(astProg)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if err := interpreter.Eval(semProg, interpreter.NewScope(), nil); err != nil {
			fmt.Println(err)
			continue
		}

	}
}
