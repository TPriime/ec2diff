package logger

import (
	"context"
	"log/slog"
	"os"

	slogctx "github.com/veqryn/slog-context"
)

type Level = slog.Level

const (
	LevelInfo Level = slog.LevelInfo
	LevelSilent Level = slog.LevelError
)

// Init sets up default logger
func Init(level Level) {
	handler := slogctx.NewHandler(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
		nil,
	)
	slog.SetDefault(slog.New(handler))
}

func With(ctx context.Context, keyvals ...any) context.Context {
	return slogctx.With(ctx, keyvals...)
}

func WithGroup(ctx context.Context, name string) context.Context {
	return slogctx.WithGroup(ctx, name)
}

func Logger(ctx context.Context) *slog.Logger {
	return slogctx.FromCtx(ctx)
}

// ---- Logging Wrappers ---- //

func Debug(ctx context.Context, msg string, args ...any) {
	slogctx.Debug(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slogctx.Info(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slogctx.Warn(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	slogctx.Error(ctx, msg, args...)
}
