package units

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"math"
	"testing"
	"time"

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
		{"1 Kb", Kb, "1.0 Kb"},
		{"1.5 Kb", 1500, "1.5 Kb"},
		{"1 Mb", Mb, "1.0 Mb"},
		{"1 Gb", Gb, "1.0 Gb"},
		{"1 Tb", Tb, "1.0 Tb"},
		{"1 Pb", Pb, "1.0 Pb"},
		{"1 Eb", Eb, "1.0 Eb"},
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

func TestLogBits(t *testing.T) {
	r := require.New(t)
	b := &bytes.Buffer{}
	h := slog.NewJSONHandler(b, nil)
	logger := slog.New(h)

	logger.Info("test", "bits", Bits(1500))
	var row map[string]any
	r.NoError(json.NewDecoder(b).Decode(&row))
	r.Contains(row, "bits")
	r.Equal("1.5 Kb", row["bits"])
}

func TestBits_PerSecond(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	r.Equal(BitsPerSecond(1500), Bits(3*Kb).PerSecond(2*time.Second))
	r.Equal(BitsPerSecond(2000), Kb.PerSecond(500*time.Millisecond))
	r.True(math.IsNaN(float64(Kb.PerSecond(0))))
	r.True(math.IsNaN(float64(Kb.PerSecond(-time.Second))))
}

func TestBitsPerSecond_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    BitsPerSecond
		want string
	}{
		{"0 bits per second", 0, "0.0 b/s"},
		{"512 bits per second", 512, "512.0 b/s"},
		{"1 Kb per second", 1000, "1.0 Kb/s"},
		{"1.5 Kb per second", 1500, "1.5 Kb/s"},
		{"1 Mb per second", 1000 * 1000, "1.0 Mb/s"},
		{"above Eb per second", BitsPerSecond(8 * float64(Eb)), "8.0 Eb/s"},
		{"not a number", BitsPerSecond(math.NaN()), "NaN b/s"},
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

func TestLogBitsPerSecond(t *testing.T) {
	r := require.New(t)
	b := &bytes.Buffer{}
	h := slog.NewJSONHandler(b, nil)
	logger := slog.New(h)

	logger.Info("test", "rate", BitsPerSecond(1500))
	var row map[string]any
	r.NoError(json.NewDecoder(b).Decode(&row))
	r.Contains(row, "rate")
	r.Equal("1.5 Kb/s", row["rate"])
}
