package semantic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// Type is the representation of an IFQL type.
//
// Type values are comparable and as such can be used as map keys and directly comparison using the == operator.
// Two types are equal if they represent identical types.
//
// DO NOT embed this type into other interfaces or structs as that will invalidate the comparison properties of the interface.
type Type interface {
	// Kind returns the specific kind of this type.
	Kind() Kind

	// PropertyType returns the type of a given property.
	// It panics if the type's Kind is not KObject
	PropertyType(name string) Type

	// Properties returns a map of all property types.
	// It panics if the type's Kind is not KObject
	Properties() map[string]Type

	// ElementType return the type of elements in the array.
	// It panics if the type's Kind is not KArray.
	ElementType() Type

	// ReturnType reports the return type of the function
	// It panics if the type's Kind is not KFunction.
	ReturnType() Type

	// Types cannot be created outside of the semantic package
	// This is needed so that we can cache type definitions.
	typ()
}

type Kind int

const (
	Invalid Kind = iota
	Nil
	String
	Int
	UInt
	Float
	Bool
	Time
	Duration
	Regex
	Array
	Object
	Function
)

var kindNames = []string{
	Invalid:  "invalid",
	Nil:      "nil",
	String:   "string",
	Int:      "int",
	UInt:     "uint",
	Float:    "float",
	Bool:     "bool",
	Time:     "time",
	Duration: "duration",
	Regex:    "regex",
	Array:    "array",
	Object:   "object",
	Function: "function",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind" + strconv.Itoa(int(k))
}

func (k Kind) Kind() Kind {
	return k
}
func (k Kind) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get type of property %q, from kind %q", name, k))
}
func (k Kind) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties from kind %s", k))
}
func (k Kind) ElementType() Type {
	panic(fmt.Errorf("cannot get elelemnt type from kind %s", k))
}
func (k Kind) ReturnType() Type {
	panic(fmt.Errorf("cannot get return type from kind %s", k))
}
func (k Kind) typ() {}

// ZeroExpression creates a new expression matching the provided type with zero values as literals.
func ZeroExpression(t Type) Expression {
	switch k := t.Kind(); k {
	case Bool:
		return &BooleanLiteral{}
	case UInt:
		return &UnsignedIntegerLiteral{}
	case Int:
		return &IntegerLiteral{}
	case Float:
		return &FloatLiteral{}
	case String:
		return &StringLiteral{}
	case Time:
		return &DateTimeLiteral{}
	case Array:
		return &ArrayExpression{
			Elements: []Expression{ZeroExpression(t.ElementType())},
		}
	case Object:
		props := t.Properties()
		o := &ObjectExpression{
			Properties: make([]*Property, 0, len(props)),
		}
		for k, pt := range props {
			o.Properties = append(o.Properties, &Property{
				Key:   &Identifier{Name: k},
				Value: ZeroExpression(pt),
			})
		}
		return o
	default:
		return nil
	}
}

type arrayType struct {
	elementType Type
}

func (t *arrayType) String() string {
	return fmt.Sprintf("[%v]", t.elementType)
}

func (t *arrayType) Kind() Kind {
	return Array
}
func (t *arrayType) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get property type of kind %s", t.Kind()))
}
func (t *arrayType) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties type of kind %s", t.Kind()))
}
func (t *arrayType) ElementType() Type {
	return t.elementType
}
func (t *arrayType) ReturnType() Type {
	panic(fmt.Errorf("cannot get return type of kind %s", t.Kind()))
}

func (t *arrayType) typ() {}

// arrayTypeCache caches *arrayType values.
//
// Since arrayTypes only have a single field elementType we can key
// all arrayTypes by their elementType.
var arrayTypeCache struct {
	sync.Mutex // Guards stores (but not loads) on m.

	// m is a map[Type]*arrayType keyed by the elementType of the array.
	// Elements in m are append-only and thus safe for concurrent reading.
	m sync.Map
}

// arrayTypeOf returns the Type for the given ArrayExpression.
func arrayTypeOf(e *ArrayExpression) Type {
	var et Type = Nil
	if len(e.Elements) > 0 {
		et = e.Elements[0].Type()
	}

	// Lookup arrayType in cache by elementType
	if t, ok := arrayTypeCache.m.Load(et); ok {
		return t.(*arrayType)
	}

	// Type not found in cache, lock and retry.
	arrayTypeCache.Lock()
	defer arrayTypeCache.Unlock()

	// First read again while holding the lock.
	if t, ok := arrayTypeCache.m.Load(et); ok {
		return t.(*arrayType)
	}

	// Still no cache entry, add it.
	at := &arrayType{
		elementType: et,
	}
	arrayTypeCache.m.Store(et, at)

	return at
}

type objectType struct {
	properties map[string]Type
}

func (t *objectType) String() string {
	var buf bytes.Buffer
	buf.Write([]byte("{"))
	for k, prop := range t.properties {
		fmt.Fprintf(&buf, "%s:%v,", k, prop)
	}
	buf.WriteRune('}')

	return buf.String()
}

func (t *objectType) Kind() Kind {
	return Object
}
func (t *objectType) PropertyType(name string) Type {
	return t.properties[name]
}
func (t *objectType) Properties() map[string]Type {
	return t.properties
}
func (t *objectType) ElementType() Type {
	panic(fmt.Errorf("cannot get elelemnt type of kind %s", t.Kind()))
}
func (t *objectType) ReturnType() Type {
	panic(fmt.Errorf("cannot get return type of kind %s", t.Kind()))
}
func (t *objectType) typ() {}

func (t *objectType) equal(o *objectType) bool {
	if t == o {
		return true
	}

	if len(t.properties) != len(o.properties) {
		return false
	}

	for k, vtyp := range t.properties {
		ovtyp, ok := o.properties[k]
		if !ok {
			return false
		}
		if ovtyp != vtyp {
			return false
		}
	}
	return true
}

// objectTypeCache caches all *objectTypes.
//
// Since objectTypes are identified by their properties,
// a hash is computed of the property names and kinds to reduce the search space.
var objectTypeCache struct {
	sync.Mutex // Guards stores (but not loads) on m.

	// m is a map[uint32][]*objectType keyed by the hash calculated of the object's properties' name and kind.
	// Elements in m are append-only and thus safe for concurrent reading.
	m sync.Map
}

// objectTypeOf returns the Type for the given ObjectExpression.
func objectTypeOf(e *ObjectExpression) Type {
	propertyTypes := make(map[string]Type, len(e.Properties))
	for _, p := range e.Properties {
		propertyTypes[p.Key.Name] = p.Value.Type()
	}

	return newObjectType(propertyTypes)
}

func NewObjectType(propertyTypes map[string]Type) Type {
	cpy := make(map[string]Type)
	for k, v := range propertyTypes {
		cpy[k] = v
	}
	return newObjectType(cpy)
}

func newObjectType(propertyTypes map[string]Type) Type {
	propertyNames := make([]string, 0, len(propertyTypes))
	for name := range propertyTypes {
		propertyNames = append(propertyNames, name)
	}
	sort.Strings(propertyNames)

	sum := fnv.New32a()
	for _, p := range propertyNames {
		t := propertyTypes[p]

		// track hash of property names and kinds
		sum.Write([]byte(p))
		binary.Write(sum, binary.LittleEndian, t.Kind())
	}

	// Create new object type
	ot := &objectType{
		properties: propertyTypes,
	}

	// Simple linear search after hash lookup
	h := sum.Sum32()
	if ts, ok := objectTypeCache.m.Load(h); ok {
		for _, t := range ts.([]*objectType) {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Type not found in cache, lock and retry.
	objectTypeCache.Lock()
	defer objectTypeCache.Unlock()

	// First read again while holding the lock.
	var types []*objectType
	if ts, ok := objectTypeCache.m.Load(h); ok {
		types = ts.([]*objectType)
		for _, t := range types {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Still no cache entry, add it.
	objectTypeCache.m.Store(h, append(types, ot))

	return ot
}

type functionType struct {
	params     map[string]Type
	returnType Type
}

func (t *functionType) String() string {
	var buf bytes.Buffer
	buf.Write([]byte("function("))
	for k, param := range t.params {
		fmt.Fprintf(&buf, "%s:%v,", k, param)
	}
	fmt.Fprintf(&buf, ") %v", t.returnType)

	return buf.String()
}

func (t *functionType) Kind() Kind {
	return Function
}
func (t *functionType) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get property type of kind %s", t.Kind()))
}
func (t *functionType) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties type of kind %s", t.Kind()))
}
func (t *functionType) ElementType() Type {
	panic(fmt.Errorf("cannot get elelemnt type of kind %s", t.Kind()))
}
func (t *functionType) ReturnType() Type {
	return t.returnType
}
func (t *functionType) typ() {}

func (t *functionType) equal(o *functionType) bool {
	if t == o {
		return true
	}

	if t.returnType != o.returnType {
		return false
	}

	if len(t.params) != len(o.params) {
		return false
	}

	for k, pt := range t.params {
		opt, ok := o.params[k]
		if !ok {
			return false
		}
		if opt != pt {
			return false
		}
	}

	return true
}

// functionTypeCache caches all *functionTypes.
//
// Since functionTypes are identified by their parameters and returnType,
// a hash is computed of the param names and kinds to reduce the search space.
var functionTypeCache struct {
	sync.Mutex // Guards stores (but not loads) on m.

	// m is a map[uint32][]*functionType keyed by the hash calculated.
	// Elements in m are append-only and thus safe for concurrent reading.
	m sync.Map
}

// functionTypeOf returns the Type for the given ObjectExpression.
func functionTypeOf(e *FunctionExpression) Type {
	paramNames := make([]string, 0, len(e.Params))
	paramTypes := make(map[string]Type, len(e.Params))
	for _, p := range e.Params {
		paramNames = append(paramNames, p.Key.Name)
		paramTypes[p.Key.Name] = p.Type()
	}
	sort.Strings(paramNames)

	sum := fnv.New32a()
	for _, p := range paramNames {

		// track hash of parameter names and kinds
		sum.Write([]byte(p))
		// TODO(nathanielc): Include parameter type information
		//binary.Write(sum, binary.LittleEndian, t.Kind())
	}

	// Determine returnType
	var returnType Type
	switch b := e.Body.(type) {
	case Expression:
		returnType = b.Type()
	case *BlockStatement:
		rs := b.ReturnStatement()
		returnType = rs.Argument.Type()
	}

	// Create new object type
	ot := &functionType{
		params:     paramTypes,
		returnType: returnType,
	}

	// Simple linear search after hash lookup
	h := sum.Sum32()
	if ts, ok := functionTypeCache.m.Load(h); ok {
		for _, t := range ts.([]*functionType) {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Type not found in cache, lock and retry.
	functionTypeCache.Lock()
	defer functionTypeCache.Unlock()

	// First read again while holding the lock.
	var types []*functionType
	if ts, ok := functionTypeCache.m.Load(h); ok {
		types = ts.([]*functionType)
		for _, t := range types {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Still no cache entry, add it.
	functionTypeCache.m.Store(h, append(types, ot))

	return ot
}
