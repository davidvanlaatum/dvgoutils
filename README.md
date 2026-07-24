![badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/davidvanlaatum/c77e2d5c6afe0044c618cf1208e6d78a/raw/coverage.json)
[![Go](https://github.com/davidvanlaatum/dvgoutils/actions/workflows/go.yml/badge.svg)](https://github.com/davidvanlaatum/dvgoutils/actions/workflows/go.yml)

# dvgoutils

Small Go utility packages for generic helpers, context-aware `slog` logging, test log assertions, and bit/byte units.

## Installation

```sh
go get github.com/davidvanlaatum/dvgoutils
```

Import only the packages you need:

```go
import (
	"github.com/davidvanlaatum/dvgoutils"
	"github.com/davidvanlaatum/dvgoutils/logging"
	"github.com/davidvanlaatum/dvgoutils/logging/testhandler"
	"github.com/davidvanlaatum/dvgoutils/units"
)
```

## Generic Helpers

The root package contains small generic helpers for common operations:

```go
numbers := []int{1, 2, 3, 4}

even := dvgoutils.FilterSlice(numbers, func(v int) bool {
	return v%2 == 0
})

evenCount := dvgoutils.CountSlice(numbers, func(v int) bool {
	return v%2 == 0
})

labels := dvgoutils.MapSlice(even, func(v int) string {
	return fmt.Sprintf("value-%d", v)
})

port := dvgoutils.Ptr(8080)
count := dvgoutils.Must(strconv.Atoi("42"))
```

`Must` panics when the error argument is non-nil, so reserve it for setup paths where fail-fast behaviour is desired.

## Logging

The `logging` package stores a `*slog.Logger` in a `context.Context`. This is useful for code that already receives a context and should log without threading a logger through every call.

```go
func Run(ctx context.Context) {
	log := logging.FromContext(ctx)
	log.Info("starting job", slog.String("component", "worker"))
}

handler := slog.NewJSONHandler(os.Stdout, nil)
logger := slog.New(handler)
ctx := logging.WithLogger(context.Background(), logger)

Run(ctx)
```

`FromContext` panics if the context does not contain a logger. This keeps missing logging setup visible during development and tests.

You can also derive a logger with default attributes and put that logger back into a child context. Every function called with the child context will include those attributes whenever it logs via `logging.FromContext`.

```go
func ProcessRequest(ctx context.Context, requestID string) error {
	log := logging.FromContext(ctx)
	ctx = logging.WithLogger(ctx, log.With(slog.String("request.id", requestID)))

	logging.FromContext(ctx).Info("request started")

	return loadCustomer(ctx)
}

func loadCustomer(ctx context.Context) error {
	logging.FromContext(ctx).Info("loading customer")
	return nil
}
```

Both log records include `request.id` because `loadCustomer` receives the context containing the derived logger.

Use `logging.FromContext(ctx)` after replacing the context. Logger values captured before `logging.WithLogger` will not include attributes attached to the derived logger.

Use `logging.Err` to attach errors consistently. Nil errors return an empty attribute, which the test handler filters out when the key is empty.

```go
if err := doWork(ctx); err != nil {
	logging.FromContext(ctx).Error("work failed", logging.Err(err))
}
```

## Testing Logs

The `logging/testhandler` package provides an in-memory `slog.Handler` that also writes formatted log lines to `testing.TB.Output()`. Use `SetupTestHandler` when tested code expects a logger in context:

```go
func runImport(ctx context.Context) {
	logging.FromContext(ctx).Info(
		"queued",
		slog.String("job", "import"),
		slog.Int("count", 3),
	)
}

func TestRunLogsResult(t *testing.T) {
	r := require.New(t)
	ctx, handler, log := testhandler.SetupTestHandler(t)
	ctx = logging.WithLogger(ctx, log.With(slog.String("request.id", "abc123")))

	runImport(ctx)

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
```

`SetupTestHandler` returns:

- `ctx`: a child of `t.Context()` containing the logger and test handle.
- `handler`: the underlying `*testhandler.TestHandler` for inspecting captured logs.
- `logger`: the `*slog.Logger` stored in the context.

For direct handler use, construct one with `NewTestHandler`:

```go
handler := testhandler.NewTestHandler(t)
logger := slog.New(handler)

logger.Warn("retrying", slog.Duration("delay", 2*time.Second))

logs := handler.Logs()
require.Len(t, logs, 1)
require.Equal(t, "retrying", logs[0].Msg)
```

For assertions, prefer `FirstMatchingLogForAssert` or `FindAllMatchingLogsForAssert`. They convert matching records to stable `map[string]any` values without timestamp or source fields:

```go
matches := handler.FindAllMatchingLogsForAssert(func(record testhandler.LogRecord) bool {
	values := record.ToMapForAssert()
	return values["request.id"] == "abc123"
})
```

Groups are flattened with dot-separated keys, so `slog.Group("request", slog.String("id", "abc123"))` becomes `request.id`. Duration values are converted to strings, matching the formatted log output.

Assertion maps use `slog`'s resolved value types. For example, `slog.Int` appears as `int64`.

### Handler Wrappers

Use `WithHandlerWrapper` to wrap the test handler with extra `slog.Handler` behaviour:

```go
ctx, handler, log := testhandler.SetupTestHandler(
	t,
	testhandler.WithHandlerWrapper(func(next slog.Handler) slog.Handler {
		return myWrapper{Handler: next}
	}),
)
```

`WithRuntime` is a built-in wrapper that adds a `test-runtime` attribute to every log record:

```go
_, handler, log := testhandler.SetupTestHandler(t, testhandler.WithRuntime())

log.Info("finished")

record := handler.FirstMatchingLogForAssert(func(record testhandler.LogRecord) bool {
	return record.Msg == "finished"
})
require.Contains(t, record, "test-runtime")
```

Use `TBFromContext` when helper code needs the active `testing.TB` from a test context:

```go
func requireSomething(ctx context.Context) {
	t := testhandler.TBFromContext(ctx)
	t.Helper()

	require.NotNil(t, logging.FromContext(ctx))
}
```

`SetupTestHandler` is the safest way to create contexts intended for `TBFromContext`. If the context was not created that way, `TBFromContext` falls back to the logger handler and panics unless it is a bare `*testhandler.TestHandler`.

## Units

The `units` package provides typed bit and byte values with human-readable string formatting and `slog.LogValuer` support:

```go
size := 42 * units.MiB
rate := 100 * units.Mb
transferRate := size.PerSecond(2 * time.Second)
bitRate := rate.PerSecond(2 * time.Second)

fmt.Println(size) // 42.0 MiB
fmt.Println(rate) // 100.0 Mb
fmt.Println(transferRate) // 21.0 MiB/s
fmt.Println(bitRate) // 50.0 Mb/s

logger.Info("transfer complete", slog.Any("size", size), slog.Any("rate", transferRate))
```

`Bytes` and `BytesPerSecond` use binary units (`KiB`, `MiB`, `GiB`, ...). `Bits` and `BitsPerSecond` use decimal units (`Kb`, `Mb`, `Gb`, ...). Use `Bytes.PerSecond` or `Bits.PerSecond` with a positive elapsed duration to calculate a transfer rate. Both methods return `NaN` when the duration is zero or negative.

## Development

Run the full test suite with:

```sh
go test -trimpath ./...
```

`-trimpath` keeps source locations in test output stable across machines and Go installations.

Runnable Go examples live in `*_test.go` files and are checked by the same command.
