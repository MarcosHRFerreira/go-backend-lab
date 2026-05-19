// Package user implements user business logic.
package user

import (
	"context"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/repository/user"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (int, error)
	Login(ctx context.Context, req *dto.LoginRequest) (string, string, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error)
}

type userService struct {
	cfg      *config.Config
	userRepo user.UserRepository
}

func NewUserService(cfg *config.Config, userRepo user.UserRepository) *userService {
	return &userService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}
