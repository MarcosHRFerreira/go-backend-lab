package comment

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LikeOrUnlikeComment(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req dto.LikeOrUnLikeCommentRequest
	)
	// Validate the payload first so only well-formed toggle requests reach the service.
	// Valida a carga primeiro para que apenas requisicoes bem formadas de alternancia cheguem ao service.
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	// Recover the current user from middleware to enforce server-side ownership of the action.
	// Recupera o usuario atual do middleware para impor no servidor a autoria da acao.
	userID := c.GetInt("userID")

	// Delegate the like/unlike decision to the service layer.
	// Delega a decisao de like ou unlike para a camada de service.
	if err := h.commentService.LikeOrUnLikeComment(ctx, req.CommentID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return a minimal success response once the comment like state has been updated.
	// Retorna uma resposta minima de sucesso assim que o estado do like do comentario foi atualizado.
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
