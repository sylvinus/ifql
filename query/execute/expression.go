package execute

import (
	"fmt"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/pkg/errors"
)

func ExpressionToStoragePredicate(e expression.Node) (*storage.Predicate, error) {
	root, err := doExpression(e)
	if err != nil {
		return nil, err
	}
	return &storage.Predicate{
		Root: root,
	}, nil
}

func doExpression(e expression.Node) (*storage.Node, error) {
	switch expr := e.(type) {
	case *expression.BinaryNode:
		left, err := doExpression(expr.Left)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := doExpression(expr.Right)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		children := []*storage.Node{left, right}
		switch expr.Operator {
		case expression.AndOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalAnd},
				Children: children,
			}, nil
		case expression.OrOperator:
			return &storage.Node{
				NodeType: storage.NodeTypeLogicalExpression,
				Value:    &storage.Node_Logical_{Logical: storage.LogicalOr},
				Children: children,
			}, nil
		}
		op, err := toComparisonOperator(expr.Operator)
		if err != nil {
			return nil, err
		}
		return &storage.Node{
			NodeType: storage.NodeTypeComparisonExpression,
			Value:    &storage.Node_Comparison_{Comparison: op},
			Children: children,
		}, nil
	case *expression.StringLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_StringValue{
				StringValue: expr.Value,
			},
		}, nil
	case *expression.IntegerLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_IntegerValue{
				IntegerValue: expr.Value,
			},
		}, nil
	case *expression.BooleanLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_BooleanValue{
				BooleanValue: expr.Value,
			},
		}, nil
	case *expression.FloatLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_FloatValue{
				FloatValue: expr.Value,
			},
		}, nil
	case *expression.RegexpLiteralNode:
		return &storage.Node{
			NodeType: storage.NodeTypeLiteral,
			Value: &storage.Node_RegexValue{
				RegexValue: expr.Value,
			},
		}, nil
	case *expression.ReferenceNode:
		switch expr.Kind {
		case "tag":
			return &storage.Node{
				NodeType: storage.NodeTypeTagRef,
				Value: &storage.Node_TagRefValue{
					TagRefValue: expr.Name,
				},
			}, nil
		case "field":
			return &storage.Node{
				NodeType: storage.NodeTypeFieldRef,
				Value: &storage.Node_FieldRefValue{
					FieldRefValue: "_field",
				},
			}, nil
		default:
			return nil, fmt.Errorf("unsupported reference kind %q", expr.Kind)
		}
	case *expression.DurationLiteralNode:
		return nil, errors.New("duration literals not supported in storage predicates")
	case *expression.TimeLiteralNode:
		return nil, errors.New("time literals not supported in storage predicates")
	default:
		return nil, fmt.Errorf("unsupported expression type %v", e.Type())
	}
}

func toComparisonOperator(o expression.Operator) (storage.Node_Comparison, error) {
	switch o {
	case expression.EqualOperator:
		return storage.ComparisonEqual, nil
	case expression.NotOperator:
		return storage.ComparisonNotEqual, nil
	case expression.StartsWithOperator:
		return storage.ComparisonStartsWith, nil
	case expression.RegexpMatchOperator:
		return storage.ComparisonRegex, nil
	case expression.RegexpNotMatchOperator:
		return storage.ComparisonNotRegex, nil
	case expression.LessThanOperator:
		return storage.ComparisonLess, nil
	case expression.LessThanEqualOperator:
		return storage.ComparisonLessEqual, nil
	case expression.GreaterThanOperator:
		return storage.ComparisonGreater, nil
	case expression.GreaterThanEqualOperator:
		return storage.ComparisonGreaterEqual, nil
	default:
		return 0, fmt.Errorf("unknown expression operator %v", o)
	}
}

func ExpressionNames(n expression.Node) []string {
	var names []string
	expression.Walk(n, func(n expression.Node) error {
		if rn, ok := n.(*expression.ReferenceNode); ok {
			found := false
			for _, n := range names {
				if n == rn.Name {
					found = true
					break
				}
			}
			if !found {
				names = append(names, rn.Name)
			}
		}
		return nil
	})
	return names
}

func CompileExpression(e expression.Expression, types map[string]DataType) (CompiledExpression, error) {
	// Validate we have types for each name
	names := ExpressionNames(e.Root)
	for _, n := range names {
		if types[n] == TInvalid {
			return nil, fmt.Errorf("missing type information for name %q", n)
		}
	}

	root, err := compile(e.Root, types)
	if err != nil {
		return nil, err
	}
	return compiledExpression{
		root:  root,
		types: types,
	}, nil
}

type compiledExpression struct {
	root  DataTypeEvaluator
	types map[string]DataType
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

func compile(n expression.Node, types map[string]DataType) (DataTypeEvaluator, error) {
	switch n := n.(type) {
	case *expression.ReferenceNode:
		return &referenceEvaluator{
			compiledType: compiledType(types[n.Name]),
			name:         n.Name,
		}, nil
	case *expression.BooleanLiteralNode:
		return &booleanEvaluator{
			compiledType: compiledType(TBool),
			b:            n.Value,
		}, nil
	case *expression.IntegerLiteralNode:
		return &integerEvaluator{
			compiledType: compiledType(TInt),
			i:            n.Value,
		}, nil
	case *expression.FloatLiteralNode:
		return &floatEvaluator{
			compiledType: compiledType(TFloat),
			f:            n.Value,
		}, nil
	case *expression.StringLiteralNode:
		return &stringEvaluator{
			compiledType: compiledType(TString),
			s:            n.Value,
		}, nil
	case *expression.TimeLiteralNode:
		return &timeEvaluator{
			compiledType: compiledType(TTime),
			t:            Time(n.Value.UnixNano()),
		}, nil
	case *expression.UnaryNode:
		node, err := compile(n.Node, types)
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
	case *expression.BinaryNode:
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
			return nil, fmt.Errorf("unsupported binary expression with types %v %v %v", sig.Left, sig.Operator, sig.Right)
		}
		return &binaryEvaluator{
			compiledType: compiledType(f.ResultType),
			left:         l,
			right:        r,
			f:            f.Func,
		}, nil
	default:
		return nil, fmt.Errorf("unknown expression node of type %T", n)
	}
}

type Scope map[string]Value

func (s Scope) Type(name string) DataType {
	return s[name].Type
}
func (s Scope) GetBool(name string) bool {
	return s[name].Bool()
}
func (s Scope) GetInt(name string) int64 {
	return s[name].Int()
}
func (s Scope) GetUInt(name string) uint64 {
	return s[name].UInt()
}
func (s Scope) GetFloat(name string) float64 {
	return s[name].Float()
}
func (s Scope) GetString(name string) string {
	return s[name].Str()
}
func (s Scope) GetTime(name string) Time {
	return s[name].Time()
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

type binaryFunc func(scope Scope, left, right DataTypeEvaluator) Value

type binarySignature struct {
	Operator    expression.Operator
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

type referenceEvaluator struct {
	compiledType
	name string
}

func (e *referenceEvaluator) EvalBool(scope Scope) bool {
	return scope.GetBool(e.name)
}

func (e *referenceEvaluator) EvalInt(scope Scope) int64 {
	return scope.GetInt(e.name)
}

func (e *referenceEvaluator) EvalUInt(scope Scope) uint64 {
	return scope.GetUInt(e.name)
}

func (e *referenceEvaluator) EvalFloat(scope Scope) float64 {
	return scope.GetFloat(e.name)
}

func (e *referenceEvaluator) EvalString(scope Scope) string {
	return scope.GetString(e.name)
}

func (e *referenceEvaluator) EvalTime(scope Scope) Time {
	return scope.GetTime(e.name)
}

// Map of binary functions
var binaryFuncs = map[binarySignature]struct {
	Func       binaryFunc
	ResultType DataType
}{
	//---------------
	// Math Operators
	//---------------
	{Operator: expression.AdditionOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.AdditionOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.AdditionOperator, Left: TFloat, Right: TFloat}: {
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
	{Operator: expression.SubtractionOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.SubtractionOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.SubtractionOperator, Left: TFloat, Right: TFloat}: {
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
	{Operator: expression.MultiplicationOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.MultiplicationOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.MultiplicationOperator, Left: TFloat, Right: TFloat}: {
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
	{Operator: expression.DivisionOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.DivisionOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.DivisionOperator, Left: TFloat, Right: TFloat}: {
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

	{Operator: expression.LessThanEqualOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.LessThanEqualOperator, Left: TFloat, Right: TFloat}: {
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

	{Operator: expression.LessThanOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.LessThanOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.LessThanOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.LessThanOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.LessThanOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.LessThanOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.LessThanOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.LessThanOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.LessThanOperator, Left: TFloat, Right: TFloat}: {
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

	{Operator: expression.GreaterThanEqualOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.GreaterThanEqualOperator, Left: TFloat, Right: TFloat}: {
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

	{Operator: expression.GreaterThanOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.GreaterThanOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.GreaterThanOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.GreaterThanOperator, Left: TFloat, Right: TFloat}: {
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

	{Operator: expression.EqualOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.EqualOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.EqualOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.EqualOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.EqualOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.EqualOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.EqualOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.EqualOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.EqualOperator, Left: TFloat, Right: TFloat}: {
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

	// NotEqualOperator

	{Operator: expression.NotEqualOperator, Left: TInt, Right: TInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TInt, Right: TUInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TInt, Right: TFloat}: {
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
	{Operator: expression.NotEqualOperator, Left: TUInt, Right: TInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TUInt, Right: TUInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TUInt, Right: TFloat}: {
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
	{Operator: expression.NotEqualOperator, Left: TFloat, Right: TInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TFloat, Right: TUInt}: {
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
	{Operator: expression.NotEqualOperator, Left: TFloat, Right: TFloat}: {
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

	//------------------
	// Logical Operators
	//------------------

	{Operator: expression.AndOperator, Left: TBool, Right: TBool}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalBool(scope)
			r := right.EvalBool(scope)
			return Value{
				Type:  TBool,
				Value: l && r,
			}
		},
		ResultType: TBool,
	},
	{Operator: expression.OrOperator, Left: TBool, Right: TBool}: {
		Func: func(scope Scope, left, right DataTypeEvaluator) Value {
			l := left.EvalBool(scope)
			r := right.EvalBool(scope)
			return Value{
				Type:  TBool,
				Value: l || r,
			}
		},
		ResultType: TBool,
	},
}
