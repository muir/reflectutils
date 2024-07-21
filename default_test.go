package reflectutils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/muir/reflectutils"
)

type DefaultExample struct {
	I5         int `default:"5"`
	IAlready2  int `default:"10"`
	INoDefault int
	IPointer7  *int `default:"7"`
	IPtrSet0   *int `default:"10"`
}

type BadDefault1 struct {
	I5 time.Duration `default:"5"` // invalid value
}

type BadDefault2 struct {
	I any `default:"5"`
}

func TestDefaultSetter(t *testing.T) {
	var zero int
	input := DefaultExample{
		IAlready2: 2,
		IPtrSet0:  &zero,
	}
	require.NoError(t, reflectutils.FillInDefaultValues(&input))
	assert.Equal(t, 5, input.I5)
	assert.Equal(t, 2, input.IAlready2)
	assert.Equal(t, 0, input.INoDefault)
	if assert.NotNil(t, input.IPointer7) {
		assert.Equal(t, 7, *input.IPointer7)
	}
	if assert.NotNil(t, input.IPtrSet0) {
		assert.Equal(t, 0, *input.IPtrSet0)
	}
	require.Error(t, reflectutils.FillInDefaultValues(&BadDefault1{}))
	require.Error(t, reflectutils.FillInDefaultValues(&BadDefault2{}))
}
