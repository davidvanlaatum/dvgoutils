package testhandler

import (
	"context"
	"log/slog"
	"time"
)

const runtimeKey = "test-runtime"

type TestRuntimeWrapper struct {
	start time.Time
	inner slog.Handler
}

func NewTestRuntimeWrapper(start time.Time, inner slog.Handler) *TestRuntimeWrapper {
	return &TestRuntimeWrapper{start: start, inner: inner}
}

func (t *TestRuntimeWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return t.inner.Enabled(ctx, level)
}

func (t *TestRuntimeWrapper) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(slog.String(runtimeKey, record.Time.Sub(t.start).String()))
	return t.inner.Handle(ctx, record)
}

func (t *TestRuntimeWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TestRuntimeWrapper{
		start: t.start,
		inner: t.inner.WithAttrs(attrs),
	}
}

func (t *TestRuntimeWrapper) WithGroup(name string) slog.Handler {
	return &TestRuntimeWrapper{
		start: t.start,
		inner: t.inner.WithGroup(name),
	}
}

var _ slog.Handler = (*TestRuntimeWrapper)(nil)
