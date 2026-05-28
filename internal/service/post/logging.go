package post

import (
	"context"
	"log/slog"

	"go-tweets/internal/observability/logctx"
)

func serviceLogger(ctx context.Context, operation string) *slog.Logger {
	return logctx.FromContext(ctx).With(
		slog.String("component", "post_service"),
		slog.String("operation", operation),
	)
}
