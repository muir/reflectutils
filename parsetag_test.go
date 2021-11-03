package reflectutils_test

import (
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
