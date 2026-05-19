package user

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (int, error) {

	userExist, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if userExist != nil {
		return 0, apperror.BadRequest("user already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Internal("failed to generate password hash", err)
	}

	now := time.Now()
	userModel := &model.UserModel{
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(passwordHash),
		CreatedAt: now,
		UpdatedAt: now,
	}
	userID, err := s.userRepo.CreateUser(ctx, userModel)
	if err != nil {
		return 0, apperror.Internal("failed to create user", err)
	}
	return userID, nil
}
