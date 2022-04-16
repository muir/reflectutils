package reflectutils

import (
	"reflect"
)

// NonPointer unwraps pointer types until a type that isn't
// a pointer is found.
func NonPointer(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// NonElement unwraps pointers, slices, arrays, and maps until
// it finds a type that doesn't support Elem.  It returns that
// type.
func NonElement(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Array, reflect.Slice:
			t = t.Elem()
		default:
			return t
		}
	}
}
