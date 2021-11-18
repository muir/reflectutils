package reflectutils_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/muir/reflectutils"

	"github.com/stretchr/testify/assert"
)

func TestSplitTag(t *testing.T) {
	ts(t, `env:"foo"`, "env", "foo")
	ts(t, `env:"YO"  flag:"foo,bar"`, "env", "YO", "flag", "foo,bar")
	ts(t, ``)
	ts(t, `env:"YO"  flag:"foo\",bar"`, "env", "YO", "flag", `foo\",bar`)
	ts(t, `env:"YO"  flag:"foo,bar" `, "env", "YO", "flag", "foo,bar")
}

func ts(t *testing.T, tag reflect.StructTag, want ...string) {
	t.Log(tag)
	tags := reflectutils.SplitTag(tag)
	s := make([]string, 0, len(tags)*2)
	if len(tags) == 0 {
		s = nil
	}
	for _, tag := range tags {
		s = append(s, tag.Tag, tag.Value)
	}
	assert.Equal(t, want, s, tag)
}

func TestFill(t *testing.T) {
	cases := []struct {
		tests     interface{}
		model     interface{}
		targetTag string
		metaTag   string
	}{
		{
			tests: struct {
				T1 string `xyz:"a b" want:"{\"P0\":[\"a\",\"b\"]}"`
			}{},
			model: struct {
				P0 []string `tf:"0,split=space" json:",omitempty"`
			}{},
			targetTag: "xyz",
			metaTag:   "tf",
		},
		{
			tests: struct {
				T2 string `xyz:"a,first" want:"{\"Name\":\"a\",\"First\":true}"`
				T3 string `xyz:"a,last" want:"{\"Name\":\"a\",\"First\":false}"`
				T4 string `xyz:"a,!first" want:"{\"Name\":\"a\",\"First\":false}"`
				T5 string `xyz:"a,!last" want:"{\"Name\":\"a\",\"First\":true}"`
			}{},
			model: struct {
				Name  string `pf:"0" json:",omitempty"`
				First *bool  `pf:"first,!last" json:",omitempty"`
			}{},
			targetTag: "xyz",
			metaTag:   "pf",
		},
	}

	for _, tc := range cases {
		tc := tc
		reflectutils.WalkStructElements(reflect.TypeOf(tc.tests), func(f reflect.StructField) bool {
			if tc.metaTag == "" {
				tc.metaTag = "pt"
			}
			if tc.targetTag == "" {
				tc.targetTag = "xyz"
			}
			t.Run(f.Name+"_"+tc.targetTag+"_"+tc.metaTag,
				func(t *testing.T) {
					t.Logf("%s: %s", f.Name, f.Tag)
					got := reflect.New(reflect.TypeOf(tc.model)).Interface()
					err := reflectutils.SplitTag(f.Tag).Set().Get(tc.targetTag).Fill(got, reflectutils.WithTag(tc.metaTag))
					if !assert.NoErrorf(t, err, "extract tag %s", f.Name) {
						return
					}
					want := reflect.New(reflect.TypeOf(tc.model)).Interface()
					err = json.Unmarshal([]byte(f.Tag.Get("want")), want)
					if !assert.NoErrorf(t, err, "extract want %s", f.Name) {
						return
					}
					assert.Equal(t, want, got, f.Name)
				})
			return true
		})
	}
}
