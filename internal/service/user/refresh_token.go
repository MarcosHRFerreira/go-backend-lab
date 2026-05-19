package user

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"go-tweets/pkg/internalsql/jwt"
	"go-tweets/pkg/internalsql/refreshtoken"
	"time"
)

func (s *userService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error) {

	userExist, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}
	if userExist == nil {
		return "", "", apperror.NotFound("user not found")
	}

	refreshTokenExists, err := s.userRepo.GetRefreshToken(ctx, userID, time.Now())
	if err != nil {
		return "", "", apperror.Internal("failed to load refresh token", err)
	}
	if refreshTokenExists == nil {
		return "", "", apperror.Unauthorized("refresh token expired")
	}
	if req.RefreshToken != refreshTokenExists.RefreshToken {
		return "", "", apperror.Unauthorized("refresh token not found")
	}

	token, err := jwt.CreateToken(userID, userExist.Username, s.cfg.SecretJwt)
	if err != nil {
		return "", "", apperror.Internal("failed to create token", err)
	}

	err = s.userRepo.DeleteRefreshToken(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal("failed to delete previous refresh token", err)
	}

	refreshToken, err := refreshtoken.GenerateRefreshToken()
	if err != nil {
		return "", "", apperror.Internal("failed to generate refresh token", err)
	}

	now := time.Now()
	err = s.userRepo.StoreRefreshToken(ctx, &model.RefreshTokenModel{
		UserID:       userID,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return "", "", apperror.Internal("failed to store refresh token", err)
	}
	return token, refreshToken, nil

}
