package user

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"go-tweets/pkg/internalsql/jwt"
	"go-tweets/pkg/internalsql/refreshtoken"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {

	userExist, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, "")
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}

	if userExist == nil {
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userExist.Password), []byte(req.Password))
	if err != nil {
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	token, err := jwt.CreateToken(userExist.ID, userExist.Username, s.cfg.SecretJwt)
	if err != nil {
		return "", "", apperror.Internal("failed to create token", err)
	}

	now := time.Now()
	refreshTokenExist, err := s.userRepo.GetRefreshToken(ctx, userExist.ID, now)
	if err != nil {
		return "", "", apperror.Internal("failed to load refresh token", err)
	}

	if refreshTokenExist != nil {
		return token, refreshTokenExist.RefreshToken, nil
	}

	refreshToken, err := refreshtoken.GenerateRefreshToken()
	if err != nil {
		return "", "", apperror.Internal("failed to generate refresh token", err)
	}

	err = s.userRepo.StoreRefreshToken(ctx, &model.RefreshTokenModel{
		UserID:       userExist.ID,
		RefreshToken: refreshToken,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	if err != nil {
		return "", "", apperror.Internal("failed to store refresh token", err)
	}

	return token, refreshToken, nil

}
