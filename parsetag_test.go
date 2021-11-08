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
	type tagData struct {
		P0 []string `tf:"0,split=space" json:",omitempty"`
	}
	type testStruct struct {
		T1 string `xyz:"a b" want:"{\"P0\":[\"a\",\"b\"]}"`
	}
	var x testStruct
	reflectutils.WalkStructElements(reflect.TypeOf(x), func(f reflect.StructField) bool {
		var got tagData
		t.Logf("%s: %s", f.Name, f.Tag)
		err := reflectutils.SplitTag(f.Tag).Set().Get("xyz").Fill(&got, reflectutils.WithTag("tf"))
		if !assert.NoErrorf(t, err, "extract tag %s", f.Name) {
			return true
		}
		var want tagData
		err = json.Unmarshal([]byte(f.Tag.Get("want")), &want)
		if !assert.NoErrorf(t, err, "extract want %s", f.Name) {
			return true
		}
		assert.Equal(t, want, got, f.Name)
		return true
	})
}
