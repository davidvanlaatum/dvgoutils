package units

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBits_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    Bits
		want string
	}{
		{"0 bits", 0, "0 b"},
		{"512 bits", 512, "512 b"},
		{"1 Kb", 1000, "1.0 Kb"},
		{"1.5 Kb", 1500, "1.5 Kb"},
		{"1 Mb", 1000 * 1000, "1.0 Mb"},
		{"1 Gb", 1000 * 1000 * 1000, "1.0 Gb"},
		{"1 Tb", 1000 * 1000 * 1000 * 1000, "1.0 Tb"},
		{"1 Pb", 1000 * 1000 * 1000 * 1000 * 1000, "1.0 Pb"},
		{"1 Eb", 1000 * 1000 * 1000 * 1000 * 1000 * 1000, "1.0 Eb"},
	}
	for _, test := range tests {
		test := test
		t.Run(
			test.name, func(t *testing.T) {
				t.Parallel()
				r := require.New(t)
				r.Equal(test.want, test.b.String())
			},
		)
	}
}
