package testhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davidvanlaatum/dvgoutils"
	"github.com/davidvanlaatum/dvgoutils/logging"
)

type LogRecord struct {
	Time   time.Time
	Level  slog.Level
	Msg    string
	Source string
	Attrs  []slog.Attr
}

func (l *LogRecord) String() string {
	var a []slog.Attr
	if !l.Time.IsZero() {
		a = append(a, slog.Time(slog.TimeKey, l.Time))
	}
	a = append(
		a,
		slog.String(slog.LevelKey, l.Level.String()),
		slog.String(slog.MessageKey, l.Msg),
	)
	a = append(a, l.Attrs...)
	rt := strings.Join(
		dvgoutils.MapSlice(
			a, func(a slog.Attr) string {
				v := &bytes.Buffer{}
				e := json.NewEncoder(v)
				value := a.Value.Resolve()
				var vv any
				switch value.Kind() {
				case slog.KindDuration:
					vv = value.Duration().String()
				default:
					vv = value.Any()
				}
				if err := e.Encode(vv); err != nil {
					panic(err)
				}
				for v.Len() > 0 && v.Bytes()[v.Len()-1] == '\n' {
					v.Truncate(v.Len() - 1)
				}
				return fmt.Sprintf("%s=%s", a.Key, v.String())
			},
		), " ",
	)
	if l.Source != "" {
		return l.Source + ": " + rt
	}
	return rt
}

type logsHolder struct {
	logs []LogRecord
	mu   sync.Mutex
}

type TestHandler struct {
	T     testing.TB
	attr  []slog.Attr
	group string
	logs  *logsHolder
}

func NewTestHandler(t testing.TB) *TestHandler {
	return &TestHandler{T: t, logs: &logsHolder{}}
}

func (t *TestHandler) Logs() []LogRecord {
	t.logs.mu.Lock()
	defer t.logs.mu.Unlock()
	return t.logs.logs
}

func (t *TestHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func expandGroup(prefix string, attrs []slog.Attr) (res []slog.Attr) {
	res = make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		if g, ok := a.Value.Any().([]slog.Attr); ok {
			subPrefix := prefix
			if a.Key != "" {
				subPrefix += a.Key + "."
			}
			res = append(res, expandGroup(subPrefix, g)...)
		} else {
			a.Key = prefix + a.Key
			res = append(res, a)
		}
	}
	return
}

func (t *TestHandler) Handle(_ context.Context, record slog.Record) error {
	t.T.Helper()
	r := &LogRecord{
		Time:  record.Time,
		Level: record.Level,
		Msg:   record.Message,
	}
	if record.PC != 0 {
		c, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
		r.Source = fmt.Sprintf("%s:%d", c.File, c.Line)
	}
	r.Attrs = make([]slog.Attr, 0, len(t.attr)+record.NumAttrs())
	r.Attrs = append(r.Attrs, t.attr...)
	var newAttr []slog.Attr
	record.Attrs(
		func(attr slog.Attr) bool {
			newAttr = append(newAttr, attr)
			return true
		},
	)
	r.Attrs = append(r.Attrs, dvgoutils.FilterSlice(expandGroup(t.group, newAttr), dropEmptyKey)...)
	func() {
		t.logs.mu.Lock()
		defer t.logs.mu.Unlock()
		t.logs.logs = append(t.logs.logs, *r)
	}()
	if _, err := t.T.Output().Write([]byte(r.String() + "\n")); err != nil {
		return err
	}
	return nil
}

func dropEmptyKey(a slog.Attr) bool {
	return a.Key != ""
}

func (t *TestHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	a := slices.Concat(t.attr, dvgoutils.FilterSlice(expandGroup(t.group, attrs), dropEmptyKey))
	return &TestHandler{
		T:     t.T,
		attr:  a,
		group: t.group,
		logs:  t.logs,
	}
}

func (t *TestHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return t
	}
	return &TestHandler{
		T:     t.T,
		attr:  t.attr,
		group: t.group + name + ".",
		logs:  t.logs,
	}
}

var _ slog.Handler = (*TestHandler)(nil)

type contextKey struct{}

func TBFromContext(ctx context.Context) testing.TB {
	if t, ok := ctx.Value(contextKey{}).(testing.TB); ok {
		return t
	}
	logger := logging.FromContext(ctx)
	if testhandler, ok := logger.Handler().(*TestHandler); ok {
		return testhandler.T
	}
	panic("logger handler is not a TestHandler")
}

type HandlerWrapperFunc func(slog.Handler) slog.Handler

type setupTestHandlerConfig struct {
	handlerWrappers []HandlerWrapperFunc
}

type SetupOption func(*setupTestHandlerConfig)

func WithHandlerWrapper(f HandlerWrapperFunc) SetupOption {
	return func(cfg *setupTestHandlerConfig) {
		cfg.handlerWrappers = append(cfg.handlerWrappers, f)
	}
}

// WithRuntime attaches a TestRuntimeWrapper to the handler, which adds an attribute with the time since the wrapper was
// created to each log record. If a start time is provided, it uses that instead of the current time. This can be used
// to measure the runtime of code being tested. The attribute key is "test-runtime" and the value is a string
// representation of the duration since the start time.
func WithRuntime(startTime ...time.Time) SetupOption {
	if len(startTime) > 1 {
		panic("WithRuntime accepts at most one startTime argument")
	}
	return WithHandlerWrapper(
		func(h slog.Handler) slog.Handler {
			start := time.Now()
			if len(startTime) > 0 {
				start = startTime[0]
			}
			return NewTestRuntimeWrapper(start, h)
		},
	)
}

func SetupTestHandler(t testing.TB, opts ...SetupOption) (
	ctx context.Context,
	handler *TestHandler,
	logger *slog.Logger,
) {
	var cfg setupTestHandlerConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	handler = NewTestHandler(t)
	var h slog.Handler = handler
	for _, w := range cfg.handlerWrappers {
		h = w(h)
	}
	logger = slog.New(h)
	ctx = logging.WithLogger(t.Context(), logger)
	ctx = context.WithValue(ctx, contextKey{}, t)
	return
}
