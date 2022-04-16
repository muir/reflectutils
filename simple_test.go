package reflectutils_test

import (
	"reflect"
	"testing"

	"github.com/muir/reflectutils"

	"github.com/stretchr/testify/assert"
)

func TestNonElement(t *testing.T) {
	a := []map[string][3]int{
		{
			"foo": {8, 3, 9},
		},
	}
	got := reflectutils.NonElement(reflect.TypeOf(&a)).String()
	t.Log(got)
	assert.Equal(t, reflect.Int.String(), reflectutils.NonElement(reflect.TypeOf(&a)).String())
}
