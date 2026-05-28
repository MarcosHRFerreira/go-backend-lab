package integration_test

import (
	"context"
	"net/http"
	"testing"

	"go-tweets/internal/dto"
	"go-tweets/internal/middleware"

	"go.opentelemetry.io/otel/trace"
)

func TestTracingLoginPropagatesTraceIDAndSpanContext(t *testing.T) {
	t.Parallel()

	var (
		receivedTraceID string
		validSpan       bool
	)

	router := newTestRouter(&userServiceStub{
		loginFunc: func(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
			spanContext := trace.SpanContextFromContext(ctx)
			validSpan = spanContext.IsValid()
			receivedTraceID = spanContext.TraceID().String()
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

	if !validSpan {
		t.Fatal("expected valid span context in service")
	}

	responseTraceID := recorder.Header().Get(middleware.TraceIDHeader)
	if responseTraceID == "" {
		t.Fatal("expected response to include trace id header")
	}

	if receivedTraceID != responseTraceID {
		t.Fatalf("expected service trace id %q, got %q", responseTraceID, receivedTraceID)
	}
}
