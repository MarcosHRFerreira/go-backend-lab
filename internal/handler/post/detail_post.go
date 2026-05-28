package post

import (
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DetailPost(c *gin.Context) {

	ctx := c.Request.Context()
	// Parse the route identifier first so malformed URLs stop at the HTTP boundary.
	// Faz o parse do identificador da rota primeiro para que URLs malformadas parem na borda HTTP.
	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}
	// Delegate the aggregate read to the service because it may combine post and comment data.
	// Delega a leitura agregada ao service porque ela pode combinar dados de post e comentario.
	result, err := h.postService.DetailPost(ctx, postID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return the already enriched post detail payload expected by the client.
	// Retorna a carga de detalhe do post ja enriquecida no formato esperado pelo cliente.
	c.JSON(http.StatusOK, result)

}
