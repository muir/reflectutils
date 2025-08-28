package reflectutils

import (
	"errors"
	"reflect"
)

// WalkStructElements recursively visits the fields in a structure calling a
// callback for each field.  It modifies the reflect.StructField object
// so that Index is relative to the root object originally passed to
// WalkStructElements.  This allows the FieldByIndex method on a reflect.Value
// matching the original struct to fetch that field.
//
// WalkStructElements should be called with a reflect.Type whose Kind() is
// reflect.Struct or whose Kind() is reflect.Ptr and Elem.Type() is reflect.Struct.
// All other types will simply be ignored.
//
// The return value from f only matters when the type of the field is a struct.  In
// that case, a false value prevents recursion.
func WalkStructElements(t reflect.Type, f func(reflect.StructField) bool) {
	if t.Kind() == reflect.Struct {
		doWalkStructElements(t, []int{}, f)
	}
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		doWalkStructElements(t.Elem(), []int{}, f)
	}
}

func doWalkStructElements(t reflect.Type, path []int, f func(reflect.StructField) bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		np := copyIntSlice(path)
		np = append(np, field.Index...)
		field.Index = np
		if f(field) && field.Type.Kind() == reflect.Struct {
			doWalkStructElements(field.Type, np, f)
		}
	}
}

// WalkStructElementsWithError recursively visits the fields in a structure calling a
// callback for each field.  It modifies the reflect.StructField object
// so that Index is relative to the root object originally passed to
// WalkStructElementsWithError.  This allows the FieldByIndex method on a reflect.Value
// matching the original struct to fetch that field.
//
// WalkStructElementsWithError should be called with a reflect.Type whose Kind() is
// reflect.Struct or whose Kind() is reflect.Ptr and Elem.Type() is reflect.Struct.
// All other types will simply be ignored.
//
// A non-nil return value from the called function stops iteration and recursion and becomes
// the return value.
//
// A special error return value, [DoNotRecurseSignalErr] is not considered an error (it will
// not become the return value, and it does not stop iteration) but it will prevent recursion if returned
// on a field that is itself a struct.
func WalkStructElementsWithError(t reflect.Type, f func(reflect.StructField) error) error {
	if t.Kind() == reflect.Struct {
		return doWalkStructElementsWithError(t, []int{}, f)
	}
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return doWalkStructElementsWithError(t.Elem(), []int{}, f)
	}
	return nil
}

//nolint:staticcheck // error name doesn't match pattern
var DoNotRecurseSignalErr = errors.New("walkstruct: do not recurse signal")

func doWalkStructElementsWithError(t reflect.Type, path []int, f func(reflect.StructField) error) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		np := copyIntSlice(path)
		np = append(np, field.Index...)
		field.Index = np
		err := f(field)
		if errors.Is(err, DoNotRecurseSignalErr) {
			continue
		}
		if err != nil {
			return err
		}
		if field.Type.Kind() == reflect.Struct {
			err = doWalkStructElementsWithError(field.Type, np, f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyIntSlice(in []int) []int {
	c := make([]int, len(in), len(in)+1)
	copy(c, in)
	return c
}
