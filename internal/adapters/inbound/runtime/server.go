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
	"strconv"
	"syscall"
	"time"

	"multix/internal/adapters/inbound/agent"
	domainSkills "multix/internal/domain/skills"
	"multix/internal/platform/logger"
)

// Server represents the local MULTIX HTTP runtime.
type Server struct {
	logger   logger.Logger
	registry *domainSkills.Registry
	mux      *http.ServeMux
	port     int
}

// NewServer initializes a new runtime Server.
func NewServer(logger logger.Logger, registry *domainSkills.Registry, port int) *Server {
	s := &Server{
		logger:   logger,
		registry: registry,
		mux:      http.NewServeMux(),
		port:     port,
	}
	s.registerRoutes()
	return s
}

// registerRoutes wires up all HTTP endpoints.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", s.healthHandler)
	s.mux.HandleFunc("/tools", s.toolsHandler)
}

// healthHandler provides a basic liveness probe.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"multix","mode":"runtime"}`))
}

// toolsHandler dynamically returns agent tool manifests based on registered skills.
func (s *Server) toolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	manifests := agent.GenerateManifests(s.registry)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manifests); err != nil {
		s.logger.Error("Failed to encode tool manifests", err)
	}
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
