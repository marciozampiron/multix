package agent

import (
	"context"

	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
)

type toolHandler struct {
	registry *domainSkills.Registry
	executor *appSkills.Executor
}

func NewToolHandler(registry *domainSkills.Registry, executor *appSkills.Executor) *toolHandler {
	return &toolHandler{
		registry: registry,
		executor: executor,
	}
}

// ExecuteTool allows an AI agent to execute a skill using raw parsed JSON arguments.
// This is exactly what Gemini/OpenAI provides when a Function is Invoked.
func (h *toolHandler) ExecuteTool(ctx context.Context, toolName string, arguments map[string]any) (any, error) {
	return h.executor.Execute(ctx, toolName, arguments)
}

// GetTools returns the JSON Schema for all registered tools so the LLM knows what they do.
func (h *toolHandler) GetTools() []any {
	var schemas []any

	for _, skill := range h.registry.ListAll() {
		schemas = append(schemas, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        skill.Name(),
				"description": skill.Description(),
				"parameters":  skill.InputSchema(),
			},
		})
	}
	return schemas
}
