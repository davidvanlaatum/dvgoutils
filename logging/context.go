package logging

import (
	"context"
	"log/slog"
)

type contextKey string

var contextKeyLogger = contextKey("logger")

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKeyLogger).(*slog.Logger); ok && logger != nil {
		return logger
	}
	panic("no logger in context")
}
