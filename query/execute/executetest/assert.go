package executetest

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Assert struct{}

func (Assert) Equal(t *testing.T, exp, got interface{}) {
	t.Helper()
	if !cmp.Equal(exp, got) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(exp, got))
	}
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (Assert) PanicsWithValue(t *testing.T, exp interface{}, f PanicTestFunc) {
	funcDidPanic, panicValue := didPanic(f)
	if !funcDidPanic {
		t.Errorf("func %s should panic\n\tPanic value:\t%v", getFunctionName(f), panicValue)
	}
	if panicValue != exp {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(exp, panicValue))
	}
}

// PanicTestFunc defines a func that should be passed to the assert.Panics and assert.NotPanics
// methods, and represents a simple func that takes no arguments, and returns nothing.
type PanicTestFunc func()

// didPanic returns true if the function passed to it panics. Otherwise, it returns false.
func didPanic(f PanicTestFunc) (bool, interface{}) {

	didPanic := false
	var message interface{}
	func() {

		defer func() {
			if message = recover(); message != nil {
				didPanic = true
			}
		}()

		// call the target function
		f()

	}()

	return didPanic, message

}
