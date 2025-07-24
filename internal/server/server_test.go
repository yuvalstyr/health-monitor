package server

import (
	"net/http"
	"testing"

	"health-monitor/internal/db"
)

// mockQuerier is a minimal mock for testing
type mockQuerier struct{}

func (m *mockQuerier) ListActiveGaugeTemplates(ctx interface{}) ([]interface{}, error) {
	return nil, nil
}

func TestSetupRouter(t *testing.T) {
	queries := &mockQuerier{}
	router := SetupRouter(queries)

	if router == nil {
		t.Error("Expected router to be created")
	}
}

func TestNew(t *testing.T) {
	handler := http.NewServeMux()
	server := New("8080", handler)

	if server.Addr != ":8080" {
		t.Errorf("Expected server address :8080, got %s", server.Addr)
	}

	if server.Handler != handler {
		t.Error("Expected handler to be set correctly")
	}
}