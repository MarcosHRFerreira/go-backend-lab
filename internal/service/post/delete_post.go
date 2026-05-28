package post

import (
	"context"
	"go-tweets/internal/apperror"
	"log/slog"
	"time"
)

func (s *postService) DeletePost(ctx context.Context, postID, userID int) error {
	logger := serviceLogger(ctx, "delete_post").With(
		slog.Int("post_id", postID),
		slog.Int("user_id", userID),
	)
	logger.Info("delete post requested")

	// Fetch the post before deleting so the service can enforce existence and ownership rules.
	// Busca o post antes de excluir para que o service aplique as regras de existencia e autoria.
	postExits, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExits == nil {
		return apperror.NotFound("tweet not found")
	}

	// Use not found for foreign ownership to avoid exposing which posts belong to other users.
	// Usa not found para autoria de terceiros a fim de nao expor quais posts pertencem a outros usuarios.
	if postExits.UserID != userID {
		logger.Warn("delete post rejected", slog.String("error_kind", "ownership_mismatch"))
		return apperror.NotFound("tweet not found")

	}

	// Perform a soft delete so the timeline can hide the post without losing audit history.
	// Executa uma exclusao logica para que a timeline oculte o post sem perder o historico de auditoria.
	err = s.postRepo.SoftDeletePost(ctx, postID, time.Now())
	if err != nil {
		return apperror.Internal("failed to delete post", err)
	}
	logger.Info("post soft deleted")
	return nil

}
