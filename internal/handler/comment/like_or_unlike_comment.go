package comment

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LikeOrUnlikeComment(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LikeOrUnLikeCommentRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")

	if err := h.commentService.LikeOrUnLikeComment(ctx, req.CommentID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
