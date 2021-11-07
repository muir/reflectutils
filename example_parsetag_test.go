package reflectutils_test

import (
	"fmt"
	"reflect"

	"github.com/muir/reflectutils"
)

type ExampleStruct struct {
	String string `foo:"something,N=9,square,!jump"`
	Bar    string `foo:"different,!square,jump,ignore=xyz"`
}

type TagExtractorType struct {
	Name   string `pt:"0"`
	Square bool   `pt:"square"`
	Jump   bool   `pt:"jump"`
	Ignore string `pt:"-"`
	N      int
}

// Fill is a helper for when you're working with tags.
func ExampleTag_Fill() {
	var es ExampleStruct
	t := reflect.TypeOf(es)
	reflectutils.WalkStructElements(t, func(f reflect.StructField) bool {
		var tet TagExtractorType
		err := reflectutils.SplitTag(f.Tag).Set().Get("foo").Fill(&tet)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s: %+v\n", f.Name, tet)
		return true
	})
	// Output: String: {Name:something Square:true Jump:false Ignore: N:9}
	// Bar: {Name:different Square:false Jump:true Ignore: N:0}
}
