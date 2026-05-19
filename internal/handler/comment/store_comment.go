package comment

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	gin "github.com/gin-gonic/gin"
)

func (h *Handler) CreateComment(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.StoreCommentRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")
	if err := h.commentService.CreateComment(ctx, &req, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "successfully",
	})
}
