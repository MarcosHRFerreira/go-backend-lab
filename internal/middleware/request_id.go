package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"go-tweets/internal/observability/logctx"

	"github.com/gin-gonic/gin"
)

const requestIDBytes = 12

func RequestID(logger *slog.Logger) gin.HandlerFunc {
	baseLogger := logger
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(logctx.RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		requestLogger := baseLogger.With(
			slog.String("component", "http"),
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		ctx := logctx.WithRequestID(c.Request.Context(), requestID)
		ctx = logctx.WithLogger(ctx, requestLogger)

		c.Header(logctx.RequestIDHeader, requestID)
		c.Set("requestID", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func generateRequestID() string {
	randomBytes := make([]byte, requestIDBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(randomBytes)
}
