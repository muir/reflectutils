package reflectutils

import (
	"encoding"
	"flag"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var textUnmarshallerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
var flagValueType = reflect.TypeOf((*flag.Value)(nil)).Elem()

type stringSetterOpts struct {
	split       string
	sliceAppend bool
}

type StringSetterArg func(*stringSetterOpts)

// WithSplitOn specifies how to split strings into slices
// or arrays.  An empty string indicates that input strings
// should not be split.  That's different than the behavior
// of strings.Split().  If unspecified, strings will be split
// on comma (,).
func WithSplitOn(s string) StringSetterArg {
	return func(o *stringSetterOpts) {
		o.split = s
	}
}

// Controls the behavior for setting existing existing slices.
// If this is true (the default) then additional setting to a
// slice appends to the existing slice.  If false, slices are
// replaced.
func SliceAppend(b bool) StringSetterArg {
	return func(o *stringSetterOpts) {
		o.sliceAppend = b
	}
}

// MakeStringSetter handles setting a reflect.Value from a string.
// Based on type, it returns a function to do the work.  It is assumed that the
// reflect.Type matches the reflect.Value.  If not, panic is likely.
//
// For arrays and slices, strings are split on comma to create the values for the
// elements.
//
// Any type that matches a type registered with RegisterStringSetter will be
// unpacked with the corresponding function.  A string setter is pre-registered
// for time.Duration
// Anything that implements encoding.TextUnmarshaler will be unpacked that way.
// Anything that implements flag.Value will be unpacked that way.
//
// Maps, structs, channels, interfaces, channels, and funcs are not supported unless
// they happen to implent encoding.TextUnmarshaler.
func MakeStringSetter(t reflect.Type, optArgs ...StringSetterArg) (func(target reflect.Value, value string) error, error) {
	opts := stringSetterOpts{
		split:       ",",
		sliceAppend: true,
	}
	for _, f := range optArgs {
		f(&opts)
	}
	if setter, ok := settersByType[t]; ok {
		return func(target reflect.Value, value string) error {
			out := setter.Call([]reflect.Value{reflect.ValueOf(value)})
			if !out[1].IsNil() {
				return out[1].Interface().(error)
			}
			target.Set(out[0])
			return nil
		}, nil
	}
	if t.AssignableTo(textUnmarshallerType) {
		return func(target reflect.Value, value string) error {
			p := reflect.New(t.Elem())
			target.Set(p)
			err := target.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}, nil
	}
	if reflect.PtrTo(t).AssignableTo(textUnmarshallerType) {
		return func(target reflect.Value, value string) error {
			err := target.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
			return errors.WithStack(err)
		}, nil
	}
	if t.AssignableTo(flagValueType) {
		return func(target reflect.Value, value string) error {
			p := reflect.New(t.Elem())
			target.Set(p)
			err := target.Interface().(flag.Value).Set(value)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}, nil
	}
	if reflect.PtrTo(t).AssignableTo(flagValueType) {
		return func(target reflect.Value, value string) error {
			err := target.Addr().Interface().(flag.Value).Set(value)
			return errors.WithStack(err)
		}, nil
	}
	switch t.Kind() {
	case reflect.Ptr:
		setElem, err := MakeStringSetter(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(target reflect.Value, value string) error {
			p := reflect.New(t.Elem())
			target.Set(p)
			err := setElem(target.Elem(), value)
			if err != nil {
				return err
			}
			return nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(target reflect.Value, value string) error {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			target.SetInt(i)
			return nil
		}, nil
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(target reflect.Value, value string) error {
			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			target.SetUint(i)
			return nil
		}, nil
	case reflect.Float32, reflect.Float64:
		return func(target reflect.Value, value string) error {
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			target.SetFloat(f)
			return nil
		}, nil
	case reflect.String:
		return func(target reflect.Value, value string) error {
			target.SetString(value)
			return nil
		}, nil
	case reflect.Complex64, reflect.Complex128:
		return func(target reflect.Value, value string) error {
			c, err := strconv.ParseComplex(value, 128)
			if err != nil {
				return errors.WithStack(err)
			}
			target.SetComplex(c)
			return nil
		}, nil
	case reflect.Bool:
		return func(target reflect.Value, value string) error {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return errors.WithStack(err)
			}
			target.SetBool(b)
			return nil
		}, nil
	case reflect.Array:
		setElem, err := MakeStringSetter(t.Elem())
		if err != nil {
			return nil, err
		}
		if opts.split == "" {
			return func(target reflect.Value, value string) error {
				return setElem(target.Index(0), value)
			}, nil
		}
		return func(target reflect.Value, value string) error {
			for i, v := range strings.SplitN(value, opts.split, target.Cap()) {
				err := setElem(target.Index(i), v)
				if err != nil {
					return err
				}
			}
			return nil
		}, nil
	case reflect.Slice:
		setElem, err := MakeStringSetter(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(target reflect.Value, value string) error {
			var values []string
			if opts.split != "" {
				values = strings.Split(value, opts.split)
			} else {
				values = []string{value}
			}
			a := reflect.MakeSlice(target.Type(), len(values), len(values))
			for i, v := range values {
				err := setElem(a.Index(i), v)
				if err != nil {
					return err
				}
			}
			if target.IsNil() || !opts.sliceAppend {
				target.Set(a)
			} else {
				target.Set(reflect.AppendSlice(target, a))
			}
			return nil
		}, nil
	case reflect.Map:
		fallthrough
	case reflect.Struct:
		fallthrough
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Invalid, reflect.UnsafePointer:
		fallthrough
	default:
		return nil, errors.Errorf("type %s not supported", t)
	}
}
