package testhandler_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/davidvanlaatum/dvgoutils/logging"
	"github.com/davidvanlaatum/dvgoutils/logging/testhandler"
	"github.com/stretchr/testify/require"
)

func TestSetupTestHandlerUsage(t *testing.T) {
	r := require.New(t)
	ctx, handler, logger := testhandler.SetupTestHandler(t)
	ctx = logging.WithLogger(ctx, logger.With(slog.String("request.id", "abc123")))

	queueImport(ctx)

	r.Equal(
		map[string]any{
			slog.MessageKey: "queued",
			slog.LevelKey:   slog.LevelInfo.String(),
			"request.id":    "abc123",
			"job":           "import",
			"count":         int64(3),
		},
		handler.FirstMatchingLogForAssert(func(record testhandler.LogRecord) bool {
			return record.Msg == "queued"
		}),
	)
}

func queueImport(ctx context.Context) {
	logging.FromContext(ctx).Info(
		"queued",
		slog.String("job", "import"),
		slog.Int("count", 3),
	)
}

func ExampleLogRecord_ToMapForAssert() {
	record := testhandler.LogRecord{
		Level: slog.LevelInfo,
		Msg:   "queued",
		Attrs: []slog.Attr{
			slog.String("request.id", "abc123"),
			slog.Int("count", 3),
		},
	}

	values := record.ToMapForAssert()

	fmt.Println(values[slog.MessageKey])
	fmt.Println(values[slog.LevelKey])
	fmt.Println(values["request.id"])
	fmt.Println(values["count"])

	// Output:
	// queued
	// INFO
	// abc123
	// 3
}
