package comment

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	gin "github.com/gin-gonic/gin"
)

func (h *Handler) CreateComment(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.StoreCommentRequest
	)
	// Validate the comment payload before invoking any rule that depends on the target post.
	// Valida a carga do comentario antes de invocar qualquer regra que dependa do post de destino.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Recover the current user from middleware so authorship is enforced server-side.
	// Recupera o usuario atual a partir do middleware para que a autoria seja garantida no servidor.
	userID := c.GetInt("userID")
	// Delegate existence checks and persistence rules to the service layer.
	// Delega as verificacoes de existencia e as regras de persistencia para a camada de service.
	if err := h.commentService.CreateComment(ctx, &req, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return a minimal confirmation because the main outcome is the side effect of creation.
	// Retorna uma confirmacao minima porque o principal resultado e o efeito colateral da criacao.
	c.JSON(http.StatusCreated, gin.H{
		"message": "successfully",
	})
}
