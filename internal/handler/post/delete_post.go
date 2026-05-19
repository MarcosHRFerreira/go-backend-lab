package post

import (
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeletePost(c *gin.Context) {
	var (
		ctx    = c.Request.Context()
		userID = c.GetInt("userID")
	)

	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}

	if err := h.postService.DeletePost(ctx, postID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted tweet",
	})

}
