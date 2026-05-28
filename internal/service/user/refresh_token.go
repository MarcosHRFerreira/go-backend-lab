package user

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"go-tweets/pkg/internalsql/jwt"
	"go-tweets/pkg/internalsql/refreshtoken"
	"log/slog"
	"time"
)

func (s *userService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error) {
	logger := serviceLogger(ctx, "refresh_token").With(slog.Int("user_id", userID))
	logger.Info("refresh token requested")

	// Reload the user so token issuance always uses current account data.
	// Recarrega o usuario para que a emissao do token sempre use os dados atuais da conta.
	userExist, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}
	if userExist == nil {
		logger.Warn("refresh token rejected", slog.String("error_kind", "user_not_found"))
		return "", "", apperror.NotFound("user not found")
	}
	logger.Info("active user reloaded")

	// Refresh-token rotation is stateful in this project, so the stored token must still be valid.
	// A rotacao do refresh token e stateful neste projeto, entao o token armazenado ainda precisa estar valido.
	refreshTokenExists, err := s.userRepo.GetRefreshToken(ctx, userID, time.Now())
	if err != nil {
		return "", "", apperror.Internal("failed to load refresh token", err)
	}
	if refreshTokenExists == nil {
		logger.Warn("refresh token rejected", slog.String("error_kind", "refresh_token_expired"))
		return "", "", apperror.Unauthorized("refresh token expired")
	}
	if req.RefreshToken != refreshTokenExists.RefreshToken {
		logger.Warn("refresh token rejected", slog.String("error_kind", "refresh_token_not_found"))
		return "", "", apperror.Unauthorized("refresh token not found")
	}

	// Issue a new access token only after the refresh token has been fully verified.
	// Emite um novo access token somente depois que o refresh token foi totalmente verificado.
	token, err := jwt.CreateToken(userID, userExist.Username, s.cfg.SecretJwt)
	if err != nil {
		return "", "", apperror.Internal("failed to create token", err)
	}

	// Remove the previous refresh token before storing the replacement to enforce rotation.
	// Remove o refresh token anterior antes de salvar o substituto para garantir a rotacao.
	err = s.userRepo.DeleteRefreshToken(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal("failed to delete previous refresh token", err)
	}
	logger.Info("previous refresh token removed")

	refreshToken, err := refreshtoken.GenerateRefreshToken()
	if err != nil {
		return "", "", apperror.Internal("failed to generate refresh token", err)
	}

	now := time.Now()
	// Persist the new refresh token so future refresh requests can be verified against storage.
	// Persiste o novo refresh token para que futuras requisicoes de refresh possam ser verificadas no armazenamento.
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
	logger.Info("new refresh token stored")
	logger.Info("refresh token rotated")
	return token, refreshToken, nil

}
