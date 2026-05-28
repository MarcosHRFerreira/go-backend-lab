package post

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LikeOrUnlikePost(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LikeOrUnlikePostRequest
	)
	// Validate the toggle payload before touching the like state machine.
	// Valida a carga de alternancia antes de tocar na maquina de estado do like.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Use the authenticated identity from middleware instead of trusting the request body.
	// Usa a identidade autenticada do middleware em vez de confiar no corpo da requisicao.
	userID := c.GetInt("userID")

	// Delegate the toggle decision to the service because it depends on current persistence state.
	// Delega a decisao de alternancia ao service porque ela depende do estado atual na persistencia.
	if err := h.postService.LikeOrUnlikePost(ctx, req.PostID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return a lightweight confirmation after the like state is reconciled.
	// Retorna uma confirmacao leve depois que o estado do like foi reconciliado.
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
