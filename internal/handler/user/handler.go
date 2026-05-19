// Package user handles HTTP requests for user operations.
package user

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go-tweets/internal/middleware"
	"go-tweets/internal/service/user"
)

type Handler struct {
	api         *gin.Engine
	validate    *validator.Validate
	userService user.UserService
}

func NewHandler(api *gin.Engine, validate *validator.Validate, userService user.UserService) *Handler {
	return &Handler{
		api:         api,
		validate:    validate,
		userService: userService,
	}
}
func (h *Handler) RouteList(secretKey string) {
	authRoute := h.api.Group("/auth")
	authRoute.POST("/register", h.Register)
	authRoute.POST("/login", h.Login)
	refreshRoute := h.api.Group("/auth")
	refreshRoute.Use(middleware.AuthRefreshTokenMiddleware(secretKey))
	refreshRoute.POST("/refresh", h.RefreshToken)
}
