package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetallPost(c *gin.Context) {
	ctx := c.Request.Context()
	page, ok := httpresponse.ParseMinInt64Query(c, "page", "1", 1)
	if !ok {
		return
	}

	limit, ok := httpresponse.ParseMinInt64Query(c, "limit", "2", 1)
	if !ok {
		return
	}

	param := dto.GetAllPostRequest{
		Limit: limit,
		Page:  page,
	}
	result, err := h.postService.GetAllPost(ctx, &param)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)

}
