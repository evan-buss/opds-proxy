package reqctx

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	requestLoggerKey  = contextKey("requestLogger")
	isLocalRequestKey = contextKey("isLocalRequest")
)

func WithRequestLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, requestLoggerKey, log)
}

func Logger(ctx context.Context) *slog.Logger {
	if v := ctx.Value(requestLoggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}

func WithIsLocal(ctx context.Context, isLocal bool) context.Context {
	return context.WithValue(ctx, isLocalRequestKey, isLocal)
}

func IsLocal(ctx context.Context) bool {
	if v := ctx.Value(isLocalRequestKey); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
