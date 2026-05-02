// File: internal/application/security/identity_posture_skill.go
// Purpose: Implements the security.identity-posture skill — flags risky principal types in active auth context.

package security

import (
	"context"
	"strings"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// IdentityPostureSkill validates the active identity per provider and emits
// posture findings (severity, category, message). v1 focuses on signals
// derivable from auth.Validate output: principal type, root/owner usage,
// missing tenancy/project context. Full IAM enumeration (users, roles,
// policies) is scoped to a follow-up issue.
type IdentityPostureSkill struct {
	providers outbound.ProviderRegistry
}

// NewIdentityPostureSkill creates a new IdentityPostureSkill.
func NewIdentityPostureSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &IdentityPostureSkill{providers: pr}
}

func (s *IdentityPostureSkill) Name() string { return "security.identity_posture" }
func (s *IdentityPostureSkill) Description() string {
	return "Reports identity posture findings (severity-tagged) for the active auth context across providers. v1 surfaces principal-type risks; deeper IAM enumeration is a follow-up."
}
func (s *IdentityPostureSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func (s *IdentityPostureSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	requested := extractProviders(input)
	allFindings := make([]map[string]any, 0)

	for _, name := range requested {
		p, err := s.providers.GetCloudAuthProvider(name)
		if err != nil {
			allFindings = append(allFindings, finding(name, "high", "auth_resolution",
				"Provider not registered or unsupported", err.Error()))
			continue
		}
		identity, err := p.Whoami(ctx)
		if err != nil {
			allFindings = append(allFindings, finding(name, "high", "whoami",
				"Could not resolve active identity", err.Error()))
			continue
		}

		principal := strings.ToLower(identity.Principal)
		switch {
		case strings.Contains(principal, ":root"):
			allFindings = append(allFindings, finding(name, "critical", "root_usage",
				"AWS root user is active — never use root for day-to-day operations", identity.Principal))
		case identity.PrincipalType == "user" && name == "aws":
			allFindings = append(allFindings, finding(name, "medium", "long_lived_user",
				"AWS access via IAM user; prefer IAM Identity Center / SSO with assumed roles", identity.Principal))
		case identity.PrincipalType == "instance_principal":
			allFindings = append(allFindings, finding(name, "info", "instance_principal",
				"Authenticated via instance principal", identity.Principal))
		case identity.PrincipalType == "" || identity.PrincipalType == "unknown":
			allFindings = append(allFindings, finding(name, "low", "unknown_principal_type",
				"Active principal type could not be determined", identity.Principal))
		default:
			allFindings = append(allFindings, finding(name, "info", "active_principal",
				"Active principal type: "+identity.PrincipalType, identity.Principal))
		}
	}

	severityCounts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0}
	for _, f := range allFindings {
		sev, _ := f["severity"].(string)
		severityCounts[sev]++
	}

	return map[string]any{
		"findings":        allFindings,
		"severity_counts": severityCounts,
	}, nil
}

func finding(provider, severity, category, message, evidence string) map[string]any {
	return map[string]any{
		"provider": provider,
		"severity": severity,
		"category": category,
		"message":  message,
		"evidence": evidence,
	}
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
