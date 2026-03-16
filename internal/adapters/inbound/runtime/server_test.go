// File: internal/adapters/inbound/runtime/server_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Unit tests for the MULTIX HTTP runtime server.

package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"multix/internal/adapters/inbound/agent"
	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
	"multix/internal/platform/logger"
)

type mockSkill struct {}
func (m *mockSkill) Name() string { return "test.skill" }
func (m *mockSkill) Description() string { return "Mock skill" }
func (m *mockSkill) InputSchema() any { return nil }
func (m *mockSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	return map[string]string{"msg": "hello from mock", "echo_key": input["key"].(string)}, nil
}

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

func TestExecuteHandler_Success(t *testing.T) {
	log := logger.New("error")
	
	// Create a real registry and executor, and prep it with a mock skill
	registry := domainSkills.NewRegistry()
	registry.Register(&mockSkill{})
	executor := appSkills.NewExecutor(registry)
	adapter := agent.NewToolAdapter(registry, executor)

	s := NewServer(log, adapter, 8080)

	body := `{"skill": "test.skill", "provider": "aws", "params": {"key": "value"}}`
	req, err := http.NewRequest(http.MethodPost, "/execute", strings.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp executeSuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse success response: %v", err)
	}
	
	if !resp.Ok || resp.Skill != "test.skill" || resp.Provider != "aws" {
		t.Errorf("unexpected success envelope: %+v", resp)
	}
}

func TestExecuteHandler_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	s := NewServer(log, agent.NewToolAdapter(nil, nil), 8080)

	req, _ := http.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{malformed`))
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", rr.Code)
	}

	var resp executeErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Ok != false || resp.Error != "invalid json format" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestExecuteHandler_MissingSkill(t *testing.T) {
	log := logger.New("error")
	registry := domainSkills.NewRegistry()
	executor := appSkills.NewExecutor(registry)
	s := NewServer(log, agent.NewToolAdapter(registry, executor), 8080)

	req, _ := http.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{"provider":"aws"}`))
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing skill, got %d", rr.Code)
	}

	var resp executeErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Ok != false || resp.Error != "missing skill" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestExecuteHandler_UnknownSkill(t *testing.T) {
	log := logger.New("error")
	registry := domainSkills.NewRegistry()
	executor := appSkills.NewExecutor(registry)
	s := NewServer(log, agent.NewToolAdapter(registry, executor), 8080)

	req, _ := http.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{"skill":"invalid.skill"}`))
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	// In absence of a real active 'invalid.skill' in the registry, the agent executor will return error 'skill not found'
	// The user specified "404 Not Found if feasible, 500 otherwise". By default, executor returns an error, 
	// which our handler maps to 500 with {ok: false}. 
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for missing real skill resolution, got %d", rr.Code)
	}

	var resp executeErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Ok != false {
		t.Errorf("expected Ok=false")
	}
}

