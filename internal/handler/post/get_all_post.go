package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetallPost(c *gin.Context) {
	ctx := c.Request.Context()
	// Parse pagination inputs early so invalid query parameters fail with a client error.
	// Faz o parse da paginacao logo no inicio para que queries invalidas falhem como erro do cliente.
	page, ok := httpresponse.ParseMinInt64Query(c, "page", "1", 1)
	if !ok {
		return
	}

	// Keep the limit bounded by validation rules before reaching the service layer.
	// Mantem o limite protegido pelas regras de validacao antes de chegar ao service.
	limit, ok := httpresponse.ParseMinInt64Query(c, "limit", "2", 1)
	if !ok {
		return
	}

	// Build a transport DTO so HTTP query parameters stay isolated from business logic.
	// Monta um DTO de transporte para isolar os parametros HTTP da logica de negocio.
	param := dto.GetAllPostRequest{
		Limit: limit,
		Page:  page,
	}
	// Delegate pagination, aggregation, and mapping to the service layer.
	// Delega paginacao, agregacao e mapeamento para a camada de service.
	result, err := h.postService.GetAllPost(ctx, &param)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return the already-shaped timeline payload expected by the client.
	// Retorna a carga da timeline ja montada no formato esperado pelo cliente.
	c.JSON(http.StatusOK, result)

}
