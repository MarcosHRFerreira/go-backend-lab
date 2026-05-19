package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LoginRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	token, refreshToken, err := h.userService.Login(ctx, &req)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
