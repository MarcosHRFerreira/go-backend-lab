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
	// Validate the request body before ownership and existence checks run downstream.
	// Valida o corpo da requisicao antes que as verificacoes de autoria e existencia rodem abaixo.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Read the authenticated user from middleware so authorship cannot be forged by the client.
	// Le o usuario autenticado do middleware para que a autoria nao possa ser forjada pelo cliente.
	userID := c.GetInt("userID")
	// Parse the target resource identifier separately from the JSON payload.
	// Faz o parse do identificador do recurso separadamente da carga JSON.
	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}

	// Delegate authorization and persistence rules to the service layer.
	// Delega as regras de autorizacao e persistencia para a camada de service.
	if err := h.postService.UpdatePost(ctx, &req, postID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return the updated identifier as a minimal confirmation payload.
	// Retorna o identificador atualizado como confirmacao minima da operacao.
	c.JSON(http.StatusOK, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}
