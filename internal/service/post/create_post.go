package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"time"
)

func (s *postService) CreatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int) (int, error) {

	now := time.Now()
	insertedID, err := s.postRepo.StorePost(ctx, &model.PostModel{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	})

	if err != nil {
		return 0, apperror.Internal("failed to create post", err)
	}
	return insertedID, nil
}
