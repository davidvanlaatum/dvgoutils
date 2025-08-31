package dvgoutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterSlice(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	s := []int{1, 2, 3, 4, 5, 6}
	even := FilterSlice(s, func(i int) bool { return i%2 == 0 })
	r.Equal([]int{2, 4, 6}, even)
}
