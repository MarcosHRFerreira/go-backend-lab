package integration_test

import (
	"context"
	"net/http"
	"testing"

	"go-tweets/internal/dto"
	"go-tweets/internal/observability/logctx"

	"github.com/gin-gonic/gin"
)

func TestObservabilityLoginGeneratesAndPropagatesRequestID(t *testing.T) {
	t.Parallel()

	var receivedRequestID string
	router := newTestRouter(&userServiceStub{
		loginFunc: func(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
			receivedRequestID = logctx.RequestID(ctx)
			return "access-token", "refresh-token", nil
		},
	}, &postServiceStub{}, &commentServiceStub{})

	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "marcos@example.com",
		"password": "secret123",
	}, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	responseRequestID := recorder.Header().Get(logctx.RequestIDHeader)
	if responseRequestID == "" {
		t.Fatal("expected response to include request id header")
	}

	if receivedRequestID != responseRequestID {
		t.Fatalf("expected service context request id %q, got %q", responseRequestID, receivedRequestID)
	}
}

func TestObservabilityRefreshPreservesIncomingRequestID(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{
		refreshTokenFunc: func(ctx context.Context, req *dto.RefreshTokenRequest, userID int) (string, string, error) {
			return "new-access-token", "new-refresh-token", nil
		},
	}, &postServiceStub{}, &commentServiceStub{})

	const requestID = "req-from-client-123"
	recorder := performJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "current-refresh-token",
	}, map[string]string{
		"Authorization":        mustCreateToken(t, 9, "marcos"),
		logctx.RequestIDHeader: requestID,
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Header().Get(logctx.RequestIDHeader) != requestID {
		t.Fatalf("expected request id %q to be preserved, got %q", requestID, recorder.Header().Get(logctx.RequestIDHeader))
	}
}

func TestObservabilityRecoveryReturnsInternalServerErrorAndRequestID(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{}, &postServiceStub{}, &commentServiceStub{})
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	recorder := performJSONRequest(t, router, http.MethodGet, "/panic", nil, nil)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}

	payload := decodeJSONResponse(t, recorder)
	if payload["message"] != "internal server error" {
		t.Fatalf("expected internal server error message, got %v", payload["message"])
	}

	if recorder.Header().Get(logctx.RequestIDHeader) == "" {
		t.Fatal("expected panic response to include request id header")
	}
}
