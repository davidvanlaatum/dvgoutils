package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogNilErr(t *testing.T) {
	r := require.New(t)
	b := &bytes.Buffer{}
	h := slog.NewJSONHandler(b, nil)
	logger := slog.New(h)

	logger.Info("test", Err(nil))
	var row map[string]any
	r.NoError(json.NewDecoder(b).Decode(&row))
	r.NotContains(row, "err")
}

func TestLogErr(t *testing.T) {
	r := require.New(t)
	b := &bytes.Buffer{}
	h := slog.NewJSONHandler(b, nil)
	logger := slog.New(h)

	logger.Info("test", Err(fmt.Errorf("a test error")))
	var row map[string]any
	r.NoError(json.NewDecoder(b).Decode(&row))
	r.Contains(row, "err")
	r.Equal("a test error", row["err"])
}
