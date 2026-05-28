// Package middleware provides Gin middlewares for authentication.
package middleware

import (
	"go-tweets/internal/httpresponse"
	"go-tweets/internal/observability/logctx"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-tweets/pkg/internalsql/jwt"
)

func AuthMiddleware(secretkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestLogger := logctx.FromContext(c.Request.Context())
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			requestLogger.Warn("authorization failed", slog.String("reason", "missing_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "missing authorization token")
			return
		}

		// Access-token validation must fail fast because downstream handlers assume an authenticated user.
		// A validacao do access token precisa falhar rapidamente porque os handlers seguintes assumem um usuario autenticado.
		userID, username, err := jwt.ValidateToken(header, secretkey, true)
		if err != nil {
			requestLogger.Warn("authorization failed", slog.String("reason", "invalid_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}
		// Store identity data in the Gin context so handlers do not need to parse the token again.
		// Armazena os dados de identidade no contexto do Gin para que os handlers nao precisem interpretar o token novamente.
		c.Set("userID", userID)
		c.Set("username", username)
		c.Request = c.Request.WithContext(logctx.WithLogger(c.Request.Context(), requestLogger.With(slog.Int("user_id", userID))))
		c.Next()

	}
}
func AuthRefreshTokenMiddleware(secretkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestLogger := logctx.FromContext(c.Request.Context())
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			requestLogger.Warn("authorization failed", slog.String("reason", "missing_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "missing authorization token")
			return
		}

		// Refresh-token validation is intentionally separated from access tokens because the semantics differ.
		// A validacao do refresh token e separada de proposito da do access token porque a semantica dos dois e diferente.
		userID, username, err := jwt.ValidateToken(header, secretkey, false)
		if err != nil {
			requestLogger.Warn("authorization failed", slog.String("reason", "invalid_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}
		c.Set("userID", userID)
		c.Set("username", username)
		c.Request = c.Request.WithContext(logctx.WithLogger(c.Request.Context(), requestLogger.With(slog.Int("user_id", userID))))
		c.Next()

	}
}
