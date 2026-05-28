package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"log/slog"
	"time"
)

func (s *postService) UpdatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int) error {
	logger := serviceLogger(ctx, "update_post").With(
		slog.Int("post_id", postID),
		slog.Int("user_id", userID),
	)
	logger.Info("update post requested")

	// Load the current post first so existence and ownership can be checked before updating.
	// Carrega primeiro o post atual para que existencia e autoria sejam verificadas antes da atualizacao.
	postExits, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExits == nil {
		return apperror.NotFound("tweet not found")
	}

	// Hide unauthorized access behind not found to avoid leaking resource ownership.
	// Esconde acesso nao autorizado atras de not found para evitar vazar a autoria do recurso.
	if postExits.UserID != userID {
		logger.Warn("update post rejected", slog.String("error_kind", "ownership_mismatch"))
		return apperror.NotFound("tweet not found")

	}
	// Persist only mutable fields and refresh the update timestamp.
	// Persiste apenas os campos mutaveis e atualiza o timestamp de modificacao.
	err = s.postRepo.UpdatePost(ctx, &model.PostModel{
		Title:     req.Title,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}, postID)
	if err != nil {
		return apperror.Internal("failed to update post", err)
	}

	logger.Info("post updated")
	return nil

}
