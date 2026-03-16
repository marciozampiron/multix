// File: internal/adapters/inbound/cli/serve_cmd.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Exposes the agent runtime HTTP server via CLI command.

package cli

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"multix/internal/platform/logger"

	"github.com/spf13/cobra"
)

// NewServeCmd initializes the 'serve' command.
func NewServeCmd(logger logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MULTIX skill runtime server",
		Long:  "Starts a local HTTP server exposing agent tools, execution endpoints, and capabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(logger)
		},
	}
	return cmd
}

func runServe(logger logger.Logger) error {
	port := "8080"
	logger.Info("Starting MULTIX runtime", "port", port)

	mux := http.NewServeMux()

	// Base endpoints to be implemented in subsequent issues
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine so that it doesn't block
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	logger.Info("Server listening", "addr", srv.Addr)

	// Block until a signal is received or an error occurs
	select {
	case err := <-errCh:
		logger.Error("Server failed to start", err)
		return err
	case sig := <-stop:
		logger.Info("Shutting down server...", "signal", sig.String())
	}

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err)
		return err
	}

	logger.Info("Server exited properly")
	return nil
}
