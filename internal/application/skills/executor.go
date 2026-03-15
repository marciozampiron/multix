package skills

import (
	"context"
	"fmt"

	domainSkills "multix/internal/domain/skills"
)

// Executor manages the lifecycle of invoking a universal Skill.
type Executor struct {
	registry *domainSkills.Registry
}

func NewExecutor(registry *domainSkills.Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute invokes a skill by its name using a JSON-like payload.
func (e *Executor) Execute(ctx context.Context, skillName string, input map[string]any) (any, error) {
	skill, err := e.registry.Get(skillName)
	if err != nil {
		return nil, err
	}

	// Logging/Metrics specific to Agent Tool Execution could be injected here.
	result, err := skill.Execute(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute skill '%s': %w", skillName, err)
	}

	return result, nil
}
