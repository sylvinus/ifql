package ifql

import (
	"github.com/influxdata/ifql/ast"
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

// NewAST parses ifql query and produces an ast.Program
func NewAST(ifql string, opts ...Option) (*ast.Program, error) {
	f, err := Parse("", []byte(ifql), opts...)
	if err != nil {
		return nil, err
	}
	return f.(*ast.Program), nil
}
