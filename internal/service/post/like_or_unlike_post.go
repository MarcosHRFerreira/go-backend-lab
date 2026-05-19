package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/model"
	"time"
)

func (s *postService) LikeOrUnlikePost(ctx context.Context, postID, userID int) error {

	postExist, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExist == nil {
		return apperror.NotFound("tweet not found")
	}

	isUserAlreadyLikePost, err := s.postRepo.IsUserAlreadyLikePost(ctx, postID, userID)
	if err != nil {
		return apperror.Internal("failed to check post like", err)
	}

	if isUserAlreadyLikePost {
		err := s.postRepo.DeleteLikePost(ctx, postID, userID)
		if err != nil {
			return apperror.Internal("failed to delete post like", err)
		}
	} else {
		now := time.Now()
		err := s.postRepo.StoreLikePost(ctx, &model.PostLikeModel{
			UserID:    userID,
			PostID:    postID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return apperror.Internal("failed to create post like", err)
		}

	}
	return nil
}
