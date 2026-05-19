// Package middleware provides Gin middlewares for authentication.
package middleware

import (
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-tweets/pkg/internalsql/jwt"
)

func AuthMiddleware(secretkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "missing authorization token")
			return
		}

		userID, username, err := jwt.ValidateToken(header, secretkey, true)
		if err != nil {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}
		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()

	}
}
func AuthRefreshTokenMiddleware(secretkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "missing authorization token")
			return
		}

		userID, username, err := jwt.ValidateToken(header, secretkey, false)
		if err != nil {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}
		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()

	}
}
