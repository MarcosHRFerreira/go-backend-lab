package comment

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/model"
	"time"
)

func (s *commentService) LikeOrUnLikeComment(ctx context.Context, commentID, userID int) error {

	commentExist, err := s.commentRepo.DetailComment(ctx, commentID)
	if err != nil {
		return apperror.Internal("failed to load comment", err)
	}
	if commentExist == nil {
		return apperror.NotFound("comment not found")
	}
	isUserAlreadyLikeComment, err := s.commentRepo.IsUserAlreadyLikeComment(ctx, commentID, userID)
	if err != nil {
		return apperror.Internal("failed to check comment like", err)
	}

	if isUserAlreadyLikeComment {
		err := s.commentRepo.DeleteLikeComment(ctx, commentID, userID)
		if err != nil {
			return apperror.Internal("failed to delete comment like", err)
		}

	} else {
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
