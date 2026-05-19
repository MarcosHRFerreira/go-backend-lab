package post

import (
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DetailPost(c *gin.Context) {

	ctx := c.Request.Context()
	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}
	result, err := h.postService.DetailPost(ctx, postID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)

}
