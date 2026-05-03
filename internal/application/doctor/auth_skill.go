// File: internal/application/doctor/auth_skill.go
// Purpose: Implements the doctor.auth skill — checks authentication health across all configured providers.

package doctor

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// AuthSkill validates authentication against one or many providers and returns
// structured findings (provider, ok, account, principal, error).
type AuthSkill struct {
	providers outbound.ProviderRegistry
}

// NewAuthSkill creates a new AuthSkill.
func NewAuthSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &AuthSkill{providers: pr}
}

func (s *AuthSkill) Name() string { return "doctor.auth" }
func (s *AuthSkill) Description() string {
	return "Validates authentication context for one or more cloud providers and reports per-provider health."
}
func (s *AuthSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional list of providers (aws, gcp, oci). If omitted, checks all three.",
			},
		},
	}
}

func (s *AuthSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	requested := extractProviders(input)
	findings := make([]map[string]any, 0, len(requested))
	healthy := 0

	for _, name := range requested {
		f := map[string]any{"provider": name}
		p, err := s.providers.GetCloudAuthProvider(name)
		if err != nil {
			f["ok"] = false
			f["error"] = err.Error()
			findings = append(findings, f)
			continue
		}
		result, err := p.Validate(ctx)
		if err != nil {
			f["ok"] = false
			f["error"] = err.Error()
			findings = append(findings, f)
			continue
		}
		f["ok"] = result.Valid
		f["account"] = result.AccountID
		f["principal"] = result.Principal
		f["message"] = result.Message
		if result.Valid {
			healthy++
		}
		findings = append(findings, f)
	}

	return map[string]any{
		"checked":  len(requested),
		"healthy":  healthy,
		"findings": findings,
	}, nil
}

func extractProviders(input map[string]any) []string {
	defaults := []string{"aws", "gcp", "oci"}
	raw, ok := input["providers"]
	if !ok || raw == nil {
		return defaults
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return defaults
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return defaults
	}
	return out
}
