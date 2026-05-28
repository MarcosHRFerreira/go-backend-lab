package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LoginRequest
	)
	// Reject malformed or invalid credentials payloads before reaching authentication logic.
	// Rejeita cargas de credenciais malformadas ou invalidas antes de chegar a logica de autenticacao.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Delegate credential verification and token issuance to the service layer.
	// Delega a verificacao das credenciais e a emissao dos tokens para a camada de service.
	token, refreshToken, err := h.userService.Login(ctx, &req)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return both tokens because the client needs them to authenticate and refresh the session.
	// Retorna os dois tokens porque o cliente precisa deles para autenticar e renovar a sessao.
	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
