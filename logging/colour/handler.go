// Package colour adds ANSI colours to the standard slog text handler.
// Removing its colour escape sequences produces the standard text output.
package colour

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Mode controls when a handler uses colour.
type Mode uint8

const (
	// Auto enables colour for terminals unless NO_COLOR is non-empty or TERM is dumb.
	// On Windows, it enables VT output processing or falls back to plain text.
	Auto Mode = iota
	// Always enables colour, including for redirected output and when NO_COLOR is set.
	Always
	// Never disables colour.
	Never
)

// ParseMode parses auto, always or never, for example from a command-line flag.
// An empty string also selects Auto. Values are case-sensitive.
func ParseMode(value string) (Mode, error) {
	switch value {
	case "", "auto":
		return Auto, nil
	case "always":
		return Always, nil
	case "never":
		return Never, nil
	default:
		return Auto, fmt.Errorf("invalid colour mode %q: must be auto, always or never", value)
	}
}

// Options configures a handler. The zero value uses automatic colour detection
// and the standard slog.HandlerOptions defaults (including level Info).
type Options struct {
	slog.HandlerOptions
	Mode Mode
	// LookupEnv defaults to os.LookupEnv. It is consulted only in Auto mode,
	// once when the handler is constructed.
	LookupEnv func(string) (string, bool)
}

// NewHandler returns a standard text handler with optional ANSI colouring.
// A nil opts uses the defaults. It does not change slog's default logger.
// Auto detects writers implementing Fd() uintptr; other writers use plain text.
// On Windows, Auto enables processed output and VT processing, preserving other
// console mode flags. If reading or setting the mode fails, it uses plain text.
// These console flags remain enabled for other users of the same screen buffer.
// Always emits ANSI escapes without preparing the console; the caller must
// provide an output destination that supports them. Never does not alter console modes.
// NewHandler panics if opts.Mode is not Auto, Always or Never.
func NewHandler(out io.Writer, opts *Options) *slog.TextHandler {
	var options Options
	if opts != nil {
		options = *opts
	}
	if options.LookupEnv == nil {
		options.LookupEnv = os.LookupEnv
	}
	var enabled bool
	switch options.Mode {
	case Auto:
		enabled = autoColour(out, options.LookupEnv, prepareTerminal)
	case Always:
		enabled = true
	case Never:
	default:
		panic(fmt.Sprintf("invalid colour mode %d", options.Mode))
	}
	if enabled {
		out = colourLogWriter{out: out}
	}
	return slog.NewTextHandler(out, &options.HandlerOptions)
}

func autoColour(out io.Writer, lookupEnv func(string) (string, bool), prepareTerminal func(uintptr) bool) bool {
	if value, _ := lookupEnv("NO_COLOR"); value != "" {
		return false
	}
	if value, _ := lookupEnv("TERM"); value == "dumb" {
		return false
	}
	file, ok := out.(interface{ Fd() uintptr })
	return ok && prepareTerminal(file.Fd())
}
