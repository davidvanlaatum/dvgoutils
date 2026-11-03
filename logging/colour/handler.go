// Package colour adds ANSI colours to the standard slog text handler.
// Removing its colour escape sequences produces the standard text output.
package colour

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// Mode controls when a handler uses colour.
type Mode uint8

const (
	// Auto enables colour for terminals unless NO_COLOR is non-empty or TERM is dumb.
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
// The writer must support ANSI escapes when colour is enabled.
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
		enabled = autoColour(out, options.LookupEnv, term.IsTerminal)
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

func autoColour(out io.Writer, lookupEnv func(string) (string, bool), isTerminal func(int) bool) bool {
	if value, _ := lookupEnv("NO_COLOR"); value != "" {
		return false
	}
	if value, _ := lookupEnv("TERM"); value == "dumb" {
		return false
	}
	file, ok := out.(interface{ Fd() uintptr })
	return ok && isTerminal(int(file.Fd()))
}
