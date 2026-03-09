package logging

import (
	"log/slog"
)

func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.Any("err", logError{err})
}

type logError struct {
	err error
}

func (e logError) LogValue() slog.Value {
	return slog.StringValue(e.err.Error())
}
