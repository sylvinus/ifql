package ifql_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ast/asttest"
	"github.com/influxdata/ifql/ifql"
)

var testScope = ifql.NewScope()

func init() {
	testScope.Set("fortyTwo", function{
		name: "fortyTwo",
		call: func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
			return ifql.NewFloatValue(42.0), nil
		},
	})
	testScope.Set("six", function{
		name: "six",
		call: func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
			return ifql.NewFloatValue(6.0), nil
		},
	})
	testScope.Set("nine", function{
		name: "nine",
		call: func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
			return ifql.NewFloatValue(9.0), nil
		},
	})
	testScope.Set("fail", function{
		name: "fail",
		call: func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
			return nil, errors.New("fail")
		},
	})
}

// TestEval tests whether a program can run to completion or not
func TestEval(t *testing.T) {
	testCases := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:  "call function",
			query: "six()",
		},
		{
			name:    "call function with fail",
			query:   "fail()",
			wantErr: true,
		},
		{
			name: "reassign nested scope",
			query: `
			six = six()
			six()
			`,
			wantErr: true,
		},
		{
			name: "binary expressions",
			query: `
			six = six()
			nine = nine()

			answer = fortyTwo() == six * nine
			`,
		},
		{
			name: "logcial expressions short circuit",
			query: `
            six = six()
            nine = nine()

            answer = (not (fortyTwo() == six * nine)) or fail()
			`,
		},
		{
			name: "arrow function",
			query: `
            plusSix = (r) => r + six()
            plusSix(r:1.0) == 7.0 or fail()
			`,
		},
		{
			name: "arrow function block",
			query: `
            f = (r) => {
                r2 = r * r
                return (r - r2) / r2
            }
            f(r:2.0) == -0.5 or fail()
			`,
		},
		{
			name: "arrow function with default param",
			query: `
            addN = (r,n=4) => r + n
            addN(r:2) == 6 or fail()
			addN(r:3,n:1) == 4 or fail()
			`,
		},
		{
			name: "extra statements after return",
			query: `
            f = (r) => {
                r2 = r * r
                return (r - r2) / r2
                x = r2 * r
            }
            f(r:2.0)
			`,
			wantErr: true,
		},
		{
			name: "scope closing",
			query: `
			x = 5
            plusX = (r) => r + x
            plusX(r:2) == 7 or fail()
			`,
		},
		{
			name: "return map from func",
			query: `
            toMap = (a,b) => ({
                a: a,
                b: b,
            })
            m = toMap(a:1, b:false)
            m.a == 1 or fail()
            not m.b or fail()
			`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			program, err := ifql.NewAST(tc.query)
			if err != nil {
				t.Fatal(err)
			}

			err = ifql.Eval(program, testScope.Nest(), nil)
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
		})
	}

}
func TestFunction_Resolve(t *testing.T) {
	var got *ast.ArrowFunctionExpression
	scope := ifql.NewScope()
	scope.Set("resolver", function{
		name: "resolver",
		call: func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
			f, err := args.GetRequiredFunction("f")
			if err != nil {
				return nil, err
			}
			got, err = f.Resolve()
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	})

	program, err := ifql.NewAST(`
	x = 42
	resolver(f: (r) => r + x)
`)
	if err != nil {
		t.Fatal(err)
	}

	if err := ifql.Eval(program, scope, nil); err != nil {
		t.Fatal(err)
	}

	want := &ast.ArrowFunctionExpression{
		Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
		Body: &ast.BinaryExpression{
			Operator: ast.AdditionOperator,
			Left:     &ast.Identifier{Name: "r"},
			Right:    &ast.IntegerLiteral{Value: 42},
		},
	}
	if !cmp.Equal(want, got, asttest.CompareOptions...) {
		t.Errorf("unexpected resoved function: -want/+got\n%s", cmp.Diff(want, got, asttest.CompareOptions...))
	}
}

type function struct {
	name string
	call func(args ifql.Arguments, d ifql.Domain) (ifql.Value, error)
}

func (f function) Type() ifql.Type {
	return ifql.TFunction
}

func (f function) Value() interface{} {
	return f
}
func (f function) Property(name string) (ifql.Value, error) {
	return nil, fmt.Errorf("property %q does not exist", name)
}

func (f function) Call(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
	return f.call(args, d)
}

func (f function) Resolve() (*ast.ArrowFunctionExpression, error) {
	return nil, fmt.Errorf("function %q cannot be resolved", f.name)
}
