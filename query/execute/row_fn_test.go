package execute_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/compiler"
	"github.com/influxdata/ifql/semantic"
)

func TestCompilationCache(t *testing.T) {
	add := &semantic.ArrowFunctionExpression{
		Params: []*semantic.FunctionParam{
			{Key: &semantic.Identifier{Name: "a"}},
			{Key: &semantic.Identifier{Name: "b"}},
		},
		Body: &semantic.BinaryExpression{
			Operator: ast.AdditionOperator,
			Left:     &semantic.Identifier{Name: "a"},
			Right:    &semantic.Identifier{Name: "b"},
		},
	}
	testCases := []struct {
		name  string
		types map[compiler.ReferencePath]compiler.Type
		scope map[compiler.ReferencePath]compiler.Value
		want  compiler.Value
	}{
		{
			name: "floats",
			types: map[compiler.ReferencePath]compiler.Type{
				"a": compiler.TFloat,
				"b": compiler.TFloat,
			},
			scope: map[compiler.ReferencePath]compiler.Value{
				"a": compiler.Value{
					Type:  compiler.TFloat,
					Value: float64(5.0),
				},
				"b": compiler.Value{
					Type:  compiler.TFloat,
					Value: float64(4.0),
				},
			},
			want: compiler.Value{
				Type:  compiler.TFloat,
				Value: float64(9.0),
			},
		},
		{
			name: "ints",
			types: map[compiler.ReferencePath]compiler.Type{
				"a": compiler.TInt,
				"b": compiler.TInt,
			},
			scope: map[compiler.ReferencePath]compiler.Value{
				"a": compiler.Value{
					Type:  compiler.TInt,
					Value: int64(5),
				},
				"b": compiler.Value{
					Type:  compiler.TInt,
					Value: int64(4),
				},
			},
			want: compiler.Value{
				Type:  compiler.TInt,
				Value: int64(9),
			},
		},
		{
			name: "uints",
			types: map[compiler.ReferencePath]compiler.Type{
				"a": compiler.TUInt,
				"b": compiler.TUInt,
			},
			scope: map[compiler.ReferencePath]compiler.Value{
				"a": compiler.Value{
					Type:  compiler.TUInt,
					Value: uint64(5),
				},
				"b": compiler.Value{
					Type:  compiler.TUInt,
					Value: uint64(4),
				},
			},
			want: compiler.Value{
				Type:  compiler.TUInt,
				Value: uint64(9),
			},
		},
	}

	//Reuse the same cache for all test cases
	cache := compiler.NewCompilationCache(add, []compiler.ReferencePath{"a", "b"})
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
