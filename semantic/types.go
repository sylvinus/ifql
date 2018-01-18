package semantic

import (
	"fmt"
	"strconv"
)

type Kind int

const (
	KInvalid  Kind = iota // Go type nil
	KString               // Go type string
	KInt                  // Go type int64
	KUInt                 // Go type uint64
	KFloat                // Go type float64
	KBool                 // Go type bool
	KTime                 // Go type time.Time
	KDuration             // Go type time.Duration
	KFunction             // Go type Function
	KArray                // Go type Array
	KMap                  // Go type Map
	KRegex                // Go type *regexp.Regexp
	KStruct               // Go type unknown
)

var kindNames = []string{
	KInvalid:  "invalid",
	KString:   "string",
	KInt:      "int",
	KUInt:     "uint",
	KFloat:    "float",
	KBool:     "bool",
	KTime:     "time",
	KDuration: "duration",
	KFunction: "function",
	KArray:    "array",
	KMap:      "map",
	KRegex:    "regex",
	KStruct:   "struct",
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
	panic(fmt.Errorf("cannot get property of kind %s", k))
}
func (k Kind) ElementType() Type {
	panic(fmt.Errorf("kind %s has no elements", k))
}

type Type interface {
	// Kind returns the specific kind of this type.
	Kind() Kind

	// PropertyType returns the type of a given property.
	// It panics if the type's Kind is not KMap or KStruct.
	PropertyType(name string) Type

	// ElementType return the type of elements in the array.
	// It panics if the type's Kind is not KArray.
	ElementType() Type
}
