package reflectutils_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/muir/reflectutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Foo string

func (fp *Foo) UnmarshalText(b []byte) error {
	*fp = Foo("~" + string(b) + "~")
	return nil
}

func TestStringSetter(t *testing.T) {
	type tsType struct {
		Int        int         `value:"38"`
		Int8       int8        `value:"-9"`
		Int16      int16       `value:"329"`
		Int32      int32       `value:"-32902"`
		Int64      int64       `value:"3292929"`
		Uint       uint        `value:"202"`
		Uint8      uint8       `value:"99"`
		Uint16     uint16      `value:"3020"`
		Uint32     uint32      `value:"92020"`
		Uint64     uint64      `value:"320202"`
		Float32    float32     `value:"3.9"`
		Float64    float64     `value:"4.32e7" want:"4.32e+07"`
		String     string      `value:"foo"`
		Bool       bool        `value:"false"`
		IntP       *int        `value:"-82"`
		Int8P      *int8       `value:"-2"`
		Int16P     *int16      `value:"-39"`
		Int32P     *int32      `value:"-329"`
		Int64P     *int64      `value:"-3292"`
		UintP      *uint       `value:"239"`
		Uint8P     *uint8      `value:"92"`
		Uint16P    *uint16     `value:"330"`
		Uint32P    *uint32     `value:"239239"`
		Uint64P    *uint64     `value:"3923"`
		Float32P   *float32    `value:"3.299"`
		Float64P   *float64    `value:"9.2"`
		StringP    *string     `value:"foop"`
		Complex64  *complex64  `value:"4+3i"     want:"(4+3i)"`
		Complex128 *complex128 `value:"3.9+2.6i" want:"(3.9+2.6i)"`
		BoolP      *bool       `value:"true"`
		IntSlice   []int       `value:"3,9,-10"  want:"[3 9 -10]"`
		IntArray   [2]int      `value:"22,11"    want:"[22 11]"`
		Foo        Foo         `value:"foo"      want:"~foo~"`
		FooArray   [2]Foo      `value:"a,b,c"    want:"[~a~ ~b,c~]"`
		FooP       *Foo        `value:"foo"      want:"~foo~"`
		SS1        []string    `value:"foo/bar"  want:"[foo/bar]"`
		SS2        []string    `value:"foo/bar"  want:"[foo bar]"   split:"/"`
		SS3        []string    `value:"foo,bar"  want:"[foo,bar]"   split:""`
		SS4        []string    `value:"foo,bar"  want:"[foo bar]"   split:","`
		SA1        [2]string   `value:"foo/bar"  want:"[foo/bar ]"`
		SA2        [2]string   `value:"foo/bar"  want:"[foo bar]"   split:"/"`
		SA3        [2]string   `value:"foo,bar"  want:"[foo,bar ]"  split:""`
		SS5        []string    `value:"foo"      want:"[foo bar]"   value2:"bar"`
		SS6        []string    `value:"foo"      want:"[bar]"       value2:"bar" sa:"f"`
	}
	var ts tsType
	vp := reflect.ValueOf(&ts)
	v := reflect.Indirect(vp)
	var count int
	reflectutils.WalkStructElements(v.Type(), func(f reflect.StructField) bool {
		t.Logf("field %s, a %s", f.Name, f.Type)
		value, ok := f.Tag.Lookup("value")
		if !assert.Truef(t, ok, "input value for %s", f.Name) {
			return true
		}
		want, ok := f.Tag.Lookup("want")
		if !ok {
			want = value
		}
		var opts []reflectutils.StringSetterArg
		if split, ok := f.Tag.Lookup("split"); ok {
			t.Log("  splitting on", split)
			opts = append(opts, reflectutils.WithSplitOn(split))
		}
		if sa, ok := f.Tag.Lookup("sa"); ok {
			b, err := strconv.ParseBool(sa)
			require.NoError(t, err, "parse sa")
			t.Log("  slice append", b)
			opts = append(opts, reflectutils.SliceAppend(b))
		}

		fn, err := reflectutils.MakeStringSetter(f.Type, opts...)
		if !assert.NoErrorf(t, err, "make string setter for %s", f.Name) {
			return true
		}
		e := v.FieldByIndex(f.Index)
		err = fn(e, value)
		if assert.NoError(t, err, "set %s to '%s'", f.Name, value) {
			value2, ok := f.Tag.Lookup("value2")
			if ok {
				err := fn(e, value2)
				assert.NoError(t, err, "set value2")
			}
			ge := e
			if f.Type.Kind() == reflect.Ptr {
				ge = e.Elem()
			}
			assert.Equalf(t, want, fmt.Sprintf("%+v", ge.Interface()), "got setting %s to '%s'", f.Name, value)
		}
		count++
		return true
	})
	assert.Equal(t, v.NumField(), count, "number of fields tested")
}
