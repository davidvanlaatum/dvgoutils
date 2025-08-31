package dvgoutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPtr(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	val := 42
	ptr := Ptr(val)
	r.NotNil(ptr)
	r.Equal(val, *ptr)
}
