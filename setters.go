package reflectutils

import (
	"reflect"
	"time"
)

func init() {
	RegisterStringSetter(time.ParseDuration)
}

var settersByType = make(map[reflect.Type]reflect.Value)

// RegisterStringSetter registers functions that can be used to transform
// strings into specific types.  The fn argument must be a function that
// takes a string and returns an arbitrary type and an error.  An example
// of such a function is time.ParseDuration.  Any call to RegisterStringSetter
// with a value that is not a function of that sort will panic.
//
// RegisterStringSetter is not thread safe and should probably only be
// used during init().
//
// These functions are used by MakeStringSetter() when there is an opportunity
// to do so.
func RegisterStringSetter(fn interface{}) {
	v := reflect.ValueOf(fn)
	if !v.IsValid() {
		panic("call to RegisterStringSetter with an invalid value")
	}
	if v.Type().Kind() != reflect.Func {
		panic("call to RegisterStringSetter with something other than a function")
	}
	if v.Type().NumIn() != 1 {
		panic("call to RegisterStringSetter with something other than a function that takes one arg")
	}
	if v.Type().NumOut() != 2 {
		panic("call to RegisterStringSetter with something other than a function that takes returns two values")
	}
	if v.Type().In(0) != reflect.TypeOf((*string)(nil)).Elem() {
		panic("call to RegisterStringSetter with something other than a function that takes something other than string")
	}
	if v.Type().Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("call to RegisterStringSetter with something other than a function that returns something other than error")
	}
	settersByType[v.Type().Out(0)] = v
}
