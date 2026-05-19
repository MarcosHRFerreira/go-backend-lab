package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	gin "github.com/gin-gonic/gin"
)

func (h *Handler) CreatePost(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.CreateOrUpdatePostRequest
	)
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID := c.GetInt("userID")
	postID, err := h.postService.CreatePost(ctx, &req, userID)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}
