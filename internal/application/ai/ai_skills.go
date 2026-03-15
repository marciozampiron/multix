// File: internal/application/ai/ai_skills.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements capabilities that leverage AI providers.

package ai

import (
	"context"
	"multix/internal/domain/ai"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
	"time"
)

// ExplainSkill queries an AI backend to explain a cloud infrastructure resource or error.
type ExplainSkill struct {
	providers outbound.ProviderRegistry
}

// NewExplainSkill creates a new ExplainSkill.
func NewExplainSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &ExplainSkill{providers: pr}
}
func (s *ExplainSkill) Name() string { return "ai.explain" }
func (s *ExplainSkill) Description() string {
	return "Queries an AI backend to explain a cloud infrastructure resource or error."
}
func (s *ExplainSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{"type": "string"},
			"prompt":   map[string]any{"type": "string", "description": "Text or JSON to explain"},
		},
		"required": []string{"provider", "prompt"},
	}
}
func (s *ExplainSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetAIProvider(providerName)
	if err != nil {
		return nil, err
	}

	text, _ := input["prompt"].(string)
	prompt := ai.Prompt{Text: text, CreatedAt: time.Now()}
	resp, err := p.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"explanation": resp.Text,
		"provider":    resp.ProviderName,
		"tokens":      resp.TokensUsed,
	}, nil
}
