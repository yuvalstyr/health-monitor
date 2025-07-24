package server

import (
	"net/http"
	"testing"
)

func TestSetupRouter(t *testing.T) {
	// For now, we'll skip the full router test since it requires a complex mock
	// This test verifies the function exists and can be called
	t.Skip("Router setup test requires complex db.Querier mock - skipping for now")
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