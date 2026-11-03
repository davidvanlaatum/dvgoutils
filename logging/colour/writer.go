package colour

import (
	"io"
	"log/slog"
	"strings"
)

// colourLogWriter decorates complete TextHandler records. Keeping the standard
// handler responsible for serialisation preserves escaping, groups and LogValuer
// behaviour, including attributes attached through Logger.With.
// TextHandler serialises writes across its derived handlers.
type colourLogWriter struct {
	out io.Writer
}

func (w colourLogWriter) Write(p []byte) (int, error) {
	var line strings.Builder
	// TextHandler quotes keys and values containing spaces or '='. Only split
	// fields and key/value separators outside quotes, respecting escaped quotes.
	start, separator := 0, -1
	quoted, escaped := false, false
	for i, c := range p {
		if escaped {
			escaped = false
			continue
		}
		if quoted && c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			quoted = !quoted
		}
		if quoted {
			continue
		}
		if c == '=' && separator < 0 {
			separator = i
		}
		if c != ' ' && c != '\n' {
			continue
		}
		if separator >= start {
			key := string(p[start:separator])
			valueColour := ""
			if key == slog.MessageKey {
				valueColour = "\x1b[32m"
			}
			if key == slog.LevelKey {
				valueColour = logLevelColour(string(p[separator+1 : i]))
			}
			line.WriteString("\x1b[36m")
			line.Write(p[start:separator])
			line.WriteString("\x1b[0m=")
			line.WriteString(valueColour)
			line.Write(p[separator+1 : i])
			line.WriteString("\x1b[0m")
		} else {
			line.Write(p[start:i])
		}
		line.WriteByte(c)
		start, separator = i+1, -1
	}
	line.Write(p[start:])
	n, err := io.WriteString(w.out, line.String())
	if err != nil {
		return 0, err
	}
	if n != line.Len() {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func logLevelColour(level string) string {
	switch {
	case strings.HasPrefix(level, "DEBUG"):
		return "\x1b[90m"
	case strings.HasPrefix(level, "WARN"):
		return "\x1b[33m"
	case strings.HasPrefix(level, "ERROR"):
		return "\x1b[31m"
	default:
		return "\x1b[34m"
	}
}
