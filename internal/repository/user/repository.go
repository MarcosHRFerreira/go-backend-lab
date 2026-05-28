// Package user provides user persistence operations.
package user

import (
	"context"

	"go-tweets/internal/model"
	"go-tweets/pkg/internalsql"
	"time"
)

type UserRepository interface {
	GetUserByEmailOrUsername(ctx context.Context, email, username string) (*model.UserModel, error)
	CreateUser(ctx context.Context, model *model.UserModel) (int, error)
	GetRefreshToken(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error)
	StoreRefreshToken(ctx context.Context, model *model.RefreshTokenModel) error
	GetUserByID(ctx context.Context, userID int) (*model.UserModel, error)
	DeleteRefreshToken(ctx context.Context, userID int) error
}

type userRepository struct {
	db internalsql.QueryExecutor
}

func NewRepository(db internalsql.QueryExecutor) UserRepository {
	return &userRepository{
		db: db,
	}
}
