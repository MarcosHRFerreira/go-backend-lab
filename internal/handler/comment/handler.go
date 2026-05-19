// Package comment handles HTTP requests for comment operations.
package comment

import (
	"go-tweets/internal/middleware"
	"go-tweets/internal/service/comment"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	api            *gin.Engine
	validate       *validator.Validate
	commentService comment.CommentService
}

func NewHandler(api *gin.Engine, validate *validator.Validate, commentService comment.CommentService) *Handler {
	return &Handler{
		api:            api,
		validate:       validate,
		commentService: commentService,
	}
}

func (h *Handler) RouteList(secretKey string) {
	routeAuth := h.api.Group("/comment")
	routeAuth.Use(middleware.AuthMiddleware(secretKey))
	routeAuth.POST("/", h.CreateComment)
	routeAuth.POST("/action", h.LikeOrUnlikeComment)
}
