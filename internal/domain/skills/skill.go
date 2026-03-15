// File: internal/domain/skills/skill.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Defines the universal Skill contract.

package skills

import "context"

// Skill defines the universal contract that all CLI capabilities must satisfy
// so they can be seamlessly consumed by an AI Agent as a "tool" or by the
// terminal handler directly.
type Skill interface {
	// Name serves as the unique identifier for registry and Agent Tool Calling (e.g., "inventory.scan")
	Name() string

	// Description is the natural language explanation consumed by the LLM
	// to decide when and how to call this tool.
	Description() string

	// InputSchema returns a JSON Schema representation of what the Execute function expects.
	// This maps exactly to the OpenAPI/JSONSchema format used by Gemini and OpenAI.
	InputSchema() any

	// Execute is the universal actuator. It receives a loosely typed map (usually decoded from JSON)
	// and returns a struct or primitive that can be marshaled back into standard JSON.
	Execute(ctx context.Context, input map[string]any) (any, error)
}
