package compiler

import (
	"fmt"
	"strings"

	"github.com/influxdata/ifql/ast"
)

type Evaluator interface {
	Type() Type
	MapMeta() MapMeta
	EvalBool(scope Scope) bool
	EvalInt(scope Scope) int64
	EvalUInt(scope Scope) uint64
	EvalFloat(scope Scope) float64
	EvalString(scope Scope) string
	EvalTime(scope Scope) Time
	EvalMap(scope Scope) Map
}

type Func interface {
	Type() Type
	MapMeta() MapMeta
	Eval(scope Scope) (Value, error)
	EvalBool(scope Scope) (bool, error)
	EvalInt(scope Scope) (int64, error)
	EvalUInt(scope Scope) (uint64, error)
	EvalFloat(scope Scope) (float64, error)
	EvalString(scope Scope) (string, error)
	EvalTime(scope Scope) (Time, error)
	EvalMap(scope Scope) (Map, error)
}

type Value struct {
	Type  Type
	Value interface{}
}

func (v Value) error(exp Type) error {
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
func (v Value) Map() Map {
	return v.Value.(Map)
}

type Reference []string

type ReferencePath string

func (r Reference) Path() ReferencePath {
	return ReferencePath(strings.Join([]string(r), "."))
}

func (r Reference) String() string {
	return string(r.Path())
}

type typeErr struct {
	Actual, Expected Type
}

func (t typeErr) Error() string {
	return fmt.Sprintf("unexpected type: got %q want %q", t.Actual, t.Expected)
}

type Scope map[ReferencePath]Value

func (s Scope) Type(rp ReferencePath) Type {
	return s[rp].Type
}
func (s Scope) Set(rp ReferencePath, v Value) {
	s[rp] = v
}
func (s Scope) GetBool(rp ReferencePath) bool {
	return s[rp].Bool()
}
func (s Scope) GetInt(rp ReferencePath) int64 {
	return s[rp].Int()
}
func (s Scope) GetUInt(rp ReferencePath) uint64 {
	return s[rp].UInt()
}
func (s Scope) GetFloat(rp ReferencePath) float64 {
	return s[rp].Float()
}
func (s Scope) GetString(rp ReferencePath) string {
	return s[rp].Str()
}
func (s Scope) GetTime(rp ReferencePath) Time {
	return s[rp].Time()
}
func (s Scope) GetMap(rp ReferencePath) Map {
	return s[rp].Map()
}

type MapMeta struct {
	Properties []MapPropertyMeta
}

type MapPropertyMeta struct {
	Key  string
	Type Type
}

func eval(e Evaluator, scope Scope) (v Value) {
	v.Type = e.Type()
	switch v.Type {
	case TBool:
		v.Value = e.EvalBool(scope)
	case TInt:
		v.Value = e.EvalInt(scope)
	case TUInt:
		v.Value = e.EvalUInt(scope)
	case TFloat:
		v.Value = e.EvalFloat(scope)
	case TString:
		v.Value = e.EvalString(scope)
	case TTime:
		v.Value = e.EvalTime(scope)
	}
	return
}

type blockEvaluator struct {
	compiledType
	body  []Evaluator
	value Value
}

func (e *blockEvaluator) eval(scope Scope) {
	for _, b := range e.body {
		e.value = eval(b, scope)
	}
}

func (e *blockEvaluator) EvalBool(scope Scope) bool {
	if Type(e.compiledType) != TBool {
		panic(e.error(TBool))
	}
	e.eval(scope)
	return e.value.Bool()
}

func (e *blockEvaluator) EvalInt(scope Scope) int64 {
	if Type(e.compiledType) != TInt {
		panic(e.error(TInt))
	}
	e.eval(scope)
	return e.value.Int()
}

func (e *blockEvaluator) EvalUInt(scope Scope) uint64 {
	if Type(e.compiledType) != TUInt {
		panic(e.error(TUInt))
	}
	e.eval(scope)
	return e.value.UInt()
}

func (e *blockEvaluator) EvalFloat(scope Scope) float64 {
	if Type(e.compiledType) != TFloat {
		panic(e.error(TFloat))
	}
	e.eval(scope)
	return e.value.Float()
}

func (e *blockEvaluator) EvalString(scope Scope) string {
	if Type(e.compiledType) != TString {
		panic(e.error(TString))
	}
	e.eval(scope)
	return e.value.Str()
}

func (e *blockEvaluator) EvalTime(scope Scope) Time {
	if Type(e.compiledType) != TTime {
		panic(e.error(TTime))
	}
	e.eval(scope)
	return e.value.Time()
}
func (e *blockEvaluator) EvalMap(scope Scope) Map {
	if Type(e.compiledType) != TMap {
		panic(e.error(TMap))
	}
	e.eval(scope)
	return e.value.Map()
}

type returnEvaluator struct {
	Evaluator
}
type declarationEvaluator struct {
	compiledType
	id   ReferencePath
	init Evaluator
}

func (e *declarationEvaluator) eval(scope Scope) {
	scope.Set(e.id, eval(e.init, scope))
}

func (e *declarationEvaluator) EvalBool(scope Scope) bool {
	e.eval(scope)
	return scope.GetBool(e.id)
}

func (e *declarationEvaluator) EvalInt(scope Scope) int64 {
	e.eval(scope)
	return scope.GetInt(e.id)
}

func (e *declarationEvaluator) EvalUInt(scope Scope) uint64 {
	e.eval(scope)
	return scope.GetUInt(e.id)
}

func (e *declarationEvaluator) EvalFloat(scope Scope) float64 {
	e.eval(scope)
	return scope.GetFloat(e.id)
}

func (e *declarationEvaluator) EvalString(scope Scope) string {
	e.eval(scope)
	return scope.GetString(e.id)
}

func (e *declarationEvaluator) EvalTime(scope Scope) Time {
	e.eval(scope)
	return scope.GetTime(e.id)
}

func (e *declarationEvaluator) EvalMap(scope Scope) Map {
	e.eval(scope)
	return scope.GetMap(e.id)
}

type mapEvaluator struct {
	compiledType
	meta       MapMeta
	properties map[string]Evaluator
}

func (e *mapEvaluator) MapMeta() MapMeta {
	return e.meta
}

func (e *mapEvaluator) EvalBool(scope Scope) bool {
	panic(e.error(TBool))
}

func (e *mapEvaluator) EvalInt(scope Scope) int64 {
	panic(e.error(TInt))
}

func (e *mapEvaluator) EvalUInt(scope Scope) uint64 {
	panic(e.error(TUInt))
}

func (e *mapEvaluator) EvalFloat(scope Scope) float64 {
	panic(e.error(TFloat))
}

func (e *mapEvaluator) EvalString(scope Scope) string {
	panic(e.error(TString))
}

func (e *mapEvaluator) EvalTime(scope Scope) Time {
	panic(e.error(TTime))
}
func (e *mapEvaluator) EvalMap(scope Scope) Map {
	values := make(map[string]Value, len(e.properties))
	for k, node := range e.properties {
		values[k] = eval(node, scope)
	}
	return Map{
		Meta:   e.meta,
		Values: values,
	}
}

type Map struct {
	Meta   MapMeta
	Values map[string]Value
}

type logicalEvaluator struct {
	compiledType
	operator    ast.LogicalOperatorKind
	left, right Evaluator
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
func (e *logicalEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type binaryFunc func(scope Scope, left, right Evaluator) Value

type binarySignature struct {
	Operator    ast.OperatorKind
	Left, Right Type
}

type binaryEvaluator struct {
	compiledType
	left, right Evaluator
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
func (e *binaryEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type unaryEvaluator struct {
	compiledType
	node Evaluator
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
func (e *unaryEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
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
func (e *integerEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
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
func (e *stringEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
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
func (e *booleanEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
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
func (e *floatEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
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
func (e *timeEvaluator) EvalMap(scope Scope) Map {
	panic(e.error(TMap))
}

type referenceEvaluator struct {
	compiledType
	referencePath ReferencePath
}

func (e *referenceEvaluator) EvalBool(scope Scope) bool {
	return scope.GetBool(e.referencePath)
}

func (e *referenceEvaluator) EvalInt(scope Scope) int64 {
	return scope.GetInt(e.referencePath)
}

func (e *referenceEvaluator) EvalUInt(scope Scope) uint64 {
	return scope.GetUInt(e.referencePath)
}

func (e *referenceEvaluator) EvalFloat(scope Scope) float64 {
	return scope.GetFloat(e.referencePath)
}

func (e *referenceEvaluator) EvalString(scope Scope) string {
	return scope.GetString(e.referencePath)
}

func (e *referenceEvaluator) EvalTime(scope Scope) Time {
	return scope.GetTime(e.referencePath)
}
func (e *referenceEvaluator) EvalMap(scope Scope) Map {
	return scope.GetMap(e.referencePath)
}

// Map of binary functions
var binaryFuncs = map[binarySignature]struct {
	Func       binaryFunc
	ResultType Type
}{
	//---------------
	// Math Operators
	//---------------
	{Operator: ast.AdditionOperator, Left: TInt, Right: TInt}: {
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
		Func: func(scope Scope, left, right Evaluator) Value {
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
