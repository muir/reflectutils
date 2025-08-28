package reflectutils

import (
	"reflect"
	"testing"

	v1 "github.com/muir/reflectutils/internal/foo"
	v2 "github.com/muir/reflectutils/internal/foo/v2"

	"github.com/stretchr/testify/assert"
)

func TestVersionedNames(t *testing.T) {
	type xbar struct {
		v2.Bar
	}
	type ybar struct {
		xbar
		v1 v1.Bar //nolint:unused // unused and we don't care
	}
	cases := []struct {
		thing interface{}
		want  string
	}{
		{
			thing: v1.Bar{},
		},
		{
			thing: v2.Bar{},
			want:  "foo/v2.Bar",
		},
		{
			thing: &v1.Bar{},
			want:  "*foo.Bar",
		},
		{
			thing: &v2.Bar{},
			want:  "*foo/v2.Bar",
		},
		{
			thing: []v2.Bar{},
			want:  "[]foo/v2.Bar",
		},
		{
			thing: (func(*v1.Bar, *v2.Bar) (string, error))(nil),
			want:  "func(*foo.Bar, *foo/v2.Bar) (string, error)",
		},
		{
			thing: [8]v2.Bar{},
			want:  "[8]foo/v2.Bar",
		},
		{
			thing: make(chan *v2.Bar),
			want:  "chan *foo/v2.Bar",
		},
		{
			thing: make(chan *v1.Bar),
			want:  "chan *foo.Bar",
		},
		{
			//nolint:gocritic // could remove some parens
			thing: (chan<- *v1.Bar)(nil),
			want:  "chan<- *foo.Bar",
		},
		{
			//nolint:gocritic // could remove some parens
			thing: (chan<- *v2.Bar)(nil),
			want:  "chan<- *foo/v2.Bar",
		},
		{
			//nolint:gocritic // could remove some parens
			thing: (<-chan *v2.Bar)(nil),
			want:  "<-chan *foo/v2.Bar",
		},
		{
			//nolint:gocritic // could remove some parens
			thing: (<-chan *v1.Bar)(nil),
			want:  "<-chan *foo.Bar",
		},
		{
			thing: (func(interface {
				V1() v1.Bar
				V2() v2.Bar
			}))(nil),
			want: "func(interface { V1() foo.Bar; V2() foo/v2.Bar })",
		},
		{
			thing: (func(interface{}))(nil),
		},
		{
			thing: struct {
				v1 v1.Bar
				v2 v2.Bar
			}{},
			want: "struct { v1 foo.Bar; v2 foo/v2.Bar }",
		},
		{
			thing: struct{}{},
		},
		{
			thing: struct {
				v2.Bar
				v1 v1.Bar
			}{},
			want: "struct { foo/v2.Bar; v1 foo.Bar }",
		},
		{
			thing: struct {
				ybar
			}{},
			// want: "struct { reflectutils.ybar }",
		},
	}

	for _, tc := range cases {
		want := tc.want
		if want == "" {
			want = reflect.TypeOf(tc.thing).String()
		}
		t.Logf("%+v wanting %s", tc.thing, want)
		assert.Equal(t, want, TypeName(reflect.TypeOf(tc.thing)))
	}
}
