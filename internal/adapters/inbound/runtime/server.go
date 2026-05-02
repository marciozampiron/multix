// File: internal/adapters/inbound/runtime/server.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Core HTTP server adapter for the MULTIX local runtime.

package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"multix/internal/adapters/inbound/agent"
	"multix/internal/platform/logger"
)

// Server represents the local MULTIX HTTP runtime.
type Server struct {
	logger  logger.Logger
	agent   *agent.ToolAdapter
	mux     *http.ServeMux
	metrics *Metrics
	port    int
}

// NewServer initializes a new runtime Server.
func NewServer(logger logger.Logger, adapter *agent.ToolAdapter, port int) *Server {
	s := &Server{
		logger:  logger,
		agent:   adapter,
		mux:     http.NewServeMux(),
		metrics: NewMetrics(),
		port:    port,
	}
	s.registerRoutes()
	return s
}

// registerRoutes wires up all HTTP endpoints.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.healthHandler)
	s.mux.HandleFunc("GET /tools", s.toolsHandler)
	s.mux.HandleFunc("POST /execute", s.executeHandler)
	s.mux.HandleFunc("GET /capabilities", s.capabilitiesHandler)
	s.mux.HandleFunc("GET /metrics", s.metricsHandler)
}

// healthHandler provides a basic liveness probe.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":  "ok",
		"service": "multix",
		"mode":    "runtime",
	}

	_ = json.NewEncoder(w).Encode(response)
}

// capabilitiesHandler returns a JSON matrix of the runtime capabilities.
func (s *Server) capabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"api_version": "v1",
		"capabilities": []string{
			"tool_execution",
			"dynamic_manifests",
		},
		"supported_providers": []string{"aws", "gcp", "oci"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		s.logger.Error("Failed to encode capabilities response", err)
	}
}

// toolsHandler dynamically returns agent tool manifests based on registered skills.
func (s *Server) toolsHandler(w http.ResponseWriter, r *http.Request) {
	manifests := s.agent.Manifests()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manifests); err != nil {
		s.logger.Error("Failed to encode tool manifests", err)
	}
}

// metricsHandler exposes Prometheus-compatible runtime metrics.
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(s.metrics.Prometheus())); err != nil {
		s.logger.Error("Failed to write metrics response", err)
	}
}

// executeRequest defines the canonical payload for the /execute endpoint.
type executeRequest struct {
	Skill    string         `json:"skill"`
	Provider string         `json:"provider,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
}

// executeSuccessResponse defines the canonical success payload.
type executeSuccessResponse struct {
	Ok       bool   `json:"ok"`
	Skill    string `json:"skill"`
	Provider string `json:"provider,omitempty"`
	Result   any    `json:"result"`
}

// executeErrorResponse defines the canonical error payload.
type executeErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// executeHandler allows dynamic execution of skills via HTTP POST using JSON payloads.
func (s *Server) executeHandler(w http.ResponseWriter, r *http.Request) {
	var req executeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode execute request payload", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(executeErrorResponse{Ok: false, Error: "invalid json format"})
		return
	}

	if req.Skill == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(executeErrorResponse{Ok: false, Error: "missing skill"})
		return
	}

	if req.Params == nil {
		req.Params = map[string]any{}
	}

	if req.Provider != "" {
		if _, ok := req.Params["provider"]; !ok {
			req.Params["provider"] = req.Provider
		}
	}

	// We still use s.agent.Execute under the hood because ToolAdapter abstractly executes from the registry
	result, err := s.agent.Execute(r.Context(), req.Skill, req.Params)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.metrics.RecordSkillInvocation(req.Skill, "error")
		s.logger.Error("Failed to execute tool", err, "skill", req.Skill)

		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(executeErrorResponse{Ok: false, Error: err.Error()})
		return
	}

	s.metrics.RecordSkillInvocation(req.Skill, "success")
	w.WriteHeader(http.StatusOK)
	resp := executeSuccessResponse{
		Ok:       true,
		Skill:    req.Skill,
		Provider: req.Provider,
		Result:   result,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("Failed to encode execute response", err)
	}
}

type metricKey struct {
	skill  string
	status string
}

// Metrics stores runtime counters exposed through the Prometheus text format.
type Metrics struct {
	mu               sync.Mutex
	skillInvocations map[metricKey]int
}

// NewMetrics creates an empty runtime metrics store.
func NewMetrics() *Metrics {
	return &Metrics{skillInvocations: make(map[metricKey]int)}
}

// RecordSkillInvocation increments the skill invocation counter for a status.
func (m *Metrics) RecordSkillInvocation(skill, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.skillInvocations[metricKey{skill: skill, status: status}]++
}

// Prometheus renders the metrics store in Prometheus text exposition format.
func (m *Metrics) Prometheus() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var b strings.Builder
	b.WriteString("# HELP multix_skill_invocations_total Total skill invocation attempts through the runtime execute endpoint.\n")
	b.WriteString("# TYPE multix_skill_invocations_total counter\n")

	keys := make([]metricKey, 0, len(m.skillInvocations))
	for key := range m.skillInvocations {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].skill == keys[j].skill {
			return keys[i].status < keys[j].status
		}
		return keys[i].skill < keys[j].skill
	})

	for _, key := range keys {
		b.WriteString(`multix_skill_invocations_total{skill="`)
		b.WriteString(escapeMetricLabel(key.skill))
		b.WriteString(`",status="`)
		b.WriteString(escapeMetricLabel(key.status))
		b.WriteString(`"} `)
		b.WriteString(strconv.Itoa(m.skillInvocations[key]))
		b.WriteString("\n")
	}

	return b.String()
}

func escapeMetricLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

// Run starts the HTTP server and blocks until graceful shutdown or failure.
func (s *Server) Run(ctx context.Context) error {
	addr := ":" + strconv.Itoa(s.port)
	s.logger.Info("Starting MULTIX runtime", "port", s.port)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	s.logger.Info("Server listening", "addr", srv.Addr)

	select {
	case err := <-errCh:
		s.logger.Error("Server failed to start", err)
		return err
	case sig := <-stop:
		s.logger.Info("Shutting down server...", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server forced to shutdown", err)
		return err
	}

	s.logger.Info("Server exited properly")
	return nil
}
