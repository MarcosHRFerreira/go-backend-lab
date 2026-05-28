package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.RegisterRequest
	)
	// Bind and validate the request body before delegating to the service layer.
	// Faz o bind e a validacao do corpo da requisicao antes de delegar para a camada de service.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// The handler only orchestrates HTTP concerns; business rules stay in the service.
	// O handler apenas orquestra preocupacoes HTTP; as regras de negocio permanecem no service.
	userID, err := h.userService.Register(ctx, &req)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return only the identifier needed by the client after successful creation.
	// Retorna apenas o identificador necessario para o cliente apos a criacao bem-sucedida.
	c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID: userID,
	})

}
