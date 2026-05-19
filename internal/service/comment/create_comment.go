package comment

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"time"
)

func (s *commentService) CreateComment(ctx context.Context, req *dto.StoreCommentRequest, userID int) error {
	// check if post exists
	postExist, err := s.postRepo.GetPostByID(ctx, req.PostID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExist == nil {
		return apperror.NotFound("tweet not found")
	}

	now := time.Now()
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

	return nil

}
