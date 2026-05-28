package unit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	obsmetrics "go-tweets/internal/observability/metrics"
)

func TestMetricsRegistryObservesDatabaseQueryDuration(t *testing.T) {
	t.Parallel()

	registry := obsmetrics.NewRegistry()
	registry.ObserveQuery("SELECT id, username FROM users WHERE id = ?", 150*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()
	registry.Handler().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "go_tweets_db_query_duration_seconds") {
		t.Fatalf("expected db query duration metric, got %s", body)
	}

	if !strings.Contains(body, "operation=\"select\"") {
		t.Fatalf("expected select operation label, got %s", body)
	}

	if !strings.Contains(body, "table=\"users\"") {
		t.Fatalf("expected users table label, got %s", body)
	}
}
