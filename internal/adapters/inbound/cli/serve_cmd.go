// File: internal/adapters/inbound/cli/serve_cmd.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Exposes the agent runtime HTTP server via CLI command.

package cli

import (
	"multix/internal/adapters/inbound/agent"
	"multix/internal/adapters/inbound/runtime"
	"multix/internal/platform/logger"

	"github.com/spf13/cobra"
)

// NewServeCmd initializes the 'serve' command.
func NewServeCmd(logger logger.Logger, adapter *agent.ToolAdapter) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MULTIX skill runtime server",
		Long:  "Starts a local HTTP server exposing agent tools, execution endpoints, and capabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := runtime.NewServer(logger, adapter, port)
			return server.Run(cmd.Context())
		},
	}
	
	cmd.Flags().IntVar(&port, "port", 8080, "Port to assign to the HTTP runtime server")
	return cmd
}
