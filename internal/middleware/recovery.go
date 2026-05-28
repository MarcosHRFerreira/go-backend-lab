package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"go-tweets/internal/httpresponse"
	"go-tweets/internal/observability/logctx"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recoveredValue := recover()
			if recoveredValue == nil {
				return
			}

			requestLogger := logctx.FromContext(c.Request.Context())
			attrs := []any{
				slog.String("route", routeOrPath(c)),
				slog.String("panic", fmt.Sprint(recoveredValue)),
				slog.String("stack", string(debug.Stack())),
			}

			requestLogger.Error("panic recovered", attrs...)
			httpresponse.AbortJSONError(c, http.StatusInternalServerError, "internal server error")
		}()

		c.Next()
	}
}

func routeOrPath(c *gin.Context) string {
	route := c.FullPath()
	if route != "" {
		return route
	}

	return c.Request.URL.Path
}
