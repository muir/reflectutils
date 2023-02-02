package reflectutils

import (
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var versionRE = regexp.MustCompile(`/v(\d+)$`)

// TypeName is an alternative to reflect.Type's .String() method.  The only
// expected difference is that if there is a package that is versioned, the
// version will appear in the package name.
//
// For example, if there is a foo/v2 package with a Bar type, and you ask
// for for the TypeName, you'll get "foo/v2.Bar" instead of the "foo.Bar" that
// reflect returns.
func TypeName(t reflect.Type) string {
	ts := t.String()
	pkgPath := t.PkgPath()
	if pkgPath != "" {
		if versionRE.MatchString(pkgPath) {
			version := path.Base(pkgPath)
			pn := path.Base(path.Dir(pkgPath))
			revised := strings.Replace(ts, pn, pn+"/"+version, 1)
			if revised != ts {
				return revised
			}
			return "(" + version + ")" + ts
		}
		return ts
	}
	switch t.Kind() { //nolint:exhaustive // not intended to be exhaustive
	case reflect.Ptr: // TODO: change to Pointer when go 1.17 support lapses
		return "*" + TypeName(t.Elem())
	case reflect.Slice:
		return "[]" + TypeName(t.Elem())
	case reflect.Map:
		return "map[" + TypeName(t.Key()) + "]" + TypeName(t.Elem())
	case reflect.Array:
		return "[" + strconv.Itoa(t.Len()) + "]" + TypeName(t.Elem())
	case reflect.Func:
		return "func" + fmtFunc(t)
	case reflect.Chan:
		switch t.ChanDir() {
		case reflect.BothDir:
			return "chan " + TypeName(t.Elem())
		case reflect.SendDir:
			return "chan<- " + TypeName(t.Elem())
		case reflect.RecvDir:
			return "<-chan " + TypeName(t.Elem())
		default:
			return ts
		}
	case reflect.Struct:
		if t.NumField() == 0 {
			return "struct {}"
		}
		fields := make([]string, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				fields[i] = TypeName(f.Type)
			} else {
				fields[i] = f.Name + " " + TypeName(f.Type)
			}
		}
		return "struct { " + strings.Join(fields, "; ") + " }"
	case reflect.Interface:
		n := t.Name()
		if n != "" {
			return n
		}
		if t.NumMethod() == 0 {
			return "interface {}"
		}
		methods := make([]string, t.NumMethod())
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			methods[i] = m.Name + fmtFunc(m.Type)
		}
		return "interface { " + strings.Join(methods, "; ") + " }"
	default:
		return ts
	}
}

func fmtFunc(t reflect.Type) string {
	inputs := make([]string, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		inputs[i] = TypeName(t.In(i))
	}
	outputs := make([]string, t.NumOut())
	for i := 0; i < t.NumOut(); i++ {
		outputs[i] = TypeName(t.Out(i))
	}
	switch t.NumOut() {
	case 0:
		return "(" + strings.Join(inputs, ", ") + ")"
	case 1:
		return "(" + strings.Join(inputs, ", ") + ") " + outputs[0]
	default:
		return "(" + strings.Join(inputs, ", ") + ") (" + strings.Join(outputs, ", ") + ")"
	}
}
