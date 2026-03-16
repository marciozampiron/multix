// File: internal/adapters/inbound/cli/serve_cmd.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Exposes the agent runtime HTTP server via CLI command.

package cli

import (
	"context"

	"multix/internal/adapters/inbound/runtime"
	domainSkills "multix/internal/domain/skills"
	"multix/internal/platform/logger"

	"github.com/spf13/cobra"
)

// NewServeCmd initializes the 'serve' command.
func NewServeCmd(logger logger.Logger, registry *domainSkills.Registry) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MULTIX skill runtime server",
		Long:  "Starts a local HTTP server exposing agent tools, execution endpoints, and capabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := runtime.NewServer(logger, registry, port)
			return server.Run(context.Background())
		},
	}
	
	cmd.Flags().IntVar(&port, "port", 8080, "Port to assign to the HTTP runtime server")
	return cmd
}
