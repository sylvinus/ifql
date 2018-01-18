package semantic

import "github.com/influxdata/ifql/ast"

type binarySignature struct {
	operator    ast.OperatorKind
	left, right Kind
}

var binaryTypesLookup = map[binarySignature]Kind{
	//---------------
	// Math Operators
	//---------------
	{operator: ast.AdditionOperator, left: KInt, right: KInt}:           KInt,
	{operator: ast.AdditionOperator, left: KUInt, right: KUInt}:         KUInt,
	{operator: ast.AdditionOperator, left: KFloat, right: KFloat}:       KFloat,
	{operator: ast.SubtractionOperator, left: KInt, right: KInt}:        KInt,
	{operator: ast.SubtractionOperator, left: KUInt, right: KUInt}:      KUInt,
	{operator: ast.SubtractionOperator, left: KFloat, right: KFloat}:    KFloat,
	{operator: ast.MultiplicationOperator, left: KInt, right: KInt}:     KInt,
	{operator: ast.MultiplicationOperator, left: KUInt, right: KUInt}:   KUInt,
	{operator: ast.MultiplicationOperator, left: KFloat, right: KFloat}: KFloat,
	{operator: ast.DivisionOperator, left: KInt, right: KInt}:           KInt,
	{operator: ast.DivisionOperator, left: KUInt, right: KUInt}:         KUInt,
	{operator: ast.DivisionOperator, left: KFloat, right: KFloat}:       KFloat,

	//---------------------
	// Comparison Operators
	//---------------------

	// LessThanEqualOperator

	{operator: ast.LessThanEqualOperator, left: KInt, right: KInt}:     KBool,
	{operator: ast.LessThanEqualOperator, left: KInt, right: KUInt}:    KBool,
	{operator: ast.LessThanEqualOperator, left: KInt, right: KFloat}:   KBool,
	{operator: ast.LessThanEqualOperator, left: KUInt, right: KInt}:    KBool,
	{operator: ast.LessThanEqualOperator, left: KUInt, right: KUInt}:   KBool,
	{operator: ast.LessThanEqualOperator, left: KUInt, right: KFloat}:  KBool,
	{operator: ast.LessThanEqualOperator, left: KFloat, right: KInt}:   KBool,
	{operator: ast.LessThanEqualOperator, left: KFloat, right: KUInt}:  KBool,
	{operator: ast.LessThanEqualOperator, left: KFloat, right: KFloat}: KBool,

	// LessThanOperator

	{operator: ast.LessThanOperator, left: KInt, right: KInt}:     KBool,
	{operator: ast.LessThanOperator, left: KInt, right: KUInt}:    KBool,
	{operator: ast.LessThanOperator, left: KInt, right: KFloat}:   KBool,
	{operator: ast.LessThanOperator, left: KUInt, right: KInt}:    KBool,
	{operator: ast.LessThanOperator, left: KUInt, right: KUInt}:   KBool,
	{operator: ast.LessThanOperator, left: KUInt, right: KFloat}:  KBool,
	{operator: ast.LessThanOperator, left: KFloat, right: KInt}:   KBool,
	{operator: ast.LessThanOperator, left: KFloat, right: KUInt}:  KBool,
	{operator: ast.LessThanOperator, left: KFloat, right: KFloat}: KBool,

	// GreaterThanEqualOperator

	{operator: ast.GreaterThanEqualOperator, left: KInt, right: KInt}:     KBool,
	{operator: ast.GreaterThanEqualOperator, left: KInt, right: KUInt}:    KBool,
	{operator: ast.GreaterThanEqualOperator, left: KInt, right: KFloat}:   KBool,
	{operator: ast.GreaterThanEqualOperator, left: KUInt, right: KInt}:    KBool,
	{operator: ast.GreaterThanEqualOperator, left: KUInt, right: KUInt}:   KBool,
	{operator: ast.GreaterThanEqualOperator, left: KUInt, right: KFloat}:  KBool,
	{operator: ast.GreaterThanEqualOperator, left: KFloat, right: KInt}:   KBool,
	{operator: ast.GreaterThanEqualOperator, left: KFloat, right: KUInt}:  KBool,
	{operator: ast.GreaterThanEqualOperator, left: KFloat, right: KFloat}: KBool,

	// GreaterThanOperator

	{operator: ast.GreaterThanOperator, left: KInt, right: KInt}:     KBool,
	{operator: ast.GreaterThanOperator, left: KInt, right: KUInt}:    KBool,
	{operator: ast.GreaterThanOperator, left: KInt, right: KFloat}:   KBool,
	{operator: ast.GreaterThanOperator, left: KUInt, right: KInt}:    KBool,
	{operator: ast.GreaterThanOperator, left: KUInt, right: KUInt}:   KBool,
	{operator: ast.GreaterThanOperator, left: KUInt, right: KFloat}:  KBool,
	{operator: ast.GreaterThanOperator, left: KFloat, right: KInt}:   KBool,
	{operator: ast.GreaterThanOperator, left: KFloat, right: KUInt}:  KBool,
	{operator: ast.GreaterThanOperator, left: KFloat, right: KFloat}: KBool,

	// EqualOperator

	{operator: ast.EqualOperator, left: KInt, right: KInt}:       KBool,
	{operator: ast.EqualOperator, left: KInt, right: KUInt}:      KBool,
	{operator: ast.EqualOperator, left: KInt, right: KFloat}:     KBool,
	{operator: ast.EqualOperator, left: KUInt, right: KInt}:      KBool,
	{operator: ast.EqualOperator, left: KUInt, right: KUInt}:     KBool,
	{operator: ast.EqualOperator, left: KUInt, right: KFloat}:    KBool,
	{operator: ast.EqualOperator, left: KFloat, right: KInt}:     KBool,
	{operator: ast.EqualOperator, left: KFloat, right: KUInt}:    KBool,
	{operator: ast.EqualOperator, left: KFloat, right: KFloat}:   KBool,
	{operator: ast.EqualOperator, left: KString, right: KString}: KBool,

	// NotEqualOperator

	{operator: ast.NotEqualOperator, left: KInt, right: KInt}:     KBool,
	{operator: ast.NotEqualOperator, left: KInt, right: KUInt}:    KBool,
	{operator: ast.NotEqualOperator, left: KInt, right: KFloat}:   KBool,
	{operator: ast.NotEqualOperator, left: KUInt, right: KInt}:    KBool,
	{operator: ast.NotEqualOperator, left: KUInt, right: KUInt}:   KBool,
	{operator: ast.NotEqualOperator, left: KUInt, right: KFloat}:  KBool,
	{operator: ast.NotEqualOperator, left: KFloat, right: KInt}:   KBool,
	{operator: ast.NotEqualOperator, left: KFloat, right: KUInt}:  KBool,
	{operator: ast.NotEqualOperator, left: KFloat, right: KFloat}: KBool,
}
