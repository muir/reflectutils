package reflectutils_test

import (
	"fmt"
	"reflect"

	"github.com/muir/reflectutils"
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
