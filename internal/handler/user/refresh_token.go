package user

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RefreshToken(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.RefreshTokenRequest
	)
	// Validate the refresh payload before touching the token rotation flow.
	// Valida a carga do refresh antes de entrar no fluxo de rotacao de tokens.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Read the authenticated user identity injected by the access-token middleware.
	// Le a identidade autenticada injetada pelo middleware de access token.
	userID := c.GetInt("userID")
	// Delegate token verification, rotation, and issuance to the service layer.
	// Delega a verificacao do token, a rotacao e a emissao para a camada de service.
	token, refreshToken, err := h.userService.RefreshToken(ctx, &req, userID)

	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return the new token pair so the client can continue the authenticated session.
	// Retorna o novo par de tokens para que o cliente continue a sessao autenticada.
	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
