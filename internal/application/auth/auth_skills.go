package auth

import (
	"context"

	domainAuth "multix/internal/domain/auth"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound" // Import for ProviderRegistry
)

type ValidateSkill struct {
	providers outbound.ProviderRegistry
}

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
		"required": []string{"provider"},
	}
}

func (s *ValidateSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetCloudAuthProvider(providerName)
	if err != nil {
		return nil, err
	}

	valid, valErr := p.Validate(ctx)
	if valErr != nil {
		return map[string]any{"is_valid": false, "reason": valErr.Error()}, nil
	}
	return map[string]any{"is_valid": valid, "reason": ""}, nil
}

type WhoamiSkill struct {
	providers outbound.ProviderRegistry
}

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
		"required": []string{"provider"},
	}
}

func (s *WhoamiSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetCloudAuthProvider(providerName)
	if err != nil {
		return nil, err
	}

	session, valErr := p.Whoami(ctx)
	if valErr != nil {
		return nil, valErr
	}
	return map[string]any{
		"account_id": session.AccountID,
		"username":   session.Username,
		"role":       session.Role,
	}, nil
}

type LoginSkill struct {
	providers outbound.ProviderRegistry
}

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
