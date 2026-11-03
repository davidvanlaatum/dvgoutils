package colour_test

import (
	"log/slog"
	"os"

	"github.com/davidvanlaatum/dvgoutils/logging/colour"
)

func ExampleNewHandler() {
	logger := slog.New(colour.NewHandler(os.Stderr, nil))
	logger.Info("starting", "component", "worker")
}

func ExampleNewHandler_options() {
	var level slog.LevelVar
	logger := slog.New(colour.NewHandler(os.Stderr, &colour.Options{
		Mode: colour.Always,
		HandlerOptions: slog.HandlerOptions{
			Level: &level,
		},
	}))
	level.Set(slog.LevelDebug)
	logger.Debug("debug logging enabled")
}
