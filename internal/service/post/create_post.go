package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"log/slog"
	"time"
)

func (s *postService) CreatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int) (int, error) {
	logger := serviceLogger(ctx, "create_post").With(slog.Int("user_id", userID))
	logger.Info("create post requested")

	// Capture timestamps once so the created record has a consistent lifecycle snapshot.
	// Captura os timestamps uma unica vez para que o registro criado tenha um retrato consistente do ciclo de vida.
	now := time.Now()
	// Translate the request DTO into the persistence model expected by the repository.
	// Traduz o DTO da requisicao para o model de persistencia esperado pelo repository.
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

	logger.Info("post created", slog.Int("post_id", insertedID))
	return insertedID, nil
}
