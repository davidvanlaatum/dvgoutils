package colour

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type descriptorWriter struct{ bytes.Buffer }

func (*descriptorWriter) Fd() uintptr { return 123 }

func TestAutomaticDetection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, noColour, term string
		terminal, want       bool
	}{
		{name: "terminal", terminal: true, want: true},
		{name: "redirected"},
		{name: "no colour", noColour: "1", terminal: true},
		{name: "zero still disables", noColour: "0", terminal: true},
		{name: "dumb", term: "dumb", terminal: true},
		{name: "normal terminal", term: "xterm-256color", terminal: true, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			lookup := func(key string) (string, bool) {
				value := map[string]string{"NO_COLOR": tc.noColour, "TERM": tc.term}[key]
				return value, value != ""
			}
			calls := 0
			isTerminal := func(fd uintptr) bool {
				calls++
				r.Equal(uintptr(123), fd)
				return tc.terminal
			}
			r.Equal(tc.want, autoColour(&descriptorWriter{}, lookup, isTerminal))
			r.False(autoColour(&bytes.Buffer{}, lookup, isTerminal))
			if tc.noColour != "" || tc.term == "dumb" {
				r.Zero(calls, "environment overrides must skip console preparation")
			} else {
				r.Equal(1, calls)
			}
		})
	}
}

func TestDefaultsAndEnvironment(t *testing.T) {
	r := require.New(t)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "dumb")
	for _, opts := range []*Options{nil, {}, {Mode: Never}, {Mode: Always}} {
		var out bytes.Buffer
		logger := slog.New(NewHandler(&out, opts))
		logger.Debug("hidden")
		logger.Info("visible")
		r.NotContains(out.String(), "hidden")
		r.Contains(out.String(), "visible")
		r.Equal(opts != nil && opts.Mode == Always, bytes.Contains(out.Bytes(), []byte("\x1b[")))
	}
	file, err := os.CreateTemp(t.TempDir(), "log")
	r.NoError(err)
	t.Cleanup(func() { r.NoError(file.Close()) })
	slog.New(NewHandler(file, nil)).Info("plain")
	data, err := os.ReadFile(file.Name())
	r.NoError(err)
	r.Contains(string(data), "msg=plain")
	r.NotContains(string(data), "\x1b[")
	r.PanicsWithValue("invalid colour mode 255", func() {
		NewHandler(io.Discard, &Options{Mode: Mode(255)})
	})
	mode, err := ParseMode("")
	r.NoError(err)
	r.Equal(Auto, mode)
}

type resolvedValue struct{}

func (resolvedValue) LogValue() slog.Value {
	return slog.GroupValue(slog.String("resolved", "yes"))
}

func TestStandardOptionsAndConcurrentDerivedHandlers(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	var out bytes.Buffer
	var level slog.LevelVar
	logger := slog.New(NewHandler(&out, &Options{
		Mode: Always,
		HandlerOptions: slog.HandlerOptions{
			Level: &level, AddSource: true,
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return attr
			},
		},
	}))
	logger.Debug("hidden")
	r.Empty(out.String())
	level.Set(slog.LevelDebug)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			logger.With("worker", i).WithGroup("task").Debug("visible", "value", resolvedValue{})
		})
	}
	wg.Wait()
	plain := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(out.String(), "")
	r.Equal(20, bytes.Count([]byte(plain), []byte("\n")))
	r.Equal(20, bytes.Count([]byte(plain), []byte("level=DEBUG")))
	r.Equal(20, bytes.Count([]byte(plain), []byte("task.value.resolved=yes")))
	r.Contains(plain, "source=")
	r.NotContains(plain, "time=")
}

func TestHandlerPropagatesWriteError(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	expected := errors.New("write failed")
	handler := NewHandler(failedLogWriter{err: expected}, &Options{Mode: Always})
	r.ErrorIs(handler.Handle(t.Context(), slog.Record{Message: "test"}), expected)
}

func TestRedirectedOutput(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	lookup := func(string) (string, bool) { return "", false }
	file, err := os.CreateTemp(t.TempDir(), "log")
	r.NoError(err)
	t.Cleanup(func() { r.NoError(file.Close()) })
	reader, writer, err := os.Pipe()
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(reader.Close())
		r.NoError(writer.Close())
	})
	for _, out := range []*os.File{file, writer} {
		r.False(prepareTerminal(out.Fd()))
		r.False(autoColour(out, lookup, prepareTerminal))
	}
	slog.New(NewHandler(file, &Options{LookupEnv: lookup})).Info("plain")
	data, err := os.ReadFile(file.Name())
	r.NoError(err)
	r.Contains(string(data), "msg=plain")
	r.NotContains(string(data), "\x1b[")
}
