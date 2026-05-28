package post

import (
	"go-tweets/internal/httpresponse"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeletePost(c *gin.Context) {
	var (
		ctx    = c.Request.Context()
		userID = c.GetInt("userID")
	)

	// Parse the route parameter before asking the service to enforce ownership rules.
	// Faz o parse do parametro da rota antes de pedir ao service para aplicar as regras de autoria.
	postID, ok := httpresponse.ParseIntParam(c, "post_id")
	if !ok {
		return
	}

	// Delegate existence checks, authorization, and soft delete semantics to the service.
	// Delega as verificacoes de existencia, autorizacao e semantica de exclusao logica ao service.
	if err := h.postService.DeletePost(ctx, postID, userID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	// Return a simple confirmation because the resource no longer needs to be rendered.
	// Retorna uma confirmacao simples porque o recurso nao precisa mais ser renderizado.
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted tweet",
	})

}
