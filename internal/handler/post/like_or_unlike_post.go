package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LikeOrUnlikePost(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LikeOrUnlikePostRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")

	if err := h.postService.LikeOrUnlikePost(ctx, req.PostID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
