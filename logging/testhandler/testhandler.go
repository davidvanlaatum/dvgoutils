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
	"testing"
	"time"

	"github.com/davidvanlaatum/dvgoutils"
)

type LogRecord struct {
	Time     time.Time
	Level    slog.Level
	Msg      string
	Location string
	Attrs    []slog.Attr
}

func (l *LogRecord) String() string {
	var a []slog.Attr
	if !l.Time.IsZero() {
		a = append(a, slog.Time(slog.TimeKey, l.Time))
	}
	if l.Location != "" {
		a = append(a, slog.String("location", l.Location))
	}
	a = append(a,
		slog.String(slog.LevelKey, l.Level.String()),
		slog.String(slog.MessageKey, l.Msg),
	)
	a = append(a, l.Attrs...)
	return strings.Join(dvgoutils.MapSlice(a, func(a slog.Attr) string {
		v := &bytes.Buffer{}
		e := json.NewEncoder(v)
		if err := e.Encode(a.Value.Resolve().Any()); err != nil {
			panic(err)
		}
		for v.Len() > 0 && v.Bytes()[v.Len()-1] == '\n' {
			v.Truncate(v.Len() - 1)
		}
		return fmt.Sprintf("%s=%s", a.Key, v.String())
	}), " ")
}

type logsHolder struct {
	logs []LogRecord
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
		pkg := strings.SplitN(c.Function, ".", 2)[0]
		file := c.File
		if idx := strings.LastIndex(file, "/"+pkg+"/"); idx >= 0 {
			file = file[idx+1:]
		}
		r.Location = fmt.Sprintf("%s:%d", file, c.Line)
	}
	r.Attrs = make([]slog.Attr, 0, len(t.attr)+record.NumAttrs())
	r.Attrs = append(r.Attrs, t.attr...)
	var newAttr []slog.Attr
	record.Attrs(func(attr slog.Attr) bool {
		newAttr = append(newAttr, attr)
		return true
	})
	r.Attrs = append(r.Attrs, dvgoutils.FilterSlice(expandGroup(t.group, newAttr), dropEmptyKey)...)
	t.logs.logs = append(t.logs.logs, *r)
	t.T.Log(r.String())
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
