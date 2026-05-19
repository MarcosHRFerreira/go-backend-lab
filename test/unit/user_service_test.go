package unit_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"go-tweets/internal/dto"
	"go-tweets/internal/model"
	userservice "go-tweets/internal/service/user"

	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceRegisterSuccess(t *testing.T) {
	t.Parallel()

	var createdUser *model.UserModel
	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByEmailOrUsernameFunc: func(ctx context.Context, email, username string) (*model.UserModel, error) {
			return nil, nil
		},
		createUserFunc: func(ctx context.Context, user *model.UserModel) (int, error) {
			createdUser = user
			return 10, nil
		},
	})

	userID, err := service.Register(context.Background(), &dto.RegisterRequest{
		Email:           "marcos@example.com",
		Username:        "marcos",
		Password:        "secret123",
		PasswordConfirm: "secret123",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if userID != 10 {
		t.Fatalf("expected user id 10, got %d", userID)
	}

	if createdUser == nil {
		t.Fatal("expected created user payload")
	}

	if createdUser.Password == "secret123" {
		t.Fatal("expected password to be hashed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(createdUser.Password), []byte("secret123")); err != nil {
		t.Fatalf("expected valid password hash, got %v", err)
	}
}

func TestUserServiceRegisterReturnsBadRequestWhenUserAlreadyExists(t *testing.T) {
	t.Parallel()

	createUserCalled := false
	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByEmailOrUsernameFunc: func(ctx context.Context, email, username string) (*model.UserModel, error) {
			return &model.UserModel{ID: 1}, nil
		},
		createUserFunc: func(ctx context.Context, user *model.UserModel) (int, error) {
			createUserCalled = true
			return 0, nil
		},
	})

	_, err := service.Register(context.Background(), &dto.RegisterRequest{
		Email:           "marcos@example.com",
		Username:        "marcos",
		Password:        "secret123",
		PasswordConfirm: "secret123",
	})

	assertStatusCode(t, err, http.StatusBadRequest)

	if err.Error() != "user already exists" {
		t.Fatalf("expected user already exists error, got %q", err.Error())
	}

	if createUserCalled {
		t.Fatal("expected create user not to be called")
	}
}

func TestUserServiceLoginReturnsExistingRefreshToken(t *testing.T) {
	t.Parallel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	storeRefreshTokenCalled := false
	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByEmailOrUsernameFunc: func(ctx context.Context, email, username string) (*model.UserModel, error) {
			return &model.UserModel{
				ID:       5,
				Username: "marcos",
				Password: string(hashedPassword),
			}, nil
		},
		getRefreshTokenFunc: func(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
			return &model.RefreshTokenModel{
				UserID:       userID,
				RefreshToken: "existing-refresh-token",
			}, nil
		},
		storeRefreshTokenFunc: func(ctx context.Context, token *model.RefreshTokenModel) error {
			storeRefreshTokenCalled = true
			return nil
		},
	})

	token, refreshToken, err := service.Login(context.Background(), &dto.LoginRequest{
		Email:    "marcos@example.com",
		Password: "secret123",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if refreshToken != "existing-refresh-token" {
		t.Fatalf("expected existing refresh token, got %q", refreshToken)
	}

	if storeRefreshTokenCalled {
		t.Fatal("expected store refresh token not to be called")
	}

	assertTokenClaims(t, token, 5, "marcos")
}

func TestUserServiceLoginReturnsUnauthorizedForInvalidPassword(t *testing.T) {
	t.Parallel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByEmailOrUsernameFunc: func(ctx context.Context, email, username string) (*model.UserModel, error) {
			return &model.UserModel{
				ID:       5,
				Username: "marcos",
				Password: string(hashedPassword),
			}, nil
		},
	})

	_, _, err = service.Login(context.Background(), &dto.LoginRequest{
		Email:    "marcos@example.com",
		Password: "wrong-password",
	})

	assertStatusCode(t, err, http.StatusUnauthorized)

	if err.Error() != "wrong email or password" {
		t.Fatalf("expected wrong email or password error, got %q", err.Error())
	}
}

func TestUserServiceRefreshTokenSuccess(t *testing.T) {
	t.Parallel()

	deletedRefreshToken := false
	var storedRefreshToken *model.RefreshTokenModel
	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByIDFunc: func(ctx context.Context, userID int) (*model.UserModel, error) {
			return &model.UserModel{
				ID:       userID,
				Username: "marcos",
			}, nil
		},
		getRefreshTokenFunc: func(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
			return &model.RefreshTokenModel{
				UserID:       userID,
				RefreshToken: "old-refresh-token",
			}, nil
		},
		deleteRefreshTokenFunc: func(ctx context.Context, userID int) error {
			deletedRefreshToken = true
			return nil
		},
		storeRefreshTokenFunc: func(ctx context.Context, token *model.RefreshTokenModel) error {
			storedRefreshToken = token
			return nil
		},
	})

	token, refreshToken, err := service.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "old-refresh-token",
	}, 7)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !deletedRefreshToken {
		t.Fatal("expected old refresh token to be deleted")
	}

	if storedRefreshToken == nil {
		t.Fatal("expected new refresh token to be stored")
	}

	if refreshToken == "old-refresh-token" {
		t.Fatal("expected a new refresh token to be generated")
	}

	if refreshToken != storedRefreshToken.RefreshToken {
		t.Fatalf("expected stored refresh token %q, got %q", storedRefreshToken.RefreshToken, refreshToken)
	}

	assertTokenClaims(t, token, 7, "marcos")
}

func TestUserServiceRefreshTokenReturnsUnauthorizedForMismatchedToken(t *testing.T) {
	t.Parallel()

	deleteCalled := false
	storeCalled := false
	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByIDFunc: func(ctx context.Context, userID int) (*model.UserModel, error) {
			return &model.UserModel{
				ID:       userID,
				Username: "marcos",
			}, nil
		},
		getRefreshTokenFunc: func(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
			return &model.RefreshTokenModel{
				UserID:       userID,
				RefreshToken: "expected-refresh-token",
			}, nil
		},
		deleteRefreshTokenFunc: func(ctx context.Context, userID int) error {
			deleteCalled = true
			return nil
		},
		storeRefreshTokenFunc: func(ctx context.Context, token *model.RefreshTokenModel) error {
			storeCalled = true
			return nil
		},
	})

	_, _, err := service.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "different-token",
	}, 7)

	assertStatusCode(t, err, http.StatusUnauthorized)

	if err.Error() != "refresh token not found" {
		t.Fatalf("expected refresh token not found error, got %q", err.Error())
	}

	if deleteCalled {
		t.Fatal("expected delete refresh token not to be called")
	}

	if storeCalled {
		t.Fatal("expected store refresh token not to be called")
	}
}

func TestUserServiceRefreshTokenReturnsInternalErrorWhenDeleteFails(t *testing.T) {
	t.Parallel()

	service := userservice.NewUserService(newTestConfig(), &userRepositoryStub{
		getUserByIDFunc: func(ctx context.Context, userID int) (*model.UserModel, error) {
			return &model.UserModel{
				ID:       userID,
				Username: "marcos",
			}, nil
		},
		getRefreshTokenFunc: func(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
			return &model.RefreshTokenModel{
				UserID:       userID,
				RefreshToken: "old-refresh-token",
			}, nil
		},
		deleteRefreshTokenFunc: func(ctx context.Context, userID int) error {
			return errors.New("delete failed")
		},
	})

	_, _, err := service.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "old-refresh-token",
	}, 7)

	assertStatusCode(t, err, http.StatusInternalServerError)

	if err.Error() != "failed to delete previous refresh token" {
		t.Fatalf("expected delete refresh token error, got %q", err.Error())
	}
}
