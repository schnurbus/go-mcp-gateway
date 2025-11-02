package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type contextKey struct{}

var key contextKey
var RequestIDKey contextKey

func NewLogger(w ...io.Writer) *slog.Logger {
	level := parseLevel(strings.ToUpper(os.Getenv("LOG_LEVEL")))
	env := strings.ToUpper(os.Getenv("GO_ENV"))

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	out := io.MultiWriter(append([]io.Writer{os.Stdout}, w...)...)
	if env == "PRODUCTION" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(key).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}

func parseLevel(s string) slog.Level {
	switch s {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
