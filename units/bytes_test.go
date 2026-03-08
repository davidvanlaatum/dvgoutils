package units

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytes_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    Bytes
		want string
	}{
		{"0 bytes", 0, "0 B"},
		{"512 bytes", 512, "512 B"},
		{"1 KB", 1024, "1.0 KiB"},
		{"1.5 KB", 1536, "1.5 KiB"},
		{"1 MB", 1024 * 1024, "1.0 MiB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GiB"},
		{"1 TB", 1024 * 1024 * 1024 * 1024, "1.0 TiB"},
		{"1 PB", 1024 * 1024 * 1024 * 1024 * 1024, "1.0 PiB"},
		{"1 EB", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, "1.0 EiB"},
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
