package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.RegisterRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID, err := h.userService.Register(ctx, &req)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID: userID,
	})

}
