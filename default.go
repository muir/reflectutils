package reflectutils

import (
	"reflect"

	"github.com/pkg/errors"
)

// FillInDefaultValues will look at struct tags, looking for a
// "default" tag. If it finds one and the value in the struct
// is not set, then it will try to turn the string into a value.
// A pointer that is not nil is considered set and will not be
// overrridden.
//
// It can handle fields with supported types. The supported types
// are: any type that implements encoding.TextUnmarshaler;
// time.Duration; and any types preregistered with
// RegisterStringSetter(); pointers to any of the above types.
//
// The argument must be a pointer to a struct. Anything else will
// return error. A nil pointer is not allowed.
func FillInDefaultValues(pointerToStruct any) error {
	ptr := reflect.ValueOf(pointerToStruct)
	if ptr.Kind() != reflect.Ptr {
		return errors.Errorf("cannot fill in defaults for anything (%s) but a valid pointer", ptr.Kind())
	}
	if ptr.IsNil() {
		return errors.Errorf("cannot fill in defaults for a nil pointer")
	}
	valueType := ptr.Type().Elem()
	if valueType.Kind() != reflect.Struct {
		return errors.Errorf("cannot fill in defaults for non-structs (%s)", valueType)
	}
	var firstError error
	WalkStructElements(valueType, func(field reflect.StructField) bool {
		tag, ok := LookupTag(field.Tag, "default")
		if !ok {
			return true
		}
		value := ptr.Elem().FieldByIndex(field.Index)
		if !value.CanSet() {
			return true
		}
		if !value.IsZero() {
			return true
		}
		setter, err := MakeStringSetter(field.Type)
		if err != nil {
			firstError = err // override since this is worse
			return true
		}
		err = setter(value, tag.Value)
		if err != nil && firstError == nil {
			firstError = err
		}
		return true
	})
	return firstError
}
