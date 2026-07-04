package logging_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/davidvanlaatum/dvgoutils/logging"
)

func exampleLogger() *slog.Logger {
	handler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return attr
			},
		},
	)
	return slog.New(handler)
}

func ExampleWithLogger() {
	ctx := logging.WithLogger(context.Background(), exampleLogger())

	logging.FromContext(ctx).Info("starting job", slog.String("component", "worker"))

	// Output:
	// level=INFO msg="starting job" component=worker
}

func ExampleWithLogger_scopedAttributes() {
	ctx := logging.WithLogger(context.Background(), exampleLogger())
	log := logging.FromContext(ctx)
	ctx = logging.WithLogger(ctx, log.With(slog.String("request.id", "abc123")))

	logging.FromContext(ctx).Info("request started")
	doWork(ctx)

	// Output:
	// level=INFO msg="request started" request.id=abc123
	// level=INFO msg="work finished" request.id=abc123
}

func doWork(ctx context.Context) {
	logging.FromContext(ctx).Info("work finished")
}

func ExampleErr() {
	ctx := logging.WithLogger(context.Background(), exampleLogger())

	err := os.ErrNotExist
	logging.FromContext(ctx).Error("open failed", logging.Err(err))

	// Output:
	// level=ERROR msg="open failed" err="file does not exist"
}
