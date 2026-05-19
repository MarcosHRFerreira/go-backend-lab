package integration_test

import (
	"context"
	"net/http"
	"testing"

	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
)

func TestAuthRegisterEndpointReturnsCreatedOnSuccess(t *testing.T) {
	t.Parallel()

	var receivedRequest *dto.RegisterRequest
	router := newTestRouter(&userServiceStub{
		registerFunc: func(ctx context.Context, req *dto.RegisterRequest) (int, error) {
			receivedRequest = req
			return 42, nil
		},
	}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":            "marcos@example.com",
		"username":         "marcos",
		"password":         "secret123",
		"password_confirm": "secret123",
	}, nil)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	if receivedRequest == nil || receivedRequest.Email != "marcos@example.com" {
		t.Fatalf("expected register request to reach service, got %+v", receivedRequest)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["id"] != float64(42) {
		t.Fatalf("expected id 42, got %v", payload["id"])
	}
}

func TestAuthRegisterEndpointReturnsBadRequestForInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    "invalid-email",
		"username": "ma",
	}, nil)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "validation failed" {
		t.Fatalf("expected validation failed message, got %v", payload["message"])
	}
}

func TestAuthLoginEndpointReturnsUnauthorizedOnInvalidCredentials(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{
		loginFunc: func(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
			return "", "", apperror.Unauthorized("wrong email or password")
		},
	}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "marcos@example.com",
		"password": "wrong-password",
	}, nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "wrong email or password" {
		t.Fatalf("expected wrong email or password message, got %v", payload["message"])
	}
}

func TestAuthRefreshEndpointReturnsTokenWhenAuthorized(t *testing.T) {
	t.Parallel()

	var receivedUserID int
	router := newTestRouter(&userServiceStub{
		refreshTokenFunc: func(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error) {
			receivedUserID = userID
			return "new-access-token", "new-refresh-token", nil
		},
	}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "current-refresh-token",
	}, map[string]string{
		"Authorization": mustCreateToken(t, 9, "marcos"),
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if receivedUserID != 9 {
		t.Fatalf("expected userID 9, got %d", receivedUserID)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["token"] != "new-access-token" {
		t.Fatalf("expected new access token, got %v", payload["token"])
	}

	if payload["refresh_token"] != "new-refresh-token" {
		t.Fatalf("expected new refresh token, got %v", payload["refresh_token"])
	}
}

func TestAuthRefreshEndpointReturnsUnauthorizedWithoutAuthorizationHeader(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "current-refresh-token",
	}, nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "missing authorization token" {
		t.Fatalf("expected missing authorization token message, got %v", payload["message"])
	}
}
