// File: internal/adapters/inbound/runtime/server_test.go
package runtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"multix/internal/adapters/inbound/agent"
	"multix/internal/platform/logger"
)

func TestHealthHandler(t *testing.T) {
	// Arrange
	log := logger.New("error")
	s := NewServer(log, nil, 8080)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	// Act: Pass request through mux to trigger Go 1.22 method matching
	s.mux.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
	if service, ok := response["service"]; !ok || service != "multix" {
		t.Errorf("handler returned unexpected service: got %v", rr.Body.String())
	}
	if mode, ok := response["mode"]; !ok || mode != "runtime" {
		t.Errorf("handler returned unexpected mode: got %v", rr.Body.String())
	}
}

func TestHealthHandler_WrongMethod(t *testing.T) {
	log := logger.New("error")
	s := NewServer(log, nil, 8080)

	req, err := http.NewRequest(http.MethodPost, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestToolsHandler(t *testing.T) {
	log := logger.New("error")
	
	// Create a dummy ToolAdapter with no skills, which should just return an empty JS array []
	adapter := agent.NewToolAdapter(nil, nil)
	s := NewServer(log, adapter, 8080)

	req, err := http.NewRequest(http.MethodGet, "/tools", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	// 1. Assert 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 2. Assert Content-Type contains application/json
	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v", ctype, "application/json")
	}

	// 3. Assert valid JSON array payload (not empty string)
	if rr.Body.Len() == 0 {
		t.Errorf("handler returned empty body")
	}

	var manifests []agent.Manifest
	if err := json.Unmarshal(rr.Body.Bytes(), &manifests); err != nil {
		t.Fatalf("Failed to parse response body as JSON array: %v", err)
	}
}
