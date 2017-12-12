package ifql

import (
	"errors"
	"fmt"

	"github.com/influxdata/ifql/ast"
)

func Eval(program *ast.Program, scope *Scope, d Domain) error {
	ev := evaluator{
		d: d,
	}
	return ev.eval(program, scope.Nest())
}

// Domain represents any specific domain being used during evaluation.
type Domain interface{}

type evaluator struct {
	d Domain
}

func (ev evaluator) eval(program *ast.Program, scope *Scope) error {
	for _, stmt := range program.Body {
		if err := ev.evalStmt(stmt, scope); err != nil {
			return err
		}
	}
	return nil
}

func (ev evaluator) evalStmt(stmt ast.Statement, scope *Scope) error {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		if err := ev.doVariableDeclaration(s, scope); err != nil {
			return err
		}
	case *ast.ExpressionStatement:
		_, err := ev.doExpression(s.Expression, scope)
		if err != nil {
			return err
		}
	case *ast.BlockStatement:
		scope = scope.Nest()
		for _, stmt := range s.Body {
			if err := ev.evalStmt(stmt, scope); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported statement type %T", stmt)
	}
	return nil
}

func (ev evaluator) doVariableDeclaration(declarations *ast.VariableDeclaration, scope *Scope) error {
	for _, vd := range declarations.Declarations {
		value, err := ev.doExpression(vd.Init, scope)
		if err != nil {
			return err
		}
		scope.Set(vd.ID.Name, value)
	}
	return nil
}

func (ev evaluator) doExpression(expr ast.Expression, scope *Scope) (Value, error) {
	switch e := expr.(type) {
	case ast.Literal:
		return ev.doLiteral(e)
	case *ast.ArrayExpression:
		return ev.doArray(e, scope)
	case *ast.Identifier:
		value, ok := scope.Lookup(e.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", e.Name)
		}
		return value, nil
	case *ast.CallExpression:
		return ev.callFunction(e, scope)
	case *ast.MemberExpression:
		obj, err := ev.doExpression(e.Object, scope)
		if err != nil {
			return nil, err
		}
		p, err := propertyName(e)
		if err != nil {
			return nil, err
		}
		return obj.Property(p)
	case *ast.BinaryExpression:
		//TODO
		return nil, errors.New("not implemented")
	case *ast.LogicalExpression:
		return nil, errors.New("not implemented")
	case *ast.ArrowFunctionExpression:
		nested := scope.Nest()
		return value{
			t: TFunction,
			v: &arrowFunc{
				e: e,
				call: func(args Arguments, d Domain) (Value, error) {
					for _, p := range e.Params {
						v, err := args.GetRequired(p.Name)
						if err != nil {
							return nil, err
						}
						nested.Set(p.Name, v)
					}
					// TODO(nathanielc): How to handle function body that is a statement?
					return ev.doExpression(e.Body.(ast.Expression), nested)
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (ev evaluator) doArray(a *ast.ArrayExpression, scope *Scope) (Value, error) {
	array := Array{
		Type: TInvalid,
	}
	array.Elements = make([]Value, len(a.Elements))
	for i, el := range a.Elements {
		v, err := ev.doExpression(el, scope)
		if err != nil {
			return nil, err
		}
		if array.Type == TInvalid {
			array.Type = v.Type()
		}
		if array.Type != v.Type() {
			return nil, fmt.Errorf("cannot mix types in an array, found both %v and %v", array.Type, v.Type())
		}
		array.Elements[i] = v
	}
	return value{
		t: TArray,
		v: array,
	}, nil
}

func (ev evaluator) doLiteral(lit ast.Literal) (Value, error) {
	switch l := lit.(type) {
	case *ast.DateTimeLiteral:
		return value{
			t: TTime,
			v: l.Value,
		}, nil
	case *ast.DurationLiteral:
		return value{
			t: TDuration,
			v: l.Value,
		}, nil
	case *ast.NumberLiteral:
		return value{
			t: TFloat,
			v: l.Value,
		}, nil
	case *ast.IntegerLiteral:
		return value{
			t: TInt,
			v: l.Value,
		}, nil
	case *ast.StringLiteral:
		return value{
			t: TString,
			v: l.Value,
		}, nil
	case *ast.BooleanLiteral:
		return value{
			t: TBool,
			v: l.Value,
		}, nil
	// TODO(nathanielc): Support lists and maps
	default:
		return nil, fmt.Errorf("unknown literal type %T", lit)
	}

}

func (ev evaluator) callFunction(call *ast.CallExpression, scope *Scope) (Value, error) {
	callee, err := ev.doExpression(call.Callee, scope)
	if err != nil {
		return nil, err
	}
	if callee.Type() != TFunction {
		return nil, fmt.Errorf("cannot call function, value is of type %v", callee.Type())
	}
	f := callee.Value().(Function)
	arguments, err := ev.doArguments(call.Arguments, scope)
	if err != nil {
		return nil, err
	}
	v, err := f.Call(arguments, ev.d)
	if err != nil {
		return nil, err
	}
	if unused := arguments.listUnused(); len(unused) > 0 {
		return nil, fmt.Errorf("unused arguments %s", unused)
	}
	return v, nil
}

func (ev evaluator) doArguments(args []ast.Expression, scope *Scope) (Arguments, error) {
	if l := len(args); l > 1 {
		return nil, fmt.Errorf("arguments not a single object expression %v", args)
	} else if l == 0 {
		return newArguments(nil), nil
	}
	params, ok := args[0].(*ast.ObjectExpression)
	if !ok {
		return nil, fmt.Errorf("arguments not a valid object expression")
	}
	paramsMap := make(map[string]Value, len(params.Properties))
	for _, p := range params.Properties {
		value, err := ev.doExpression(p.Value, scope)
		if err != nil {
			return nil, err
		}
		if _, ok := paramsMap[p.Key.Name]; ok {
			return nil, fmt.Errorf("duplicate keyword parameter specified: %q", p.Key.Name)
		}
		paramsMap[p.Key.Name] = value
	}
	return newArguments(paramsMap), nil
}

func propertyName(m *ast.MemberExpression) (string, error) {
	switch p := m.Property.(type) {
	case *ast.Identifier:
		return p.Name, nil
	case *ast.StringLiteral:
		return p.Value, nil
	default:
		return "", fmt.Errorf("unsupported member property expression of type %T", m.Property)
	}
}

type Scope struct {
	parent *Scope
	values map[string]Value
}

func NewScope() *Scope {
	return &Scope{
		values: make(map[string]Value),
	}
}

func (s *Scope) Lookup(name string) (Value, bool) {
	if s == nil {
		return nil, false
	}
	v, ok := s.values[name]
	if !ok {
		return s.parent.Lookup(name)
	}
	return v, ok
}

func (s *Scope) Set(name string, value Value) {
	s.values[name] = value
}

// Nest returns a new nested scope.
func (s *Scope) Nest() *Scope {
	c := NewScope()
	c.parent = s
	return c
}

// Value represents any value that can be the result of evaluating any expression.
type Value interface {
	// Type reports the type of value
	Type() Type
	// Value returns the actual value represented.
	Value() interface{}
	// Property returns a new value which is a property of this value.
	Property(name string) (Value, error)
}

type value struct {
	t Type
	v interface{}
}

func (v value) Type() Type {
	return v.t
}
func (v value) Value() interface{} {
	return v.v
}
func (v value) Property(name string) (Value, error) {
	return nil, fmt.Errorf("property %q does not exist", name)
}

// Type represents the supported types within IFQL
type Type int

const (
	TInvalid  Type = iota // Go type nil
	TString               // Go type string
	TInt                  // Go type int64
	TFloat                // Go type float64
	TBool                 // Go type bool
	TTime                 // Go type time.Time
	TDuration             // Go type time.Duration
	TFunction             // Go type Function
	TArray                // Go type Array
	TMap                  // Go type Map
	endBuiltInTypes
)

var lastType = endBuiltInTypes
var extraTypes = make(map[string]Type)
var extraTypesLookup = make(map[Type]string)

func RegisterType(name string) Type {
	if _, ok := extraTypes[name]; ok {
		panic(fmt.Errorf("duplicate registration for ifql type %q", name))
	}
	lastType++
	extraTypes[name] = lastType
	extraTypesLookup[lastType] = name
	return lastType
}

// String converts Type into a string representation of the type's name
func (t Type) String() string {
	switch t {
	case TString:
		return "string"
	case TInt:
		return "int"
	case TFloat:
		return "float"
	case TBool:
		return "bool"
	case TTime:
		return "time"
	case TDuration:
		return "duration"
	case TFunction:
		return "function"
	case TArray:
		return "list"
	case TMap:
		return "map"
	default:
		name, ok := extraTypesLookup[t]
		if !ok {
			return fmt.Sprintf("unknown type %d", int(t))
		}
		return name
	}
}

// Function represents a callable type
type Function interface {
	Call(args Arguments, d Domain) (Value, error)
	// Resolve rewrites the function resolving any identifiers not listed in the function params.
	Resolve() (*ast.ArrowFunctionExpression, error)
}

type arrowFunc struct {
	e    *ast.ArrowFunctionExpression
	call func(Arguments, Domain) (Value, error)
}

func (f arrowFunc) Call(args Arguments, d Domain) (Value, error) {
	return f.call(args, d)
}

// Resolve rewrites the function resolving any identifiers not listed in the function params.
func (f arrowFunc) Resolve() (*ast.ArrowFunctionExpression, error) {
	//TODO(nathanielc): actually resolve the function
	return f.e, nil
}

// Array represents an sequence of elements
// All elements must be the same type
type Array struct {
	Type     Type
	Elements []Value
}

func (a Array) AsStrings() []string {
	if a.Type != TString {
		return nil
	}
	strs := make([]string, len(a.Elements))
	for i, v := range a.Elements {
		strs[i] = v.Value().(string)
	}
	return strs
}

// Map represents an association of keys to values of Type
// All elements must be the same type
type Map struct {
	Type     Type
	Elements map[string]Value
}

// Arguments provides access to the keyword arguments passed to a function.
// The Get{Type} methods return three values: the typed value of the arg,
// whether the argument was specified and any errors about the argument type.
// The GetRequired{Type} methods return only two values, the typed value of the arg and any errors, a missing argument is considered an error in this case.
type Arguments interface {
	Get(name string) (Value, bool)
	GetRequired(name string) (Value, error)

	GetString(name string) (string, bool, error)
	GetInt(name string) (int64, bool, error)
	GetFloat(name string) (float64, bool, error)
	GetBool(name string) (bool, bool, error)
	GetFunction(name string) (Function, bool, error)
	GetArray(name string, t Type) (Array, bool, error)

	GetRequiredString(name string) (string, error)
	GetRequiredInt(name string) (int64, error)
	GetRequiredFloat(name string) (float64, error)
	GetRequiredBool(name string) (bool, error)
	GetRequiredFunction(name string) (Function, error)
	GetRequiredArray(name string, t Type) (Array, error)

	// listUnused returns the list of provided arguments that were not used by the function.
	listUnused() []string
}

type arguments struct {
	params map[string]Value
	used   map[string]bool
}

func newArguments(params map[string]Value) *arguments {
	return &arguments{
		params: params,
		used:   make(map[string]bool, len(params)),
	}
}

func (a *arguments) Get(name string) (Value, bool) {
	a.used[name] = true
	v, ok := a.params[name]
	return v, ok
}

func (a *arguments) GetRequired(name string) (Value, error) {
	a.used[name] = true
	v, ok := a.params[name]
	if !ok {
		return nil, fmt.Errorf("missing required keyword argument %q", name)
	}
	return v, nil
}

func (a *arguments) GetString(name string) (string, bool, error) {
	v, ok, err := a.get(name, TString, false)
	if err != nil || !ok {
		return "", ok, err
	}
	return v.Value().(string), ok, nil
}
func (a *arguments) GetRequiredString(name string) (string, error) {
	v, _, err := a.get(name, TString, true)
	if err != nil {
		return "", err
	}
	return v.Value().(string), nil
}
func (a *arguments) GetInt(name string) (int64, bool, error) {
	v, ok, err := a.get(name, TInt, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Value().(int64), ok, nil
}
func (a *arguments) GetRequiredInt(name string) (int64, error) {
	v, _, err := a.get(name, TInt, true)
	if err != nil {
		return 0, err
	}
	return v.Value().(int64), nil
}
func (a *arguments) GetFloat(name string) (float64, bool, error) {
	v, ok, err := a.get(name, TFloat, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Value().(float64), ok, nil
}
func (a *arguments) GetRequiredFloat(name string) (float64, error) {
	v, _, err := a.get(name, TFloat, true)
	if err != nil {
		return 0, err
	}
	return v.Value().(float64), nil
}
func (a *arguments) GetBool(name string) (bool, bool, error) {
	v, ok, err := a.get(name, TBool, false)
	if err != nil || !ok {
		return false, ok, err
	}
	return v.Value().(bool), ok, nil
}
func (a *arguments) GetRequiredBool(name string) (bool, error) {
	v, _, err := a.get(name, TBool, true)
	if err != nil {
		return false, err
	}
	return v.Value().(bool), nil
}

func (a *arguments) GetArray(name string, t Type) (Array, bool, error) {
	v, ok, err := a.get(name, TArray, false)
	if err != nil || !ok {
		return Array{}, ok, err
	}
	arr := v.Value().(Array)
	if arr.Type != t {
		return Array{}, true, fmt.Errorf("keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type)
	}
	return v.Value().(Array), ok, nil
}
func (a *arguments) GetRequiredArray(name string, t Type) (Array, error) {
	v, _, err := a.get(name, TArray, true)
	if err != nil {
		return Array{}, err
	}
	arr := v.Value().(Array)
	if arr.Type != t {
		return Array{}, fmt.Errorf("keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type)
	}
	return arr, nil
}
func (a *arguments) GetFunction(name string) (Function, bool, error) {
	v, ok, err := a.get(name, TFunction, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Value().(Function), ok, nil
}
func (a *arguments) GetRequiredFunction(name string) (Function, error) {
	v, _, err := a.get(name, TFunction, true)
	if err != nil {
		return nil, err
	}
	return v.Value().(Function), nil
}

func (a *arguments) get(name string, typ Type, required bool) (Value, bool, error) {
	a.used[name] = true
	v, ok := a.params[name]
	if !ok {
		if required {
			return nil, false, fmt.Errorf("missing required keyword argument %q", name)
		}
		return nil, false, nil
	}
	if v.Type() != typ {
		return nil, true, fmt.Errorf("keyword argument %q should be of type %v, but got %v", name, typ, v.Type())
	}
	return v, true, nil
}

func (a *arguments) listUnused() []string {
	var unused []string
	for k := range a.params {
		if !a.used[k] {
			unused = append(unused, k)
		}
	}

	return unused
}
