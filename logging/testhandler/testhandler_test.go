package testhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"text/template"
	"time"

	"github.com/davidvanlaatum/dvgoutils/logging"
	"github.com/stretchr/testify/require"
)

func setMapPath(m map[string]any, path []string, value any) {
	if len(path) == 1 {
		m[path[0]] = value
		return
	}
	sub, ok := m[path[0]].(map[string]any)
	if !ok {
		sub = make(map[string]any)
		m[path[0]] = sub
	}
	setMapPath(sub, path[1:], value)
}

func (l *LogRecord) attrMap() map[string]any {
	m := make(map[string]any)
	s := l.String()
	if l.Source != "" {
		ss := strings.SplitN(s, ": ", 2)
		if len(ss) == 2 {
			s = ss[1]
			m[slog.SourceKey] = ss[0]
		}
	}
	for {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			break
		}
		d := json.NewDecoder(strings.NewReader(p[1]))
		var v any
		if err := d.Decode(&v); err != nil {
			panic(err)
		}
		s = strings.TrimLeft(p[1][d.InputOffset():], " ")
		setMapPath(m, strings.Split(p[0], "."), v)
	}
	return m
}

type WriterFunc func(p []byte) (n int, err error)

func (f WriterFunc) Write(p []byte) (n int, err error) { return f(p) }

type DummyTB struct {
	testing.TB
	logs []string
}

func (d *DummyTB) Log(args ...any) {
	d.Helper()
	d.TB.Log(args...)
	d.logs = append(d.logs, fmt.Sprint(args...))
}

func (d *DummyTB) Output() io.Writer {
	d.Helper()
	return WriterFunc(
		func(p []byte) (n int, err error) {
			d.Helper()
			d.logs = append(d.logs, strings.TrimRight(string(p), "\r\n"))
			return 0, fmt.Errorf("output is not supported in DummyTB")
		},
	)
}

var _ testing.TB = (*DummyTB)(nil)

func TestEmptyWithGroup(t *testing.T) {
	r := require.New(t)
	h := NewTestHandler(t)
	r.Same(h, h.WithGroup(""))
}

var expectedLogs = map[string]string{
	"built-ins":                 `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="message"`,
	"attrs":                     `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="message" k="v"`,
	"empty-attr":                `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" c="d"`,
	"zero-time":                 `testing/slogtest/slogtest.go:{{.line}}: level="INFO" msg="msg" k="v"`,
	"WithAttrs":                 `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" k="v"`,
	"groups":                    `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" G.c="d" e="f"`,
	"empty-group":               `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" e="f"`,
	"inline-group":              `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" c="d" e="f"`,
	"WithGroup":                 `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" G.a="b"`,
	"multi-With":                `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" G.c="d" G.H.e="f"`,
	"empty-group-record":        `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" G.c="d"`,
	"nested-empty-group-record": `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" a="b" G.c="d"`,
	"resolve":                   `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" k="replaced"`,
	"resolve-groups":            `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" G.a="v1" G.b="v2"`,
	"resolve-WithAttrs":         `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" k="replaced"`,
	"resolve-WithAttrs-groups":  `testing/slogtest/slogtest.go:{{.line}}: time="{{.time}}" level="INFO" msg="msg" G.a="v1" G.b="v2"`,
	"empty-PC":                  `time="{{.time}}" level="INFO" msg="message"`,
}

func TestSlogHandler(t *testing.T) {
	var h *TestHandler
	var d *DummyTB
	slogtest.Run(
		t, func(t *testing.T) slog.Handler {
			d = &DummyTB{TB: t}
			h = NewTestHandler(d)
			return h
		}, func(t *testing.T) map[string]any {
			l := h.Logs()
			r := require.New(t)
			r.Len(l, 1, "expected exactly one log entry")
			temp := template.Must(template.New("").Parse(expectedLogs[strings.SplitN(t.Name(), "/", 2)[1]]))
			b := &strings.Builder{}
			data := map[string]string{
				"time": h.logs.logs[0].Time.Format(time.RFC3339Nano),
			}
			if l[0].Source != "" {
				data["line"] = strings.SplitN(l[0].Source, ":", 2)[1]
			}
			r.NoError(temp.Execute(b, data))
			r.Len(d.logs, 1, "expected exactly one log entry in DummyTB")
			// just so we can see the result in logs and see if It's clickable in IDEs etc
			_, _ = t.Output().Write([]byte(d.logs[0] + "\n"))
			r.Equal(b.String(), d.logs[0])
			a := l[0].attrMap()
			return a
		},
	)
}

type errorOnJSONMarshal struct {
}

func (e *errorOnJSONMarshal) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal error")
}

func TestSlogHandler_MarshalError(t *testing.T) {
	r := require.New(t)
	d := &DummyTB{TB: t}
	h := NewTestHandler(d)
	logger := slog.New(h)
	_, e := json.Marshal(&errorOnJSONMarshal{})
	r.PanicsWithError(
		e.Error(), func() {
			logger.Info("msg", slog.Any("k", &errorOnJSONMarshal{}))
		},
	)
}

func TestTBFromContext(t *testing.T) {
	r := require.New(t)
	h := NewTestHandler(t)
	ctx := logging.WithLogger(t.Context(), slog.New(h))
	tb := TBFromContext(ctx)
	r.Same(t, tb)
}

func TestTBFromContextNotTestHandler(t *testing.T) {
	r := require.New(t)
	h := slog.NewTextHandler(&bytes.Buffer{}, nil)
	ctx := logging.WithLogger(t.Context(), slog.New(h))
	r.PanicsWithValue(
		"logger handler is not a TestHandler",
		func() {
			TBFromContext(ctx)
		},
	)
}

func TestTBFromContextSetup(t *testing.T) {
	r := require.New(t)
	ctx, _, _ := SetupTestHandler(t)
	r.Same(t, TBFromContext(ctx))
}

func TestSetupTestHandler(t *testing.T) {
	r := require.New(t)
	ctx, h, log := SetupTestHandler(t)
	r.NotNil(ctx)
	r.NotNil(h)
	r.NotNil(log)
	r.NotSame(t.Context(), ctx)
	r.Same(h, log.Handler())
	r.Same(log, logging.FromContext(ctx))
}

func TestWithHandlerWrapper(t *testing.T) {
	type mockHandlerWrapper struct {
		slog.Handler
	}
	r := require.New(t)
	called := false
	ctx, h, log := SetupTestHandler(
		t, WithHandlerWrapper(
			func(next slog.Handler) slog.Handler {
				called = true
				return &mockHandlerWrapper{next}
			},
		),
	)
	r.NotNil(ctx)
	r.NotNil(h)
	r.NotNil(log)
	r.NotSame(t.Context(), ctx)
	r.NotSame(h, log.Handler())
	r.IsType(&mockHandlerWrapper{}, log.Handler())
	r.Same(h, log.Handler().(*mockHandlerWrapper).Handler)
	r.Same(log, logging.FromContext(ctx))
	r.True(called, "expected handler wrapper to be called")
}

func TestWithHandlerWrapperMultiple(t *testing.T) {
	type mockHandlerWrapper struct {
		slog.Handler
	}
	type mockHandlerWrapper2 struct {
		slog.Handler
	}
	r := require.New(t)
	called := false
	called2 := false
	ctx, h, log := SetupTestHandler(
		t, WithHandlerWrapper(
			func(next slog.Handler) slog.Handler {
				called = true
				return &mockHandlerWrapper{next}
			},
		), WithHandlerWrapper(
			func(next slog.Handler) slog.Handler {
				called2 = true
				return &mockHandlerWrapper2{next}
			},
		),
	)
	r.NotNil(ctx)
	r.NotNil(h)
	r.NotNil(log)
	r.NotSame(t.Context(), ctx)
	r.NotSame(h, log.Handler())
	r.IsType(&mockHandlerWrapper2{}, log.Handler())
	r.IsType(&mockHandlerWrapper{}, log.Handler().(*mockHandlerWrapper2).Handler)
	r.Same(h, log.Handler().(*mockHandlerWrapper2).Handler.(*mockHandlerWrapper).Handler)
	r.Same(log, logging.FromContext(ctx))
	r.True(called, "expected handler wrapper to be called")
	r.True(called2, "expected handler wrapper to be called2")
}

func TestFindAllMatchingLogsForAssert(t *testing.T) {
	r := require.New(t)
	h := NewTestHandler(t)
	logger := slog.New(h)

	logger.Info("first", slog.String("kind", "keep"), slog.Int("count", 1))
	logger.Info("second", slog.String("kind", "skip"), slog.Int("count", 2))
	logger.Warn("third", slog.String("kind", "keep"), slog.Duration("delay", 2*time.Second))

	logs := h.FindAllMatchingLogsForAssert(
		func(record LogRecord) bool {
			for _, attr := range record.Attrs {
				if attr.Key == "kind" && attr.Value.String() == "keep" {
					return true
				}
			}
			return false
		},
	)

	r.Equal(
		[]map[string]any{
			{
				slog.MessageKey: "first",
				slog.LevelKey:   slog.LevelInfo.String(),
				"kind":          "keep",
				"count":         int64(1),
			},
			{
				slog.MessageKey: "third",
				slog.LevelKey:   slog.LevelWarn.String(),
				"kind":          "keep",
				"delay":         "2s",
			},
		},
		logs,
	)
}

func TestFindAllMatchingLogsForAssertHandlesSimilarAttrs(t *testing.T) {
	r := require.New(t)
	h := NewTestHandler(t)
	logger := slog.New(h)

	logger.Info(
		"nested-empty",
		slog.Group("request", slog.String("id", "target"), slog.String("meta.trace", "")),
		slog.Any("optional", nil),
	)
	logger.Info(
		"nested-filled",
		slog.Group("request", slog.String("id", "target"), slog.String("meta.trace", "present")),
		slog.Any("optional", nil),
	)
	logger.Info(
		"wrong-id",
		slog.Group("request", slog.String("id", "other"), slog.String("meta.trace", "")),
		slog.Any("optional", nil),
	)

	logs := h.FindAllMatchingLogsForAssert(
		func(record LogRecord) bool {
			values := record.ToMapForAssert()
			requestID, ok := values["request.id"].(string)
			if !ok || requestID != "target" {
				return false
			}
			trace, ok := values["request.meta.trace"].(string)
			if !ok || trace != "" {
				return false
			}
			optional, ok := values["optional"]
			return ok && optional == nil
		},
	)

	r.Equal(
		[]map[string]any{
			{
				slog.MessageKey:      "nested-empty",
				slog.LevelKey:        slog.LevelInfo.String(),
				"request.id":         "target",
				"request.meta.trace": "",
				"optional":           nil,
			},
		},
		logs,
	)
}

func TestLogRecordString(t *testing.T) {
	r := require.New(t)
	l := &LogRecord{
		Level: slog.LevelInfo,
		Msg:   "msg",
		Attrs: []slog.Attr{
			slog.Duration("duration", time.Second),
			slog.Time("time", time.Date(2024, time.June, 1, 12, 0, 0, 0, time.UTC)),
			slog.Int("int", 42),
			slog.String("string", "hello"),
			slog.String("url", "http://example.com?foo=bar&baz=qux"),
		},
	}
	r.Equal(
		`level="INFO" msg="msg" duration="1s" time="2024-06-01T12:00:00Z" int=42 string="hello" url="http://example.com?foo=bar&baz=qux"`,
		l.String(),
	)
}

func TestFirstMatchingLogForAssertConvertsToMap(t *testing.T) {
	r := require.New(t)
	_, h, log := SetupTestHandler(t)
	log.Info("msg", slog.String("k", "v"), slog.Duration("duration", time.Second), slog.Int("int", 42))
	r.Equal(
		map[string]any{
			"msg":      "msg",
			"level":    "INFO",
			"k":        "v",
			"int":      int64(42),
			"duration": time.Second.String(),
		}, h.FirstMatchingLogForAssert(
			func(record LogRecord) bool {
				return true
			},
		),
	)
}

func TestFirstMatchingLogForAssertFindsCorrectLog(t *testing.T) {
	r := require.New(t)
	_, h, log := SetupTestHandler(t)
	log.Info("log1")
	log.Info("log2")
	log.Info("log3")
	r.Equal(
		map[string]any{
			"msg":   "log2",
			"level": "INFO",
		}, h.FirstMatchingLogForAssert(
			func(record LogRecord) bool {
				return record.Msg == "log2"
			},
		),
	)
}

func TestFirstMatchingLogForAssertReturnsNilIfNoLogFound(t *testing.T) {
	r := require.New(t)
	_, h, log := SetupTestHandler(t)
	log.Info("log1")
	log.Info("log2")
	log.Info("log3")
	r.Nil(
		h.FirstMatchingLogForAssert(
			func(record LogRecord) bool {
				return record.Msg == "log4"
			},
		),
	)
}

func TestFirstMatchingLogForAssertHandlesSimilarAttrs(t *testing.T) {
	r := require.New(t)
	_, h, log := SetupTestHandler(t)

	log.Info(
		"wrong-empty",
		slog.Group("payload", slog.String("id", "target"), slog.String("value", "")),
		slog.Any("optional", "set"),
	)
	log.Info(
		"match",
		slog.Group("payload", slog.String("id", "target"), slog.String("value", "")),
		slog.Any("optional", nil),
	)
	log.Info(
		"wrong-nested",
		slog.Group("payload", slog.String("id", "target"), slog.String("value", "non-empty")),
		slog.Any("optional", nil),
	)

	match := h.FirstMatchingLogForAssert(
		func(record LogRecord) bool {
			values := record.ToMapForAssert()
			payloadID, ok := values["payload.id"].(string)
			if !ok || payloadID != "target" {
				return false
			}
			payloadValue, ok := values["payload.value"].(string)
			if !ok || payloadValue != "" {
				return false
			}
			optional, ok := values["optional"]
			return ok && optional == nil
		},
	)

	r.Equal(
		map[string]any{
			slog.MessageKey: "match",
			slog.LevelKey:   slog.LevelInfo.String(),
			"payload.id":    "target",
			"payload.value": "",
			"optional":      nil,
		},
		match,
	)
}
