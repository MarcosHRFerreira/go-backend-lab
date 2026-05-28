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

	"golang.org/x/crypto/bcrypt"
)

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
	logger := serviceLogger(ctx, "login")
	logger.Info("login attempted")

	// Look up the account by email first because authentication starts from user identity.
	// Procura a conta primeiro pelo email porque a autenticacao comeca pela identidade do usuario.
	userExist, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, "")
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}

	// Keep credential errors intentionally vague to avoid leaking which field failed.
	// Mantem os erros de credencial intencionalmente vagos para nao vazar qual campo falhou.
	if userExist == nil {
		logger.Warn("login failed", slog.String("error_kind", "invalid_credentials"))
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	// Compare the submitted password against the stored hash instead of handling plain text.
	// Compara a senha enviada com o hash armazenado em vez de lidar com texto puro.
	err = bcrypt.CompareHashAndPassword([]byte(userExist.Password), []byte(req.Password))
	if err != nil {
		logger.Warn("login failed", slog.String("error_kind", "invalid_credentials"))
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	// Issue a short-lived access token once the credentials are proven valid.
	// Emite um access token de curta duracao quando as credenciais foram validadas.
	token, err := jwt.CreateToken(userExist.ID, userExist.Username, s.cfg.SecretJwt)
	if err != nil {
		return "", "", apperror.Internal("failed to create token", err)
	}

	now := time.Now()
	// Reuse an existing active refresh token so repeated logins do not multiply session rows.
	// Reutiliza um refresh token ativo existente para que logins repetidos nao multipliquem linhas de sessao.
	refreshTokenExist, err := s.userRepo.GetRefreshToken(ctx, userExist.ID, now)
	if err != nil {
		return "", "", apperror.Internal("failed to load refresh token", err)
	}

	if refreshTokenExist != nil {
		logger.Info(
			"login succeeded",
			slog.Int("user_id", userExist.ID),
			slog.Bool("refresh_token_reused", true),
		)
		return token, refreshTokenExist.RefreshToken, nil
	}

	// Generate and persist a new refresh token only when there is no active one in storage.
	// Gera e persiste um novo refresh token somente quando nao existe um ativo armazenado.
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

	logger.Info(
		"login succeeded",
		slog.Int("user_id", userExist.ID),
		slog.Bool("refresh_token_reused", false),
	)

	return token, refreshToken, nil

}
