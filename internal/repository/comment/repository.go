// Package comment provides comment persistence operations.
package comment

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
)

type CommentRepository interface {
	StoreComment(ctx context.Context, model *model.CommentModel) error
	DetailComment(ctx context.Context, commentID int) (*model.CommentModel, error)
	IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int) (bool, error)
	DeleteLikeComment(ctx context.Context, commentID, userID int) error
	StoreLikeComment(ctx context.Context, model *model.CommentLikeModel) error
	GetCommentsByPostIDs(ctx context.Context, postIDs []int) ([]model.CommentModel, error)
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}
