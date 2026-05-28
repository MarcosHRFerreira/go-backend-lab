// Package logctx stores request-scoped observability data in contexts.
package logctx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const RequestIDHeader = "X-Request-ID"

type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}

	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	baseLogger := slog.Default()
	if ctx != nil {
		if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
			baseLogger = logger
		} else if requestID := RequestID(ctx); requestID != "" {
			baseLogger = baseLogger.With(slog.String("request_id", requestID))
		}
	}

	if traceID := TraceID(ctx); traceID != "" {
		return baseLogger.With(slog.String("trace_id", traceID))
	}

	return baseLogger
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return ""
	}

	return spanContext.TraceID().String()
}
