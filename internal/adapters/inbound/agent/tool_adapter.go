// File: internal/adapters/inbound/agent/tool_adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Exposes application skills as tool-like interfaces for AI agents.

package agent

import (
	"context"
	"fmt"
	"strings"

	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
)

// ToolAdapter exposes the underlying skills to AI agents without coupling to a specific vendor SDK.
type ToolAdapter struct {
	registry *domainSkills.Registry
	executor *appSkills.Executor
}

// NewToolAdapter creates a new ToolAdapter.
func NewToolAdapter(registry *domainSkills.Registry, executor *appSkills.Executor) *ToolAdapter {
	return &ToolAdapter{
		registry: registry,
		executor: executor,
	}
}

// Execute allows an AI agent to execute a skill using raw parsed arguments.
// This keeps tool execution decoupled from specific SDKs and agent protocols.
func (a *ToolAdapter) Execute(ctx context.Context, toolName string, arguments map[string]any) (any, error) {
	if strings.TrimSpace(toolName) == "" {
		return nil, fmt.Errorf("tool name is required")
	}

	if arguments == nil {
		arguments = map[string]any{}
	}

	return a.executor.Execute(ctx, toolName, arguments)
}

// Manifests returns tool manifests for all registered skills.
func (a *ToolAdapter) Manifests() []Manifest {
	return GenerateManifests(a.registry)
}
