package integration_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"go-tweets/internal/dto"
)

func TestMetricsEndpointExposesHTTPMetrics(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&userServiceStub{
		loginFunc: func(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
			return "access-token", "refresh-token", nil
		},
	}, &postServiceStub{}, &commentServiceStub{})

	performJSONRequest(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "marcos@example.com",
		"password": "secret123",
	}, nil)

	recorder := performJSONRequest(t, router, http.MethodGet, "/metrics", nil, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "go_tweets_http_requests_total") {
		t.Fatalf("expected http requests metric, got %s", body)
	}

	if !strings.Contains(body, "route=\"/auth/login\"") {
		t.Fatalf("expected login route metric, got %s", body)
	}

	if !strings.Contains(body, "go_tweets_http_request_duration_seconds") {
		t.Fatalf("expected http duration metric, got %s", body)
	}
}
