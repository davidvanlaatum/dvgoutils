package colour

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/davidvanlaatum/dvgoutils/logging/testhandler"
	"github.com/stretchr/testify/require"
)

func TestLogColoursPreserveTextRecords(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	ctx, _, _ := testhandler.SetupTestHandler(t)
	var plain, coloured bytes.Buffer
	plainHandler := NewHandler(&plain, &Options{Mode: Never, HandlerOptions: slog.HandlerOptions{Level: slog.LevelDebug}})
	colourHandler := NewHandler(&coloured, &Options{Mode: Always, HandlerOptions: slog.HandlerOptions{Level: slog.LevelDebug}})
	record := slog.NewRecord(time.Time{}, slog.LevelWarn, "waiting for \"host\"\nnext line", 0)
	record.AddAttrs(slog.String("attribute with = and \"quotes\"", "value with spaces = \"quoted\"\n\x1b"), slog.Group("details", "attempt", 2), slog.Any("error", errors.New("not ready")))
	for _, handler := range []slog.Handler{plainHandler, colourHandler} {
		handler = handler.WithAttrs([]slog.Attr{slog.String("host", "node1")}).WithGroup("work")
		r.NoError(handler.Handle(ctx, record))
	}
	strip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	r.Equal(plain.String(), strip.ReplaceAllString(coloured.String(), ""))
	r.Contains(coloured.String(), "\x1b[36mhost\x1b[0m=node1\x1b[0m")
	r.Contains(coloured.String(), "\x1b[36mmsg\x1b[0m=\x1b[32m")
	r.Contains(coloured.String(), "\x1b[36mwork.details.attempt\x1b[0m=2\x1b[0m")
	r.Equal(1, bytes.Count(coloured.Bytes(), []byte{'\n'}))
}

func TestLogColourSelection(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"auto", "always", "never", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			r := require.New(t)
			ctx, _, _ := testhandler.SetupTestHandler(t)
			var out bytes.Buffer
			parsed, err := ParseMode(mode)
			if mode == "invalid" {
				r.EqualError(err, `invalid colour mode "invalid": must be auto, always or never`)
				return
			}
			r.NoError(err)
			handler := NewHandler(&out, &Options{Mode: parsed, LookupEnv: func(key string) (string, bool) { return "1", key == "NO_COLOR" }})
			logger := slog.New(handler)
			logger.DebugContext(ctx, "hidden")
			logger.InfoContext(ctx, "visible")
			r.NotContains(out.String(), "hidden")
			r.Contains(out.String(), "visible")
			r.Equal(mode == "always", bytes.Contains(out.Bytes(), []byte("\x1b[")))
		})
	}
}

type failedLogWriter struct{ err error }

func (w failedLogWriter) Write([]byte) (int, error) { return 0, w.err }

func TestColourLogWriteFailure(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	expected := errors.New("write failed")
	for _, err := range []error{expected, nil} {
		n, got := (colourLogWriter{out: failedLogWriter{err: err}}).Write([]byte("msg=test\n"))
		r.Zero(n)
		if err != nil {
			r.ErrorIs(got, expected)
		} else {
			r.ErrorIs(got, io.ErrShortWrite)
		}
	}
}

func TestLogLevelColours(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		level        slog.Level
		text, colour string
	}{
		{slog.LevelDebug - 1, "DEBUG-1", "\x1b[90m"},
		{slog.LevelDebug, "DEBUG", "\x1b[90m"},
		{slog.LevelInfo, "INFO", "\x1b[34m"},
		{slog.LevelWarn, "WARN", "\x1b[33m"},
		{slog.LevelError, "ERROR", "\x1b[31m"},
		{slog.LevelError + 1, "ERROR+1", "\x1b[31m"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			r := require.New(t)
			ctx, _, _ := testhandler.SetupTestHandler(t)
			var out bytes.Buffer
			handler := NewHandler(&out, &Options{Mode: Always, HandlerOptions: slog.HandlerOptions{Level: slog.LevelDebug - 1}})
			slog.New(handler).Log(ctx, tc.level, "test")
			r.Contains(out.String(), "\x1b[36mlevel\x1b[0m="+tc.colour+tc.text+"\x1b[0m")
		})
	}
}
