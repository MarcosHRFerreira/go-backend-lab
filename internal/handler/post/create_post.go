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
	// Reject malformed post payloads before any domain rule or persistence is executed.
	// Rejeita cargas de post malformadas antes de executar qualquer regra de dominio ou persistencia.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Use the authenticated user identity from middleware instead of trusting client input.
	// Usa a identidade autenticada vinda do middleware em vez de confiar em entrada do cliente.
	userID := c.GetInt("userID")
	// Delegate post creation and ownership rules to the service layer.
	// Delega a criacao do post e as regras de autoria para a camada de service.
	postID, err := h.postService.CreatePost(ctx, &req, userID)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return only the generated identifier because the client already knows the submitted content.
	// Retorna apenas o identificador gerado porque o cliente ja conhece o conteudo enviado.
	c.JSON(http.StatusCreated, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}
