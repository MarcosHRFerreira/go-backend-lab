package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"time"
)

func (s *postService) UpdatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int) error {
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
	err = s.postRepo.UpdatePost(ctx, &model.PostModel{
		Title:     req.Title,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}, postID)
	if err != nil {
		return apperror.Internal("failed to update post", err)
	}
	return nil

}
