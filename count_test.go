package dvgoutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountSlice(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		slice     []int
		predicate func(int) bool
		expected  int
	}{
		"matching values": {
			slice:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(i int) bool { return i%2 == 0 },
			expected:  3,
		},
		"no matching values": {
			slice:     []int{1, 3, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			expected:  0,
		},
		"empty slice": {
			slice:     nil,
			predicate: func(int) bool { return true },
			expected:  0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.expected, CountSlice(test.slice, test.predicate))
		})
	}
}
