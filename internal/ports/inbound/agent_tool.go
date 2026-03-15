package inbound

import "context"

// AgentToolHandler defines the interface for an AI Agent to execute a capability.
type AgentToolHandler interface {
	// ExecuteTool is called by the LLM function calling mechanism.
	ExecuteTool(ctx context.Context, toolName string, arguments map[string]any) (any, error)

	// GetTools returns the OpenAPI schemas of all registered skills to feed the LLM.
	GetTools() []any
}
