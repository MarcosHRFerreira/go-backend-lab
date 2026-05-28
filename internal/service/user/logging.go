package user

import (
	"context"
	"log/slog"

	"go-tweets/internal/observability/logctx"
)

func serviceLogger(ctx context.Context, operation string) *slog.Logger {
	return logctx.FromContext(ctx).With(
		slog.String("component", "user_service"),
		slog.String("operation", operation),
	)
}
