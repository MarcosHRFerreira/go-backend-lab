package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdatePost(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.CreateOrUpdatePostRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")
	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}

	if err := h.postService.UpdatePost(ctx, &req, postID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}
