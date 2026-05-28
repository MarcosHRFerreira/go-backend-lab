package user

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (int, error) {
	logger := serviceLogger(ctx, "register")
	logger.Info("user registration requested")

	// Check business uniqueness first so the API returns a controlled domain error.
	// Verifica primeiro a unicidade de negocio para que a API devolva um erro de dominio controlado.
	userExist, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if userExist != nil {
		logger.Warn("user registration rejected", slog.String("error_kind", "user_already_exists"))
		return 0, apperror.BadRequest("user already exists")
	}

	// Persist only a password hash; plaintext passwords must never cross the repository boundary.
	// Persiste apenas o hash da senha; senhas em texto puro nunca devem atravessar a fronteira do repository.
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Internal("failed to generate password hash", err)
	}

	now := time.Now()
	// Convert the validated DTO into the persistence model expected by the repository layer.
	// Converte o DTO validado no model de persistencia esperado pela camada de repository.
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

	logger.Info("user registered", slog.Int("user_id", userID))
	return userID, nil
}
