// File: internal/application/auth/auth_skills.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements authentication skills for cloud providers.

package auth

import (
	"context"
	"strings"

	domainAuth "multix/internal/domain/auth"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

const defaultCloudProvider = "aws"

// ValidateSkill validates if the user's current cloud provider credentials are valid and active.
type ValidateSkill struct {
	providers outbound.ProviderRegistry
}

// NewValidateSkill creates a new ValidateSkill.
func NewValidateSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &ValidateSkill{providers: pr}
}

func (s *ValidateSkill) Name() string { return "auth.validate" }
func (s *ValidateSkill) Description() string {
	return "Validates if the user's current cloud provider credentials are valid and active."
}
func (s *ValidateSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{
				"type":        "string",
				"description": "Cloud provider name (aws, gcp, azure, etc.)",
			},
		},
	}
}

func (s *ValidateSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName := resolveProvider(input)
	p, err := s.providers.GetCloudAuthProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := p.Validate(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// WhoamiSkill returns the active identity information, IAM role or account ID for the provider.
type WhoamiSkill struct {
	providers outbound.ProviderRegistry
}

// NewWhoamiSkill creates a new WhoamiSkill.
func NewWhoamiSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &WhoamiSkill{providers: pr}
}

func (s *WhoamiSkill) Name() string { return "auth.whoami" }
func (s *WhoamiSkill) Description() string {
	return "Returns the active identity information, IAM role or account ID for the provider."
}
func (s *WhoamiSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{"type": "string"},
		},
	}
}

func (s *WhoamiSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName := resolveProvider(input)
	p, err := s.providers.GetCloudAuthProvider(providerName)
	if err != nil {
		return nil, err
	}

	identity, valErr := p.Whoami(ctx)
	if valErr != nil {
		return nil, valErr
	}
	return identity, nil
}

// LoginSkill authenticates the machine to a specific provider.
type LoginSkill struct {
	providers outbound.ProviderRegistry
}

// NewLoginSkill creates a new LoginSkill.
func NewLoginSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &LoginSkill{providers: pr}
}
func (s *LoginSkill) Name() string        { return "auth.login" }
func (s *LoginSkill) Description() string { return "Authenticates the machine to a specific provider." }
func (s *LoginSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]string{"type": "string"},
		},
		"required": []string{"provider"},
	}
}
func (s *LoginSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetCloudAuthProvider(providerName)
	if err != nil {
		return nil, err
	}

	session, valErr := p.Login(ctx, domainAuth.Credentials{})
	return map[string]any{"provider": session.Provider, "valid": session.IsValid}, valErr
}

func resolveProvider(input map[string]any) string {
	providerName, _ := input["provider"].(string)
	if strings.TrimSpace(providerName) == "" {
		return defaultCloudProvider
	}
	return providerName
}
