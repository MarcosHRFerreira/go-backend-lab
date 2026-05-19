package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RefreshToken(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.RefreshTokenRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")
	token, refreshToken, err := h.userService.RefreshToken(ctx, &req, userID)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
