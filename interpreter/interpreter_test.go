package interpreter_test

import (
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ast/asttest"
	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
)

var testScope = interpreter.NewScope()

func init() {
	testScope.Set("fortyTwo", function{
		name: "fortyTwo",
		call: func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
			return interpreter.NewFloatValue(42.0), nil
		},
	})
	testScope.Set("six", function{
		name: "six",
		call: func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
			return interpreter.NewFloatValue(6.0), nil
		},
	})
	testScope.Set("nine", function{
		name: "nine",
		call: func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
			return interpreter.NewFloatValue(9.0), nil
		},
	})
	testScope.Set("fail", function{
		name: "fail",
		call: func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
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
		{
			name: "import",
			query: `
            import "boolean" 1.0.0
            boolean.true or fail()
            boolean.false and fail()
			`,
		},
		{
			name: "import as",
			query: `
            import "boolean" 1.0.0 as b
            b.true or fail()
            b.false and fail()
			`,
		},
		{
			name: "import err",
			query: `
            import "nonexistant" 1.0.0
			`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			program, err := parser.NewAST(tc.query)
			if err != nil {
				t.Fatal(err)
			}

			err = interpreter.Eval(program, testScope.Nest(), nil, testImporter{})
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

type testImporter struct{}

func (i testImporter) Import(importPath, dir string) (ifql.Package, error) {
	scope := ifql.NewScope()
	// Define some test packages to import
	switch importPath {
	case "boolean":
		scope.Set("true", ifql.NewBoolValue(true))
		scope.Set("false", ifql.NewBoolValue(false))
	default:
		return nil, fmt.Errorf("unknown package %q", importPath)
	}
	return &pkg{
		name:  path.Base(importPath),
		path:  importPath,
		scope: scope,
	}, nil
}

type pkg struct {
	name  string
	path  string
	scope *ifql.Scope
}

func (p *pkg) Name() string {
	return p.name
}
func (p *pkg) Path() string {
	return p.path
}
func (p *pkg) Scope() *ifql.Scope {
	return p.scope
}
func (p *pkg) SetScope(s *ifql.Scope) {
	p.scope = s
}

func (p *pkg) Complete() bool {
	return true
}
func (p *pkg) Program() *ast.Program {
	return nil
}

func TestFunction_Resolve(t *testing.T) {
	var got *ast.ArrowFunctionExpression
	scope := interpreter.NewScope()
	scope.Set("resolver", function{
		name: "resolver",
		call: func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
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

	program, err := parser.NewAST(`
	x = 42
	resolver(f: (r) => r + x)
`)
	if err != nil {
		t.Fatal(err)
	}

	if err := interpreter.Eval(program, scope, nil, nil); err != nil {
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
	call func(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error)
}

func (f function) Type() interpreter.Type {
	return interpreter.TFunction
}

func (f function) Value() interface{} {
	return f
}
func (f function) Property(name string) (interpreter.Value, error) {
	return nil, fmt.Errorf("property %q does not exist", name)
}

func (f function) Call(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
	return f.call(args, d)
}

func (f function) Resolve() (*ast.ArrowFunctionExpression, error) {
	return nil, fmt.Errorf("function %q cannot be resolved", f.name)
}
