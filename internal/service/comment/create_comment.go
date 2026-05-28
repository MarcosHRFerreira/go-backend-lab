package comment

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"log/slog"
	"time"
)

func (s *commentService) CreateComment(ctx context.Context, req *dto.StoreCommentRequest, userID int) error {
	logger := serviceLogger(ctx, "create_comment").With(
		slog.Int("post_id", req.PostID),
		slog.Int("user_id", userID),
	)
	logger.Info("create comment requested")

	// Fail early when the target post does not exist instead of creating an orphan comment.
	// Falha cedo quando o post de destino nao existe, em vez de criar um comentario orfao.
	postExist, err := s.postRepo.GetPostByID(ctx, req.PostID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExist == nil {
		logger.Warn("create comment rejected", slog.String("error_kind", "post_not_found"))
		return apperror.NotFound("tweet not found")
	}

	now := time.Now()
	// Stamp creation and update times here so the persistence layer receives a complete model.
	// Define os tempos de criacao e atualizacao aqui para que a camada de persistencia receba um model completo.
	err = s.commentRepo.StoreComment(ctx, &model.CommentModel{
		PostID:    req.PostID,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return apperror.Internal("failed to create comment", err)
	}

	logger.Info("comment created")
	return nil

}
