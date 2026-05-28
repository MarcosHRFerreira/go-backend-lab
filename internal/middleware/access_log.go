package middleware

import (
	"log/slog"
	"time"

	"go-tweets/internal/observability/logctx"

	"github.com/gin-gonic/gin"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		requestLogger := logctx.FromContext(c.Request.Context())
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		attrs := []slog.Attr{
			slog.String("route", route),
			slog.Int("status_code", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(startedAt).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if userID, exists := c.Get("userID"); exists {
			attrs = append(attrs, slog.Any("user_id", userID))
		}

		message := "http request completed"
		switch statusCode := c.Writer.Status(); {
		case statusCode >= 500:
			requestLogger.Error(message, attrsToArgs(attrs)...)
		case statusCode >= 400:
			requestLogger.Warn(message, attrsToArgs(attrs)...)
		default:
			requestLogger.Info(message, attrsToArgs(attrs)...)
		}
	}
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}

	return args
}
