package execute_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query/execute"
)

func TestCompilationCache(t *testing.T) {
	add := &ast.ArrowFunctionExpression{
		Params: []*ast.Property{
			{Key: &ast.Identifier{Name: "a"}},
			{Key: &ast.Identifier{Name: "b"}},
		},
		Body: &ast.BinaryExpression{
			Operator: ast.AdditionOperator,
			Left:     &ast.Identifier{Name: "a"},
			Right:    &ast.Identifier{Name: "b"},
		},
	}
	testCases := []struct {
		name  string
		types map[execute.ReferencePath]execute.DataType
		scope map[execute.ReferencePath]execute.Value
		want  execute.Value
	}{
		{
			name: "floats",
			types: map[execute.ReferencePath]execute.DataType{
				"a": execute.TFloat,
				"b": execute.TFloat,
			},
			scope: map[execute.ReferencePath]execute.Value{
				"a": execute.Value{
					Type:  execute.TFloat,
					Value: float64(5.0),
				},
				"b": execute.Value{
					Type:  execute.TFloat,
					Value: float64(4.0),
				},
			},
			want: execute.Value{
				Type:  execute.TFloat,
				Value: float64(9.0),
			},
		},
		{
			name: "ints",
			types: map[execute.ReferencePath]execute.DataType{
				"a": execute.TInt,
				"b": execute.TInt,
			},
			scope: map[execute.ReferencePath]execute.Value{
				"a": execute.Value{
					Type:  execute.TInt,
					Value: int64(5),
				},
				"b": execute.Value{
					Type:  execute.TInt,
					Value: int64(4),
				},
			},
			want: execute.Value{
				Type:  execute.TInt,
				Value: int64(9),
			},
		},
		{
			name: "uints",
			types: map[execute.ReferencePath]execute.DataType{
				"a": execute.TUInt,
				"b": execute.TUInt,
			},
			scope: map[execute.ReferencePath]execute.Value{
				"a": execute.Value{
					Type:  execute.TUInt,
					Value: uint64(5),
				},
				"b": execute.Value{
					Type:  execute.TUInt,
					Value: uint64(4),
				},
			},
			want: execute.Value{
				Type:  execute.TUInt,
				Value: uint64(9),
			},
		},
	}

	//Reuse the same cache for all test cases
	cache := execute.NewCompilationCache(add, []execute.ReferencePath{"a", "b"})
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f0, err := cache.Compile(tc.types)
			if err != nil {
				t.Fatal(err)
			}
			f1, err := cache.Compile(tc.types)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(f0, f1) {
				t.Errorf("unexpected new compilation result")
			}

			got0, err := f0.Eval(tc.scope)
			if err != nil {
				t.Fatal(err)
			}
			got1, err := f1.Eval(tc.scope)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got0, tc.want) {
				t.Errorf("unexpected eval result -want/+got\n%s", cmp.Diff(tc.want, got0))
			}
			if !cmp.Equal(got0, got1) {
				t.Errorf("unexpected differing results -got0/+got1\n%s", cmp.Diff(got0, got1))
			}

		})
	}
}
