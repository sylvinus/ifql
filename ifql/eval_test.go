package ifql_test

import (
	"testing"
)

func TestFortyTwo(t *testing.T) {
	program := `
var six = bar()
var nine = baz()

var answer = foo() == six * nine
`

	t.Fatal(program)
}
