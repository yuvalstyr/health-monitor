package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
)

func TestSetupRouter(t *testing.T) {
	// Create a mock querier using the existing mock implementation
	mockDB := &db.MockQueries{}
	
	// Setup router with mock (pass nil for database since we're using mock, and test version)
	router := SetupRouter(mockDB, nil, "test-version")
	
	// Test that router is created and has basic middleware
	if router == nil {
		t.Fatal("Expected router to be created, got nil")
	}
	
	// Test a basic request to ensure middleware is working
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should get 404 for non-existent route, which means router is working
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent route, got %d", w.Code)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectPanic bool
	}{
		{
			name:        "valid port",
			port:        "8080",
			expectPanic: false,
		},
		{
			name:        "empty port",
			port:        "",
			expectPanic: true,
		},
		{
			name:        "invalid port - too high",
			port:        "70000",
			expectPanic: true,
		},
		{
			name:        "invalid port - zero",
			port:        "0",
			expectPanic: true,
		},
		{
			name:        "invalid port - negative",
			port:        "-1",
			expectPanic: true,
		},
		{
			name:        "invalid port - non-numeric",
			port:        "abc",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.NewServeMux()
			
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic for port %s, but didn't panic", tt.port)
					}
				}()
			}
			
			server := New(tt.port, handler)
			
			if !tt.expectPanic {
				expectedAddr := ":" + tt.port
				if server.Addr != expectedAddr {
					t.Errorf("Expected server address %s, got %s", expectedAddr, server.Addr)
				}
				
				if server.Handler != handler {
					t.Error("Expected handler to be set correctly")
				}
			}
		})
	}
}

func TestNeuteredFileSystem(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a test file
	testFile := tempDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	nfs := neuteredFileSystem{http.Dir(tempDir)}
	
	// Test opening a file (should work)
	f, err := nfs.Open("/test.txt")
	if err != nil {
		t.Errorf("Expected to open file successfully, got error: %v", err)
	}
	if f != nil {
		f.Close()
	}
	
	// Test opening a directory (should fail)
	_, err = nfs.Open("/")
	if err == nil {
		t.Error("Expected error when opening directory, got nil")
	}
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestSetupRouterStaticFiles(t *testing.T) {
	mockDB := &db.MockQueries{}
	router := SetupRouter(mockDB, nil, "test-version")
	
	// Test static file route exists
	req := httptest.NewRequest("GET", "/static/test.css", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should get 404 since file doesn't exist, but route should be handled
	// (not a 405 Method Not Allowed which would indicate route doesn't exist)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent static file, got %d", w.Code)
	}
}

func TestSetupRouterHealthEndpoint(t *testing.T) {
	mockDB := &db.MockQueries{}
	
	// Create a mock database connection for health check
	// In a real test, you might use a test database, but for this test we'll pass nil
	// since the health handler will handle nil gracefully
	router := SetupRouter(mockDB, nil, "test-version")
	
	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should get 503 since database is nil (unhealthy)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 for unhealthy service, got %d", w.Code)
	}
	
	// Check response content type
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
	
	// Check response body structure
	var response handlers.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}
	
	if response.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", response.Status)
	}
	
	if response.Database != false {
		t.Errorf("Expected database false, got %v", response.Database)
	}
	
	if response.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", response.Version)
	}
}