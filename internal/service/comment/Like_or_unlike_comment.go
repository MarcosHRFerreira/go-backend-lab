package comment

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/model"
	"time"
)

func (s *commentService) LikeOrUnLikeComment(ctx context.Context, commentID, userID int) error {

	// Ensure the comment exists before toggling any reaction state.
	// Garante que o comentario existe antes de alternar qualquer estado de reacao.
	commentExist, err := s.commentRepo.DetailComment(ctx, commentID)
	if err != nil {
		return apperror.Internal("failed to load comment", err)
	}
	if commentExist == nil {
		return apperror.NotFound("comment not found")
	}
	// Ask persistence whether the like already exists so the toggle stays deterministic.
	// Consulta a persistencia para saber se o like ja existe, mantendo a alternancia deterministica.
	isUserAlreadyLikeComment, err := s.commentRepo.IsUserAlreadyLikeComment(ctx, commentID, userID)
	if err != nil {
		return apperror.Internal("failed to check comment like", err)
	}

	if isUserAlreadyLikeComment {
		// Delete the existing like when the user is undoing a previous reaction.
		// Exclui o like existente quando o usuario esta desfazendo uma reacao anterior.
		err := s.commentRepo.DeleteLikeComment(ctx, commentID, userID)
		if err != nil {
			return apperror.Internal("failed to delete comment like", err)
		}

	} else {
		// Create the like row with timestamps when the user reacts for the first time.
		// Cria a linha do like com timestamps quando o usuario reage pela primeira vez.
		now := time.Now()
		err := s.commentRepo.StoreLikeComment(ctx, &model.CommentLikeModel{
			UserID:    userID,
			CommentID: commentID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return apperror.Internal("failed to create comment like", err)
		}
	}

	return nil

}
