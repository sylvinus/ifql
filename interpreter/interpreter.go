package interpreter

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/semantic"
	"github.com/pkg/errors"
)

func Eval(program *semantic.Program, scope *Scope, d Domain) error {
	itrp := interpreter{
		d: d,
	}
	return itrp.eval(program, scope)
}

// Domain represents any specific domain being used during evaluation.
type Domain interface{}

type interpreter struct {
	d Domain
}

func (itrp interpreter) eval(program *semantic.Program, scope *Scope) error {
	for _, stmt := range program.Body {
		if err := itrp.doStatement(stmt, scope); err != nil {
			return err
		}
	}
	return nil
}

func (itrp interpreter) doStatement(stmt semantic.Statement, scope *Scope) error {
	switch s := stmt.(type) {
	case *semantic.VariableDeclaration:
		if err := itrp.doVariableDeclaration(s, scope); err != nil {
			return err
		}
	case *semantic.ExpressionStatement:
		_, err := itrp.doExpression(s.Expression, scope)
		if err != nil {
			return err
		}
	case *semantic.BlockStatement:
		nested := scope.Nest()
		for i, stmt := range s.Body {
			if err := itrp.doStatement(stmt, nested); err != nil {
				return err
			}
			// Validate a return statement is the last statement
			if _, ok := stmt.(*semantic.ReturnStatement); ok {
				if i != len(s.Body)-1 {
					return errors.New("return statement is not the last statement in the block")
				}
			}
		}
		// Propgate any return value from the nested scope out.
		// Since a return statement is always last we do not have to worry about overriding an existing return value.
		scope.SetReturn(nested.Return())
	case *semantic.ReturnStatement:
		v, err := itrp.doExpression(s.Argument, scope)
		if err != nil {
			return err
		}
		scope.SetReturn(v)
	default:
		return fmt.Errorf("unsupported statement type %T", stmt)
	}
	return nil
}

func (itrp interpreter) doVariableDeclaration(declarations *semantic.VariableDeclaration, scope *Scope) error {
	for _, vd := range declarations.Declarations {
		value, err := itrp.doExpression(vd.Init, scope)
		if err != nil {
			return err
		}
		scope.Set(vd.ID.Name, value)
	}
	return nil
}

func (itrp interpreter) doExpression(expr semantic.Expression, scope *Scope) (Value, error) {
	switch e := expr.(type) {
	case semantic.Literal:
		return itrp.doLiteral(e)
	case *semantic.ArrayExpression:
		return itrp.doArray(e, scope)
	case *semantic.Identifier:
		value, ok := scope.Lookup(e.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", e.Name)
		}
		return value, nil
	case *semantic.CallExpression:
		v, err := itrp.doCall(e, scope)
		if err != nil {
			// Determine function name
			return nil, errors.Wrapf(err, "error calling funcion %q", functionName(e))
		}
		return v, nil
	case *semantic.MemberExpression:
		obj, err := itrp.doExpression(e.Object, scope)
		if err != nil {
			return nil, err
		}
		return obj.Property(e.Property)
	case *semantic.ObjectExpression:
		return itrp.doMap(e, scope)
	case *semantic.UnaryExpression:
		v, err := itrp.doExpression(e.Argument, scope)
		if err != nil {
			return nil, err
		}
		switch e.Operator {
		case ast.NotOperator:
			if v.Type() != TBool {
				return nil, fmt.Errorf("operand to unary expression is not a boolean value, got %v", v.Type())
			}
			return NewBoolValue(!v.Value().(bool)), nil
		case ast.SubtractionOperator:
			switch t := v.Type(); t {
			case TInt:
				return NewIntValue(-v.Value().(int64)), nil
			case TFloat:
				return NewFloatValue(-v.Value().(float64)), nil
			case TDuration:
				return NewDurationValue(-v.Value().(time.Duration)), nil
			default:
				return nil, fmt.Errorf("operand to unary expression is not a number value, got %v", v.Type())
			}
		default:
			return nil, fmt.Errorf("unsupported operator %q to unary expression", e.Operator)
		}

	case *semantic.BinaryExpression:
		l, err := itrp.doExpression(e.Left, scope)
		if err != nil {
			return nil, err
		}

		r, err := itrp.doExpression(e.Right, scope)
		if err != nil {
			return nil, err
		}

		bf, ok := binaryFuncLookup[binaryFuncSignature{
			operator: e.Operator,
			left:     l.Type(),
			right:    r.Type(),
		}]
		if !ok {
			return nil, fmt.Errorf("unsupported binary operation: %v %v %v", l.Type(), e.Operator, r.Type())
		}
		return bf(l, r), nil
	case *semantic.LogicalExpression:
		l, err := itrp.doExpression(e.Left, scope)
		if err != nil {
			return nil, err
		}
		if l.Type() != TBool {
			return nil, fmt.Errorf("left operand to logcial expression is not a boolean value, got %v", l.Type())
		}
		left := l.Value().(bool)

		if e.Operator == ast.AndOperator && !left {
			// Early return
			return NewBoolValue(false), nil
		} else if e.Operator == ast.OrOperator && left {
			// Early return
			return NewBoolValue(true), nil
		}

		r, err := itrp.doExpression(e.Right, scope)
		if err != nil {
			return nil, err
		}
		if r.Type() != TBool {
			return nil, errors.New("right operand to logcial expression is not a boolean value")
		}
		right := r.Value().(bool)

		switch e.Operator {
		case ast.AndOperator:
			return NewBoolValue(left && right), nil
		case ast.OrOperator:
			return NewBoolValue(left || right), nil
		default:
			return nil, fmt.Errorf("invalid logical operator %v", e.Operator)
		}
	case *semantic.ArrowFunctionExpression:
		return value{
			t: TFunction,
			v: arrowFunc{
				e:     e,
				scope: scope.Nest(),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (itrp interpreter) doArray(a *semantic.ArrayExpression, scope *Scope) (Value, error) {
	array := Array{
		Type:     TInvalid,
		Elements: make([]Value, len(a.Elements)),
	}
	for i, el := range a.Elements {
		v, err := itrp.doExpression(el, scope)
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

func (itrp interpreter) doMap(m *semantic.ObjectExpression, scope *Scope) (Value, error) {
	mapValue := Map{
		Elements: make(map[string]Value, len(m.Properties)),
	}
	for _, p := range m.Properties {
		v, err := itrp.doExpression(p.Value, scope)
		if err != nil {
			return nil, err
		}
		if _, ok := mapValue.Elements[p.Key.Name]; ok {
			return nil, fmt.Errorf("duplicate key in map: %q", p.Key.Name)
		}
		mapValue.Elements[p.Key.Name] = v
	}
	return mapValue, nil
}

func (itrp interpreter) doLiteral(lit semantic.Literal) (Value, error) {
	switch l := lit.(type) {
	case *semantic.DateTimeLiteral:
		return value{
			t: TTime,
			v: l.Value,
		}, nil
	case *semantic.DurationLiteral:
		return value{
			t: TDuration,
			v: l.Value,
		}, nil
	case *semantic.FloatLiteral:
		return value{
			t: TFloat,
			v: l.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return value{
			t: TInt,
			v: l.Value,
		}, nil
	case *semantic.UnsignedIntegerLiteral:
		return value{
			t: TUInt,
			v: l.Value,
		}, nil
	case *semantic.StringLiteral:
		return value{
			t: TString,
			v: l.Value,
		}, nil
	case *semantic.BooleanLiteral:
		return value{
			t: TBool,
			v: l.Value,
		}, nil
	// TODO(nathanielc): Support lists and maps
	default:
		return nil, fmt.Errorf("unknown literal type %T", lit)
	}

}

func functionName(call *semantic.CallExpression) string {
	switch callee := call.Callee.(type) {
	case *semantic.Identifier:
		return callee.Name
	case *semantic.MemberExpression:
		return callee.Property
	default:
		return "<anonymous function>"
	}
}

func (itrp interpreter) doCall(call *semantic.CallExpression, scope *Scope) (Value, error) {
	callee, err := itrp.doExpression(call.Callee, scope)
	if err != nil {
		return nil, err
	}
	if callee.Type() != TFunction {
		return nil, fmt.Errorf("cannot call function, value is of type %v", callee.Type())
	}
	f := callee.Value().(Function)
	arguments, err := itrp.doArguments(call.Arguments, scope)
	if err != nil {
		return nil, err
	}

	// Check if the function is an arrowFunc and rebind it.
	if af, ok := f.(arrowFunc); ok {
		af.itrp = itrp
		f = af
	}

	// Call the function
	v, err := f.Call(arguments, itrp.d)
	if err != nil {
		return nil, err
	}
	if unused := arguments.listUnused(); len(unused) > 0 {
		return nil, fmt.Errorf("unused arguments %s", unused)
	}
	return v, nil
}

func (itrp interpreter) doArguments(args *semantic.ObjectExpression, scope *Scope) (Arguments, error) {
	if args == nil || len(args.Properties) == 0 {
		return newArguments(nil), nil
	}
	paramsMap := make(map[string]Value, len(args.Properties))
	for _, p := range args.Properties {
		value, err := itrp.doExpression(p.Value, scope)
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

type Scope struct {
	parent      *Scope
	values      map[string]Value
	returnValue Value
}

func NewScope() *Scope {
	return &Scope{
		values:      make(map[string]Value),
		returnValue: value{t: TInvalid},
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

// SetReturn sets the return value of this scope.
func (s *Scope) SetReturn(value Value) {
	s.returnValue = value
}

// Return reports the return value for this scope. If no return value has been set a value with type TInvalid is returned.
func (s *Scope) Return() Value {
	return s.returnValue
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

// Type represents the builtin supported types within IFQL
type Type int

const (
	TInvalid  Type = iota // Go type nil
	TString               // Go type string
	TInt                  // Go type int64
	TUInt                 // Go type uint64
	TFloat                // Go type float64
	TBool                 // Go type bool
	TTime                 // Go type time.Time
	TDuration             // Go type time.Duration
	TFunction             // Go type Function
	TArray                // Go type Array
	TMap                  // Go type Map
	endBuiltInTypes
)

func NewBoolValue(v bool) Value {
	return value{
		t: TBool,
		v: v,
	}
}
func NewIntValue(v int64) Value {
	return value{
		t: TInt,
		v: v,
	}
}
func NewUIntValue(v uint64) Value {
	return value{
		t: TUInt,
		v: v,
	}
}
func NewFloatValue(v float64) Value {
	return value{
		t: TFloat,
		v: v,
	}
}
func NewStringValue(v string) Value {
	return value{
		t: TString,
		v: v,
	}
}
func NewTimeValue(v time.Time) Value {
	return value{
		t: TTime,
		v: v,
	}
}
func NewDurationValue(v time.Duration) Value {
	return value{
		t: TDuration,
		v: v,
	}
}

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
	case TInvalid:
		return "invalid"
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
		return "array"
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
	Resolve() (*semantic.ArrowFunctionExpression, error)
}

type arrowFunc struct {
	e     *semantic.ArrowFunctionExpression
	scope *Scope
	call  func(Arguments, Domain) (Value, error)

	itrp interpreter
}

func (f arrowFunc) Call(args Arguments, d Domain) (Value, error) {
	for _, p := range f.e.Params {
		if p.Default == nil {
			v, err := args.GetRequired(p.Key.Name)
			if err != nil {
				return nil, err
			}
			f.scope.Set(p.Key.Name, v)
		} else {
			v, ok := args.Get(p.Key.Name)
			if !ok {
				// Use default value
				var err error
				v, err = f.itrp.doLiteral(p.Default)
				if err != nil {
					return nil, err
				}
			}
			f.scope.Set(p.Key.Name, v)
		}
	}
	switch n := f.e.Body.(type) {
	case semantic.Expression:
		return f.itrp.doExpression(n, f.scope)
	case semantic.Statement:
		err := f.itrp.doStatement(n, f.scope)
		if err != nil {
			return nil, err
		}
		v := f.scope.Return()
		if v.Type() == TInvalid {
			return nil, errors.New("arrow function has no return value")
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported arrow function body type %T", f.e.Body)
	}
}

// Resolve rewrites the function resolving any identifiers not listed in the function params.
func (f arrowFunc) Resolve() (*semantic.ArrowFunctionExpression, error) {
	n := f.e.Copy()
	node, err := f.resolveIdentifiers(n)
	if err != nil {
		return nil, err
	}
	return node.(*semantic.ArrowFunctionExpression), nil
}

func (f arrowFunc) resolveIdentifiers(n semantic.Node) (semantic.Node, error) {
	switch n := n.(type) {
	case *semantic.Identifier:
		for _, p := range f.e.Params {
			if n.Name == p.Key.Name {
				// Identifier is a parameter do not resolve
				return n, nil
			}
		}
		v, ok := f.scope.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("name %q does not exist in scope", n.Name)
		}
		return resolveValue(v)
	case *semantic.BlockStatement:
		for i, s := range n.Body {
			node, err := f.resolveIdentifiers(s)
			if err != nil {
				return nil, err
			}
			n.Body[i] = node.(semantic.Statement)
		}
	case *semantic.ExpressionStatement:
		node, err := f.resolveIdentifiers(n.Expression)
		if err != nil {
			return nil, err
		}
		n.Expression = node.(semantic.Expression)
	case *semantic.ReturnStatement:
		node, err := f.resolveIdentifiers(n.Argument)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)
	case *semantic.VariableDeclaration:
		for i, d := range n.Declarations {
			node, err := f.resolveIdentifiers(d)
			if err != nil {
				return nil, err
			}
			n.Declarations[i] = node.(*semantic.VariableDeclarator)
		}
	case *semantic.VariableDeclarator:
		node, err := f.resolveIdentifiers(n.Init)
		if err != nil {
			return nil, err
		}
		n.Init = node.(semantic.Expression)
	case *semantic.CallExpression:
		node, err := f.resolveIdentifiers(n.Arguments)
		if err != nil {
			return nil, err
		}
		n.Arguments = node.(*semantic.ObjectExpression)
	case *semantic.ArrowFunctionExpression:
		node, err := f.resolveIdentifiers(n.Body)
		if err != nil {
			return nil, err
		}
		n.Body = node
	case *semantic.BinaryExpression:
		node, err := f.resolveIdentifiers(n.Left)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Right)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.UnaryExpression:
		node, err := f.resolveIdentifiers(n.Argument)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)
	case *semantic.LogicalExpression:
		node, err := f.resolveIdentifiers(n.Left)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)
		node, err = f.resolveIdentifiers(n.Right)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.ArrayExpression:
		for i, el := range n.Elements {
			node, err := f.resolveIdentifiers(el)
			if err != nil {
				return nil, err
			}
			n.Elements[i] = node.(semantic.Expression)
		}
	case *semantic.ObjectExpression:
		for i, p := range n.Properties {
			node, err := f.resolveIdentifiers(p)
			if err != nil {
				return nil, err
			}
			n.Properties[i] = node.(*semantic.Property)
		}
	case *semantic.ConditionalExpression:
		node, err := f.resolveIdentifiers(n.Test)
		if err != nil {
			return nil, err
		}
		n.Test = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Alternate)
		if err != nil {
			return nil, err
		}
		n.Alternate = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Consequent)
		if err != nil {
			return nil, err
		}
		n.Consequent = node.(semantic.Expression)
	case *semantic.Property:
		node, err := f.resolveIdentifiers(n.Value)
		if err != nil {
			return nil, err
		}
		n.Value = node.(semantic.Expression)
	}
	return n, nil
}

func resolveValue(v Value) (semantic.Node, error) {
	switch t := v.Type(); t {
	case TString:
		return &semantic.StringLiteral{
			Value: v.Value().(string),
		}, nil
	case TInt:
		return &semantic.IntegerLiteral{
			Value: v.Value().(int64),
		}, nil
	case TUInt:
		return &semantic.UnsignedIntegerLiteral{
			Value: v.Value().(uint64),
		}, nil
	case TFloat:
		return &semantic.FloatLiteral{
			Value: v.Value().(float64),
		}, nil
	case TBool:
		return &semantic.BooleanLiteral{
			Value: v.Value().(bool),
		}, nil
	case TTime:
		return &semantic.DateTimeLiteral{
			Value: v.Value().(time.Time),
		}, nil
	case TDuration:
		return &semantic.DurationLiteral{
			Value: v.Value().(time.Duration),
		}, nil
	case TFunction:
		return v.Value().(Function).Resolve()
	case TArray:
		arr := v.Value().(Array)
		node := new(semantic.ArrayExpression)
		node.Elements = make([]semantic.Expression, len(arr.Elements))
		for i, el := range arr.Elements {
			n, err := resolveValue(el)
			if err != nil {
				return nil, err
			}
			node.Elements[i] = n.(semantic.Expression)
		}
		return node, nil
	case TMap:
		m := v.Value().(Map)
		node := new(semantic.ObjectExpression)
		node.Properties = make([]*semantic.Property, 0, len(m.Elements))
		for k, el := range m.Elements {
			n, err := resolveValue(el)
			if err != nil {
				return nil, err
			}
			node.Properties = append(node.Properties, &semantic.Property{
				Key:   &semantic.Identifier{Name: k},
				Value: n.(semantic.Expression),
			})
		}
		return node, nil
	default:
		return nil, fmt.Errorf("cannot resove value of type %v", t)
	}
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

// Map represents an association of keys to values.
// Map values may be of any type.
type Map struct {
	Elements map[string]Value
}

func (m Map) Type() Type {
	return TMap
}
func (m Map) Value() interface{} {
	return m
}
func (m Map) Property(name string) (Value, error) {
	v, ok := m.Elements[name]
	if ok {
		return v, nil
	}
	return nil, fmt.Errorf("property %q does not exist", name)
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
	GetMap(name string) (Map, bool, error)

	GetRequiredString(name string) (string, error)
	GetRequiredInt(name string) (int64, error)
	GetRequiredFloat(name string) (float64, error)
	GetRequiredBool(name string) (bool, error)
	GetRequiredFunction(name string) (Function, error)
	GetRequiredArray(name string, t Type) (Array, error)
	GetRequiredMap(name string) (Map, error)

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

func (a *arguments) GetMap(name string) (Map, bool, error) {
	v, ok, err := a.get(name, TMap, false)
	if err != nil || !ok {
		return Map{}, ok, err
	}
	return v.Value().(Map), ok, nil
}
func (a *arguments) GetRequiredMap(name string) (Map, error) {
	v, _, err := a.get(name, TMap, true)
	if err != nil {
		return Map{}, err
	}
	return v.Value().(Map), nil
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

type binaryFunc func(l, r Value) Value

type binaryFuncSignature struct {
	operator    ast.OperatorKind
	left, right Type
}

var binaryFuncLookup = map[binaryFuncSignature]binaryFunc{
	//---------------
	// Math Operators
	//---------------
	{operator: ast.AdditionOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewIntValue(l + r)
	},
	{operator: ast.AdditionOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewUIntValue(l + r)
	},
	{operator: ast.AdditionOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewFloatValue(l + r)
	},
	{operator: ast.SubtractionOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewIntValue(l - r)
	},
	{operator: ast.SubtractionOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewUIntValue(l - r)
	},
	{operator: ast.SubtractionOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewFloatValue(l - r)
	},
	{operator: ast.MultiplicationOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewIntValue(l * r)
	},
	{operator: ast.MultiplicationOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewUIntValue(l * r)
	},
	{operator: ast.MultiplicationOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewFloatValue(l * r)
	},
	{operator: ast.DivisionOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewIntValue(l / r)
	},
	{operator: ast.DivisionOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewUIntValue(l / r)
	},
	{operator: ast.DivisionOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewFloatValue(l / r)
	},

	//---------------------
	// Comparison Operators
	//---------------------

	// LessThanEqualOperator

	{operator: ast.LessThanEqualOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l <= r)
	},
	{operator: ast.LessThanEqualOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(uint64(l) <= r)
	},
	{operator: ast.LessThanEqualOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) <= r)
	},
	{operator: ast.LessThanEqualOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(l <= uint64(r))
	},
	{operator: ast.LessThanEqualOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l <= r)
	},
	{operator: ast.LessThanEqualOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) <= r)
	},
	{operator: ast.LessThanEqualOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l <= float64(r))
	},
	{operator: ast.LessThanEqualOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l <= float64(r))
	},
	{operator: ast.LessThanEqualOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l <= r)
	},

	// LessThanOperator

	{operator: ast.LessThanOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l < r)
	},
	{operator: ast.LessThanOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(uint64(l) < r)
	},
	{operator: ast.LessThanOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) < r)
	},
	{operator: ast.LessThanOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(l < uint64(r))
	},
	{operator: ast.LessThanOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l < r)
	},
	{operator: ast.LessThanOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) < r)
	},
	{operator: ast.LessThanOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l < float64(r))
	},
	{operator: ast.LessThanOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l < float64(r))
	},
	{operator: ast.LessThanOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l < r)
	},

	// GreaterThanEqualOperator

	{operator: ast.GreaterThanEqualOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l >= r)
	},
	{operator: ast.GreaterThanEqualOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(uint64(l) >= r)
	},
	{operator: ast.GreaterThanEqualOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) >= r)
	},
	{operator: ast.GreaterThanEqualOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(l >= uint64(r))
	},
	{operator: ast.GreaterThanEqualOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l >= r)
	},
	{operator: ast.GreaterThanEqualOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) >= r)
	},
	{operator: ast.GreaterThanEqualOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l >= float64(r))
	},
	{operator: ast.GreaterThanEqualOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l >= float64(r))
	},
	{operator: ast.GreaterThanEqualOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l >= r)
	},

	// GreaterThanOperator

	{operator: ast.GreaterThanOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l > r)
	},
	{operator: ast.GreaterThanOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(uint64(l) > r)
	},
	{operator: ast.GreaterThanOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) > r)
	},
	{operator: ast.GreaterThanOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(l > uint64(r))
	},
	{operator: ast.GreaterThanOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l > r)
	},
	{operator: ast.GreaterThanOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) > r)
	},
	{operator: ast.GreaterThanOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l > float64(r))
	},
	{operator: ast.GreaterThanOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l > float64(r))
	},
	{operator: ast.GreaterThanOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l > r)
	},

	// EqualOperator

	{operator: ast.EqualOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l == r)
	},
	{operator: ast.EqualOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(uint64(l) == r)
	},
	{operator: ast.EqualOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) == r)
	},
	{operator: ast.EqualOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(false)
		}
		return NewBoolValue(l == uint64(r))
	},
	{operator: ast.EqualOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l == r)
	},
	{operator: ast.EqualOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) == r)
	},
	{operator: ast.EqualOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l == float64(r))
	},
	{operator: ast.EqualOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l == float64(r))
	},
	{operator: ast.EqualOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l == r)
	},
	{operator: ast.EqualOperator, left: TString, right: TString}: func(lv, rv Value) Value {
		l := lv.Value().(string)
		r := rv.Value().(string)
		return NewBoolValue(l == r)
	},

	// NotEqualOperator

	{operator: ast.NotEqualOperator, left: TInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(int64)
		return NewBoolValue(l != r)
	},
	{operator: ast.NotEqualOperator, left: TInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(uint64)
		if l < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(uint64(l) != r)
	},
	{operator: ast.NotEqualOperator, left: TInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(int64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) != r)
	},
	{operator: ast.NotEqualOperator, left: TUInt, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(int64)
		if r < 0 {
			return NewBoolValue(true)
		}
		return NewBoolValue(l != uint64(r))
	},
	{operator: ast.NotEqualOperator, left: TUInt, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(uint64)
		return NewBoolValue(l != r)
	},
	{operator: ast.NotEqualOperator, left: TUInt, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(uint64)
		r := rv.Value().(float64)
		return NewBoolValue(float64(l) != r)
	},
	{operator: ast.NotEqualOperator, left: TFloat, right: TInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(int64)
		return NewBoolValue(l != float64(r))
	},
	{operator: ast.NotEqualOperator, left: TFloat, right: TUInt}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(uint64)
		return NewBoolValue(l != float64(r))
	},
	{operator: ast.NotEqualOperator, left: TFloat, right: TFloat}: func(lv, rv Value) Value {
		l := lv.Value().(float64)
		r := rv.Value().(float64)
		return NewBoolValue(l != r)
	},
}
