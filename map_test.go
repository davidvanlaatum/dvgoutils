package dvgoutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapSlice(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	s := []int{1, -2, 3, 4, 5, 6}
	squared := MapSlice(s, func(i int) uint { return uint(i * i) })
	r.Equal([]uint{1, 4, 9, 16, 25, 36}, squared)
}
