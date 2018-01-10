package execute_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query/execute"
)

func TestRowMapFn(t *testing.T) {
	testCases := []struct {
		name string
		fn   *ast.ArrowFunctionExpression
		cols []execute.ColMeta
		row  []interface{}
		want execute.Map
	}{
		{
			name: "literal single expression",
			fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body:   &ast.IntegerLiteral{Value: 42},
			},
			cols: []execute.ColMeta{
				{Label: "_time", Kind: execute.TimeColKind, Type: execute.TTime},
				{Label: "_value", Kind: execute.ValueColKind, Type: execute.TInt},
			},
			row: []interface{}{execute.Time(0), int64(0)},
			want: execute.Map{
				Meta: execute.MapMeta{
					Properties: []execute.MapPropertyMeta{
						{Key: "_value", Type: execute.TInt},
					},
				},
				Values: map[string]execute.Value{
					"_value": execute.Value{
						Type:  execute.TInt,
						Value: int64(42),
					},
				},
			},
		},
		{
			name: "simple single expression",
			fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body: &ast.BinaryExpression{
					Operator: ast.AdditionOperator,
					Left: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "r"},
						Property: &ast.Identifier{Name: "_value"},
					},
					Right: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "r"},
						Property: &ast.Identifier{Name: "_value"},
					},
				},
			},
			cols: []execute.ColMeta{
				{Label: "_time", Kind: execute.TimeColKind, Type: execute.TTime},
				{Label: "_value", Kind: execute.ValueColKind, Type: execute.TInt},
			},
			row: []interface{}{execute.Time(0), int64(1)},
			want: execute.Map{
				Meta: execute.MapMeta{
					Properties: []execute.MapPropertyMeta{
						{Key: "_value", Type: execute.TInt},
					},
				},
				Values: map[string]execute.Value{
					"_value": execute.Value{
						Type:  execute.TInt,
						Value: int64(2),
					},
				},
			},
		},
		{
			name: "simple map expression",
			fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body: &ast.ObjectExpression{
					Properties: []*ast.Property{
						{
							Key: &ast.Identifier{Name: "a"},
							Value: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "r"},
									Property: &ast.Identifier{Name: "_value"},
								},
								Right: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "r"},
									Property: &ast.Identifier{Name: "_value"},
								},
							},
						},
						{
							Key: &ast.Identifier{Name: "b"},
							Value: &ast.BinaryExpression{
								Operator: ast.MultiplicationOperator,
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "r"},
									Property: &ast.Identifier{Name: "_value"},
								},
								Right: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "r"},
									Property: &ast.Identifier{Name: "_value"},
								},
							},
						},
					},
				},
			},
			cols: []execute.ColMeta{
				{Label: "_time", Kind: execute.TimeColKind, Type: execute.TTime},
				{Label: "_value", Kind: execute.ValueColKind, Type: execute.TInt},
			},
			row: []interface{}{execute.Time(0), int64(1)},
			want: execute.Map{
				Meta: execute.MapMeta{
					Properties: []execute.MapPropertyMeta{
						{Key: "a", Type: execute.TInt},
						{Key: "b", Type: execute.TInt},
					},
				},
				Values: map[string]execute.Value{
					"a": execute.Value{
						Type:  execute.TInt,
						Value: int64(2),
					},
					"b": execute.Value{
						Type:  execute.TInt,
						Value: int64(1),
					},
				},
			},
		},
		{
			name: "single expression with call",
			fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body: &ast.BinaryExpression{
					Operator: ast.AdditionOperator,
					Left: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "r"},
						Property: &ast.Identifier{Name: "_value"},
					},
					Right: &ast.CallExpression{
						Callee: &ast.ArrowFunctionExpression{
							Params: []*ast.Property{{Key: &ast.Identifier{Name: "x"}}},
							Body: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "x"},
								Property: &ast.Identifier{Name: "_value"},
							},
						},
						Arguments: []ast.Expression{&ast.ObjectExpression{
							Properties: []*ast.Property{{
								Key:   &ast.Identifier{Name: "x"},
								Value: &ast.Identifier{Name: "r"},
							}},
						}},
					},
				},
			},
			cols: []execute.ColMeta{
				{Label: "_time", Kind: execute.TimeColKind, Type: execute.TTime},
				{Label: "_value", Kind: execute.ValueColKind, Type: execute.TInt},
			},
			row: []interface{}{execute.Time(0), int64(1)},
			want: execute.Map{
				Meta: execute.MapMeta{
					Properties: []execute.MapPropertyMeta{
						{Key: "_value", Type: execute.TInt},
					},
				},
				Values: map[string]execute.Value{
					"_value": execute.Value{
						Type:  execute.TInt,
						Value: int64(2),
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rf, err := execute.NewRowMapFn(tc.fn)
			if err != nil {
				t.Fatal(err)
			}

			if err := rf.Prepare(tc.cols); err != nil {
				t.Fatal(err)
			}
			rr := rowReader{
				cols: tc.cols,
				rows: [][]interface{}{tc.row},
			}
			got, err := rf.Eval(0, rr)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tc.want) {
				t.Errorf("unexpected map value -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

type rowReader struct {
	cols []execute.ColMeta
	rows [][]interface{}
}

func (r rowReader) Cols() []execute.ColMeta {
	return r.cols
}

func (r rowReader) AtBool(i int, j int) bool {
	return r.rows[i][j].(bool)
}

func (r rowReader) AtInt(i int, j int) int64 {
	return r.rows[i][j].(int64)
}

func (r rowReader) AtUInt(i int, j int) uint64 {
	return r.rows[i][j].(uint64)
}

func (r rowReader) AtFloat(i int, j int) float64 {
	return r.rows[i][j].(float64)
}

func (r rowReader) AtString(i int, j int) string {
	return r.rows[i][j].(string)
}

func (r rowReader) AtTime(i int, j int) execute.Time {
	return r.rows[i][j].(execute.Time)
}

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
