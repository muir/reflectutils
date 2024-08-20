package reflectutils_test

import (
	"fmt"
	"reflect"

	"github.com/muir/reflectutils"
	"github.com/pkg/errors"
)

type S struct {
	I1 int
	S  string
	M  T
}

type T struct {
	I2 int
}

func makeIntDoubler(t reflect.Type) func(v reflect.Value) {
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("makeIntDoubler only supports pointers to structs")
	}
	var ints []reflect.StructField
	reflectutils.WalkStructElements(t, func(f reflect.StructField) bool {
		if f.Type.Kind() == reflect.Int {
			ints = append(ints, f)
		}
		return true
	})
	return func(v reflect.Value) {
		v = v.Elem()
		for _, f := range ints {
			i := v.FieldByIndex(f.Index)
			i.SetInt(int64(i.Interface().(int)) * 2)
		}
	}
}

func makeIntDoublerWithError(t reflect.Type) (func(v reflect.Value), error) {
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("makeIntDoublerWithError only supports pointers to structs")
	}
	var ints []reflect.StructField
	err := reflectutils.WalkStructElementsWithError(t, func(f reflect.StructField) (bool, error) {
		if f.Type.Kind() == reflect.Int {
			ints = append(ints, f)
		} else {
			return false, errors.Errorf("not allow element is non-Int")
		}
		return true, nil
	})
	return func(v reflect.Value) {
		v = v.Elem()
		for _, f := range ints {
			i := v.FieldByIndex(f.Index)
			i.SetInt(int64(i.Interface().(int)) * 2)
		}
	}, err
}

func Example() {
	s := S{
		I1: 3,
		S:  "string",
		M: T{
			I2: 5,
		},
	}
	v := reflect.ValueOf(&s)
	doubler := makeIntDoubler(v.Type())
	doubler(v)
	fmt.Printf("%v\n", v.Interface())

	// Output: &{6 string {10}}
}

func ExampleError() {
	s := S{
		I1: 3,
		S:  "string",
		M: T{
			I2: 5,
		},
	}
	v := reflect.ValueOf(&s)
	doubler, err := makeIntDoublerWithError(v.Type())
	doubler(v)
	fmt.Printf("%v\n", v.Interface())
	fmt.Printf("%v", err)

	// Output: &{6 string {5}}
	// not allow element is non-Int
}
