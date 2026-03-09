package units

import (
	"bytes"
	"encoding/json"
	"log/slog"
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
		{"1 KB", KiB, "1.0 KiB"},
		{"1.5 KB", 1536, "1.5 KiB"},
		{"1 MB", MiB, "1.0 MiB"},
		{"1 GB", GiB, "1.0 GiB"},
		{"1 TB", TiB, "1.0 TiB"},
		{"1 PB", PiB, "1.0 PiB"},
		{"1 EB", EiB, "1.0 EiB"},
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

func TestLogBytes(t *testing.T) {
	r := require.New(t)
	b := &bytes.Buffer{}
	h := slog.NewJSONHandler(b, nil)
	logger := slog.New(h)

	logger.Info("test", "bytes", Bytes(1024*1.5))
	var row map[string]any
	r.NoError(json.NewDecoder(b).Decode(&row))
	r.Contains(row, "bytes")
	r.Equal("1.5 KiB", row["bytes"])
}
