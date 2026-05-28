package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"go-tweets/internal/observability/logctx"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const TraceIDHeader = "X-Trace-ID"

func Trace(tracer trace.Tracer) gin.HandlerFunc {
	if tracer == nil {
		tracer = otel.Tracer("go-tweets/http")
	}

	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		parentContext := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		spanName := fmt.Sprintf("%s %s", c.Request.Method, route)
		ctx, span := tracer.Start(parentContext, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		requestLogger := logctx.FromContext(c.Request.Context())
		spanContext := span.SpanContext()
		if spanContext.IsValid() {
			ctx = logctx.WithLogger(ctx, requestLogger.With(
				slog.String("trace_id", spanContext.TraceID().String()),
				slog.String("span_id", spanContext.SpanID().String()),
			))
			c.Header(TraceIDHeader, spanContext.TraceID().String())
		} else {
			ctx = logctx.WithLogger(ctx, requestLogger)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		statusCode := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))
		if statusCode >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(statusCode))
		}
	}
}
