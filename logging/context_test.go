package logging

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	r := require.New(t)
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	r.Same(l, FromContext(WithLogger(context.Background(), l)))
}

func TestNoLoggerInContext(t *testing.T) {
	r := require.New(t)
	r.Panics(func() { FromContext(context.Background()) })
}
