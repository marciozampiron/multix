// File: internal/adapters/inbound/agent/tool_adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Exposes skills as tools for AI agents.

package agent

import (
	"context"

	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
)

// ToolAdapter exposes the underlying skills to AI agents (e.g. Gemini, OpenAI, MCP).
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

// Execute allows an AI agent to execute a skill using raw parsed JSON arguments.
// This decouples the execution from specific SDKs, acting as a universal tool interface.
func (a *ToolAdapter) Execute(ctx context.Context, toolName string, arguments map[string]any) (any, error) {
	return a.executor.Execute(ctx, toolName, arguments)
}

// Manifests returns the JSON Schema for all registered tools.
func (a *ToolAdapter) Manifests() []Manifest {
	return GenerateManifests(a.registry)
}
