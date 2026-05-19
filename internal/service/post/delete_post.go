package post

import (
	"context"
	"go-tweets/internal/apperror"
	"time"
)

func (s *postService) DeletePost(ctx context.Context, postID, userID int) error {
	postExits, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExits == nil {
		return apperror.NotFound("tweet not found")
	}

	if postExits.UserID != userID {
		return apperror.NotFound("tweet not found")

	}

	err = s.postRepo.SoftDeletePost(ctx, postID, time.Now())
	if err != nil {
		return apperror.Internal("failed to delete post", err)
	}
	return nil

}
