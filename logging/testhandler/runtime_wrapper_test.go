package testhandler

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_RuntimeWrapper(t *testing.T) {
	r := require.New(t)
	now := time.Now()
	h := NewTestHandler(t)
	w := NewTestRuntimeWrapper(now, h)
	r.True(w.Enabled(t.Context(), slog.LevelDebug))
	r.NotSame(w, w.WithAttrs([]slog.Attr{slog.String("key", "value")}))
	r.NotSame(w, w.WithGroup("group"))
	r.NoError(
		w.WithAttrs([]slog.Attr{slog.String("test", "value")}).Handle(
			t.Context(), slog.Record{
				Time:    now.Add(100 * time.Millisecond),
				Message: "test message",
			},
		),
	)
	r.Len(h.Logs(), 1)
	r.Contains(h.Logs()[0].Msg, "test message")
	r.Len(h.Logs()[0].Attrs, 2)
	r.Equal("test", h.Logs()[0].Attrs[0].Key)
	r.Equal("value", h.Logs()[0].Attrs[0].Value.Any())
	r.Equal(runtimeKey, h.Logs()[0].Attrs[1].Key)
	r.Equal("100ms", h.Logs()[0].Attrs[1].Value.Any())
}

func TestRuntimeWrapperEnabled(t *testing.T) {
	r := require.New(t)
	h := NewTestRuntimeWrapper(time.Now(), slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo}))
	r.True(h.Enabled(t.Context(), slog.LevelInfo))
	r.False(h.Enabled(t.Context(), slog.LevelDebug))
}
