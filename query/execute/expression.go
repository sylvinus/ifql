package execute

import (
	"fmt"

	"github.com/influxdata/ifql/ast"
)

// ObjectProperties returns all object/property pairs referenced in the expression.
func ObjectProperties(f *ast.ArrowFunctionExpression) ([]ObjectProperty, error) {
	return objectProperties(f.Body.(ast.Expression))
}

func objectProperties(n ast.Expression) ([]ObjectProperty, error) {
	switch n := n.(type) {
	case *ast.MemberExpression:
		op, err := objectProperty(n)
		if err != nil {
			return nil, err
		}
		return []ObjectProperty{op}, nil
	case *ast.Identifier:
		return nil, fmt.Errorf("unexpected identifier %q", n.Name)
	case *ast.UnaryExpression:
		return objectProperties(n.Argument)
	case *ast.LogicalExpression:
		l, err := objectProperties(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := objectProperties(n.Right)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	case *ast.BinaryExpression:
		l, err := objectProperties(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := objectProperties(n.Right)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	default:
		return nil, nil
	}
}

func objectProperty(m *ast.MemberExpression) (ObjectProperty, error) {
	p, err := propertyName(m)
	if err != nil {
		return ObjectProperty{}, err
	}
	if ident, ok := m.Object.(*ast.Identifier); ok {
		return ObjectProperty{
			Object:   ident.Name,
			Property: p,
		}, nil
	}
	return ObjectProperty{}, fmt.Errorf("unsupported member object expression type %T", m.Object)
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

func CompileExpression(f *ast.ArrowFunctionExpression, types map[ObjectProperty]DataType) (CompiledExpression, error) {
	objProperties, err := ObjectProperties(f)
	if err != nil {
		return nil, err
	}
	// Validate we have types for each object property
	for _, op := range objProperties {
		if types[op] == TInvalid {
			return nil, fmt.Errorf("missing type information for %q", op)
		}
	}

	root, err := compile(f.Body.(ast.Expression), types)
	if err != nil {
		return nil, err
	}
	cpy := make(map[ObjectProperty]DataType, len(types))
	for k, v := range types {
		cpy[k] = v
	}
	return compiledExpression{
		root:  root,
		types: cpy,
	}, nil
}

type ObjectProperty struct {
	Object   string
	Property string
}

func (op ObjectProperty) String() string {
	return fmt.Sprintf("%s.%s", op.Object, op.Property)
}

type compiledExpression struct {
	root  DataTypeEvaluator
	types map[ObjectProperty]DataType
}

func (c compiledExpression) validate(scope Scope) error {
	// Validate scope
	for k, t := range c.types {
		if scope.Type(k) != t {
			return fmt.Errorf("missing or incorrectly typed value found in scope for name %q", k)
		}
	}
	return nil
}

func (c compiledExpression) Type() DataType {
	return c.root.Type()
}
func (c compiledExpression) Eval(scope Scope) (Value, error) {
	if err := c.validate(scope); err != nil {
		return Value{}, err
	}
	var val interface{}
	switch c.Type() {
	case TBool:
		val = c.root.EvalBool(scope)
	case TInt:
		val = c.root.EvalInt(scope)
	case TUInt:
		val = c.root.EvalUInt(scope)
	case TFloat:
		val = c.root.EvalFloat(scope)
	case TString:
		val = c.root.EvalString(scope)
	case TTime:
		val = c.root.EvalTime(scope)
	default:
		return Value{}, fmt.Errorf("unsupported type %s", c.Type())
	}
	return Value{
		Type:  c.Type(),
		Value: val,
	}, nil
}

func (c compiledExpression) EvalBool(scope Scope) (bool, error) {
	if err := c.validate(scope); err != nil {
		return false, err
	}
	return c.root.EvalBool(scope), nil
}
func (c compiledExpression) EvalInt(scope Scope) (int64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalInt(scope), nil
}
func (c compiledExpression) EvalUInt(scope Scope) (uint64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalUInt(scope), nil
}
func (c compiledExpression) EvalFloat(scope Scope) (float64, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalFloat(scope), nil
}
func (c compiledExpression) EvalString(scope Scope) (string, error) {
	if err := c.validate(scope); err != nil {
		return "", err
	}
	return c.root.EvalString(scope), nil
}
func (c compiledExpression) EvalTime(scope Scope) (Time, error) {
	if err := c.validate(scope); err != nil {
		return 0, err
	}
	return c.root.EvalTime(scope), nil
}

func compile(n ast.Expression, types map[ObjectProperty]DataType) (DataTypeEvaluator, error) {
	switch n := n.(type) {
	case *ast.MemberExpression:
		op, err := objectProperty(n)
		if err != nil {
			return nil, err
		}
		return &objectPropertyEvaluator{
			compiledType: compiledType(types[op]),
			op:           op,
		}, nil
	case *ast.BooleanLiteral:
		return &booleanEvaluator{
			compiledType: compiledType(TBool),
			b:            n.Value,
		}, nil
	case *ast.IntegerLiteral:
		return &integerEvaluator{
			compiledType: compiledType(TInt),
			i:            n.Value,
		}, nil
	case *ast.NumberLiteral:
		return &floatEvaluator{
			compiledType: compiledType(TFloat),
			f:            n.Value,
		}, nil
	case *ast.StringLiteral:
		return &stringEvaluator{
			compiledType: compiledType(TString),
			s:            n.Value,
		}, nil
	case *ast.DateTimeLiteral:
		return &timeEvaluator{
			compiledType: compiledType(TTime),
			t:            Time(n.Value.UnixNano()),
		}, nil
	case *ast.UnaryExpression:
		node, err := compile(n.Argument, types)
		if err != nil {
			return nil, err
		}
		nt := node.Type()
		if nt != TBool && nt != TInt && nt != TFloat {
			return nil, fmt.Errorf("invalid unary operator %v on type %v", n.Operator, nt)
		}
		return &unaryEvaluator{
			compiledType: compiledType(nt),
			node:         node,
		}, nil
	case *ast.LogicalExpression:
		l, err := compile(n.Left, types)
		if err != nil {
			return nil, err
		}
		if l.Type() != TBool {
			return nil, fmt.Errorf("invalid left operand type %v in logical ast", l.Type())
		}
		r, err := compile(n.Right, types)
		if err != nil {
			return nil, err
		}
		if r.Type() != TBool {
			return nil, fmt.Errorf("invalid right operand type %v in logical ast", r.Type())
		}
		return &logicalEvaluator{
			compiledType: compiledType(TBool),
			operator:     n.Operator,
			left:         l,
			right:        r,
		}, nil
	case *ast.BinaryExpression:
		l, err := compile(n.Left, types)
		if err != nil {
			return nil, err
		}
		lt := l.Type()
		r, err := compile(n.Right, types)
		if err != nil {
			return nil, err
		}
		rt := l.Type()
		sig := binarySignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
		}
		f, ok := binaryFuncs[sig]
		if !ok {
			return nil, fmt.Errorf("unsupported binary ast with types %v %v %v", sig.Left, sig.Operator, sig.Right)
		}
		return &binaryEvaluator{
			compiledType: compiledType(f.ResultType),
			left:         l,
			right:        r,
			f:            f.Func,
		}, nil
	default:
		return nil, fmt.Errorf("unknown ast node of type %T", n)
	}
}

type Scope map[ObjectProperty]Value

func (s Scope) Type(op ObjectProperty) DataType {
	return s[op].Type
}
func (s Scope) GetBool(op ObjectProperty) bool {
	return s[op].Bool()
}
func (s Scope) GetInt(op ObjectProperty) int64 {
	return s[op].Int()
}
func (s Scope) GetUInt(op ObjectProperty) uint64 {
	return s[op].UInt()
}
func (s Scope) GetFloat(op ObjectProperty) float64 {
	return s[op].Float()
}
func (s Scope) GetString(op ObjectProperty) string {
	return s[op].Str()
}
func (s Scope) GetTime(op ObjectProperty) Time {
	return s[op].Time()
}

type DataTypeEvaluator interface {
	Type() DataType
	EvalBool(scope Scope) bool
	EvalInt(scope Scope) int64
	EvalUInt(scope Scope) uint64
	EvalFloat(scope Scope) float64
	EvalString(scope Scope) string
	EvalTime(scope Scope) Time
}
type CompiledExpression interface {
	Type() DataType
	Eval(scope Scope) (Value, error)
	EvalBool(scope Scope) (bool, error)
	EvalInt(scope Scope) (int64, error)
	EvalUInt(scope Scope) (uint64, error)
	EvalFloat(scope Scope) (float64, error)
	EvalString(scope Scope) (string, error)
	EvalTime(scope Scope) (Time, error)
}

type compiledType DataType

func (c compiledType) Type() DataType {
	return DataType(c)
}
func (c compiledType) error(exp DataType) error {
	return typeErr{Actual: DataType(c), Expected: exp}
}

type typeErr struct {
	Actual, Expected DataType
}

func (t typeErr) Error() string {
	return fmt.Sprintf("unexpected type: got %q want %q", t.Actual, t.Expected)
}

type Value struct {
	Type  DataType
	Value interface{}
}

func (v Value) error(exp DataType) error {
	return typeErr{Actual: v.Type, Expected: exp}
}

func (v Value) Bool() bool {
	return v.Value.(bool)
}
func (v Value) Int() int64 {
	return v.Value.(int64)
}
func (v Value) UInt() uint64 {
	return v.Value.(uint64)
}
func (v Value) Float() float64 {
	return v.Value.(float64)
}
func (v Value) Str() string {
	return v.Value.(string)
}
func (v Value) Time() Time {
	return v.Value.(Time)
}

type logicalEvaluator struct {
	compiledType
	operator    ast.LogicalOperatorKind
	left, right DataTypeEvaluator
}

func (e *logicalEvaluator) EvalBool(scope Scope) bool {
	switch e.operator {
	case ast.AndOperator:
		return e.left.EvalBool(scope) && e.right.EvalBool(scope)
	case ast.OrOperator:
		return e.left.EvalBool(scope) || e.right.EvalBool(scope)
	default:
		panic(fmt.Errorf("unknown logical operator %v", e.operator))
	}
}

func (e *logicalEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *logicalEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *logicalEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *logicalEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *logicalEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type binaryFunc func(scope Scope, left, right DataTypeEvaluator) Value

type binarySignature struct {
	Operator    ast.OperatorKind
	Left, Right DataType
}

type binaryEvaluator struct {
	compiledType
	left, right DataTypeEvaluator
	f           binaryFunc
}

func (e *binaryEvaluator) EvalBool(scope Scope) bool {
	return e.f(scope, e.left, e.right).Bool()
}

func (e *binaryEvaluator) EvalInt(scope Scope) int64 {
	return e.f(scope, e.left, e.right).Int()
}

func (e *binaryEvaluator) EvalUInt(scope Scope) uint64 {
	return e.f(scope, e.left, e.right).UInt()
}

func (e *binaryEvaluator) EvalFloat(scope Scope) float64 {
	return e.f(scope, e.left, e.right).Float()
}

func (e *binaryEvaluator) EvalString(scope Scope) string {
	return e.f(scope, e.left, e.right).Str()
}

func (e *binaryEvaluator) EvalTime(scope Scope) Time {
	return e.f(scope, e.left, e.right).Time()
}

type unaryEvaluator struct {
	compiledType
	node DataTypeEvaluator
}

func (e *unaryEvaluator) EvalBool(scope Scope) bool {
	// There is only one boolean unary operator
	return !e.node.EvalBool(scope)
}

func (e *unaryEvaluator) EvalInt(scope Scope) int64 {
	// There is only one integer unary operator
	return -e.node.EvalInt(scope)
}

func (e *unaryEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *unaryEvaluator) EvalFloat(scope Scope) float64 {
	// There is only one float unary operator
	return -e.node.EvalFloat(scope)
}

func (e *unaryEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *unaryEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type integerEvaluator struct {
	compiledType
	i int64
}

func (e *integerEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *integerEvaluator) EvalInt(scope Scope) int64 {
	return e.i
}

func (e *integerEvaluator) EvalUInt(scope Scope) uint64 {
	return uint64(e.i)
}

func (e *integerEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *integerEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *integerEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type stringEvaluator struct {
	compiledType
	s string
}

func (e *stringEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *stringEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *stringEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *stringEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *stringEvaluator) EvalString(scope Scope) string {
	return e.s
}

func (e *stringEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type booleanEvaluator struct {
	compiledType
	b bool
}

func (e *booleanEvaluator) EvalBool(scope Scope) bool {
	return e.b
}

func (e *booleanEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *booleanEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *booleanEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *booleanEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *booleanEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type floatEvaluator struct {
	compiledType
	f float64
}

func (e *floatEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *floatEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *floatEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *floatEvaluator) EvalFloat(scope Scope) float64 {
	return e.f
}

func (e *floatEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *floatEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}

type timeEvaluator struct {
	compiledType
	t Time
}

func (e *timeEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *timeEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *timeEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *timeEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *timeEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *timeEvaluator) EvalTime(scope Scope) Time {
	return e.t
}

type objectPropertyEvaluator struct {
	compiledType
	op ObjectProperty
}

func (e *objectPropertyEvaluator) EvalBool(scope Scope) bool {
	return scope.GetBool(e.op)
}

func (e *objectPropertyEvaluator) EvalInt(scope Scope) int64 {
	return scope.GetInt(e.op)
}

func (e *objectPropertyEvaluator) EvalUInt(scope Scope) uint64 {
	return scope.GetUInt(e.op)
}

func (e *objectPropertyEvaluator) EvalFloat(scope Scope) float64 {
	return scope.GetFloat(e.op)
}

func (e *objectPropertyEvaluator) EvalString(scope Scope) string {
	return scope.GetString(e.op)
}

func (e *objectPropertyEvaluator) EvalTime(scope Scope) Time {
	return scope.GetTime(e.op)
}

// Map of binary functions
var binaryFuncs = map[binarySignature]struct {
	Func       binaryFunc
	ResultType DataType
}{
	//---------------
	// Math Operators
	//---------------
	{Operator: ast.AdditionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l + r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.AdditionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l + r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.AdditionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l + r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.SubtractionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l - r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.SubtractionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l - r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.SubtractionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l - r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.MultiplicationOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l * r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.MultiplicationOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l * r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.MultiplicationOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l * r,
			}
		},
		ResultType: TFloat,
	},
	{Operator: ast.DivisionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TInt,
				Value: l / r,
			}
		},
		ResultType: TInt,
	},
	{Operator: ast.DivisionOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TUInt,
				Value: l / r,
			}
		},
		ResultType: TUInt,
	},
	{Operator: ast.DivisionOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TFloat,
				Value: l / r,
			}
		},
		ResultType: TFloat,
	},

	//---------------------
	// Comparison Operators
	//---------------------

	// LessThanEqualOperator

	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l <= uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) <= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l <= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l <= r,
			}
		},
		ResultType: TBool,
	},

	// LessThanOperator

	{Operator: ast.LessThanOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l < uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) < r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l < float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l < float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.LessThanOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l < r,
			}
		},
		ResultType: TBool,
	},

	// GreaterThanEqualOperator

	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l >= uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) >= r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l >= float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l >= r,
			}
		},
		ResultType: TBool,
	},

	// GreaterThanOperator

	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l > uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) > r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l > float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l > float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.GreaterThanOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l > r,
			}
		},
		ResultType: TBool,
	},

	// EqualOperator

	{Operator: ast.EqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: false,
				}
			}
			return Value{
				Type:  TBool,
				Value: l == uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l == float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l == float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.EqualOperator, Left: TString, Right: TString}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalString(scope)
			r := right.EvalString(scope)
			return Value{
				Type:  TBool,
				Value: l == r,
			}
		},
		ResultType: TBool,
	},

	// NotEqualOperator

	{Operator: ast.NotEqualOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalUInt(scope)
			if l < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: uint64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalInt(scope)
			if r < 0 {
				return Value{
					Type:  TBool,
					Value: true,
				}
			}
			return Value{
				Type:  TBool,
				Value: l != uint64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TUInt, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalUInt(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: float64(l) != r,
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalInt(scope)
			return Value{
				Type:  TBool,
				Value: l != float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TUInt}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalUInt(scope)
			return Value{
				Type:  TBool,
				Value: l != float64(r),
			}
		},
		ResultType: TBool,
	},
	{Operator: ast.NotEqualOperator, Left: TFloat, Right: TFloat}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalFloat(scope)
			r := right.EvalFloat(scope)
			return Value{
				Type:  TBool,
				Value: l != r,
			}
		},
		ResultType: TBool,
	},
}
