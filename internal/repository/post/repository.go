// Package post provides post persistence operations.
package post

import (
	"context"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"go-tweets/pkg/internalsql"
	"time"
)

type PostRepository interface {
	StorePost(ctx context.Context, model *model.PostModel) (int, error)
	GetPostByID(ctx context.Context, postID int) (*model.PostWithUserModel, error)
	UpdatePost(ctx context.Context, model *model.PostModel, postID int) error
	SoftDeletePost(ctx context.Context, postID int, now time.Time) error
	IsUserAlreadyLikePost(ctx context.Context, postID, userID int) (bool, error)
	DeleteLikePost(ctx context.Context, postID, userID int) error
	StoreLikePost(ctx context.Context, model *model.PostLikeModel) error
	TotalPost(ctx context.Context) (int64, error)
	GetAllPost(ctx context.Context, param *dto.GetAllPostRequest, offset int) ([]model.PostWithUserModel, error)
}

type postRepository struct {
	db internalsql.QueryExecutor
}

func NewPostRepository(db internalsql.QueryExecutor) PostRepository {
	return &postRepository{
		db: db,
	}
}
