// File: internal/adapters/inbound/runtime/server_test.go
package runtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

	// Act
	s.healthHandler(rr, req)

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
}

func TestHealthHandler_WrongMethod(t *testing.T) {
	log := logger.New("error")
	s := NewServer(log, nil, 8080)

	req, err := http.NewRequest(http.MethodPost, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	s.healthHandler(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
