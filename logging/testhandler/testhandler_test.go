package testhandler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"text/template"
	"time"

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

type DummyTB struct {
	testing.TB
	logs []string
}

func (d *DummyTB) Log(args ...any) {
	d.Helper()
	d.TB.Log(args...)
	d.logs = append(d.logs, fmt.Sprint(args...))
}

var _ testing.TB = (*DummyTB)(nil)

func TestEmptyWithGroup(t *testing.T) {
	r := require.New(t)
	h := NewTestHandler(t)
	r.Same(h, h.WithGroup(""))
}

var expectedLogs = map[string]string{
	"built-ins":                 `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="message"`,
	"attrs":                     `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="message" k="v"`,
	"empty-attr":                `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" c="d"`,
	"zero-time":                 `location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" k="v"`,
	"WithAttrs":                 `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" k="v"`,
	"groups":                    `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" G.c="d" e="f"`,
	"empty-group":               `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" e="f"`,
	"inline-group":              `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" c="d" e="f"`,
	"WithGroup":                 `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" G.a="b"`,
	"multi-With":                `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" G.c="d" G.H.e="f"`,
	"empty-group-record":        `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" G.c="d"`,
	"nested-empty-group-record": `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" a="b" G.c="d"`,
	"resolve":                   `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" k="replaced"`,
	"resolve-groups":            `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" G.a="v1" G.b="v2"`,
	"resolve-WithAttrs":         `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" k="replaced"`,
	"resolve-WithAttrs-groups":  `time="{{.time}}" location="testing/slogtest/slogtest.go:{{.line}}" level="INFO" msg="msg" G.a="v1" G.b="v2"`,
	"empty-PC":                  `time="{{.time}}" level="INFO" msg="message"`,
}

func TestSlogHandler(t *testing.T) {
	var h *TestHandler
	var d *DummyTB
	slogtest.Run(t, func(t *testing.T) slog.Handler {
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
		if l[0].Location != "" {
			data["line"] = strings.SplitN(l[0].Location, ":", 2)[1]
		}
		r.NoError(temp.Execute(b, data))
		r.Equal(b.String(), d.logs[0])
		a := l[0].attrMap()
		return a
	})
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
	r.PanicsWithError(e.Error(), func() {
		logger.Info("msg", slog.Any("k", &errorOnJSONMarshal{}))
	})
}
