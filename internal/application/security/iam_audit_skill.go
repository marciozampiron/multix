// File: internal/application/security/iam_audit_skill.go
// Company: Hassan
// Creator: Zamp
// Created: 03/05/2026
// Updated: 03/05/2026
// Purpose: security.iam_audit — identity posture findings + AI-generated IAM remediation.

package security

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"multix/internal/domain/ai"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

const iamDefaultAIProvider = "gemini"

// IAMAuditSkill builds on the identity-posture rules (root usage, long-lived
// IAM users, etc.) and asks an AIProvider to translate the findings into
// remediation language an operator can act on. v1 reuses Whoami/Validate
// signals — full IAM enumeration (users, roles, policies, attached
// permissions) is out of scope and tracked separately.
type IAMAuditSkill struct {
	providers outbound.ProviderRegistry
}

// NewIAMAuditSkill creates the skill bound to the registry.
func NewIAMAuditSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &IAMAuditSkill{providers: pr}
}

func (s *IAMAuditSkill) Name() string { return "security.iam_audit" }

func (s *IAMAuditSkill) Description() string {
	return "Multi-cloud IAM audit with AI-generated remediation. v1 evaluates the active principal (root usage, long-lived IAM users, unknown principal types) and asks the AI for least-privilege rewrites. Full IAM enumeration of users/roles/policies is a follow-up."
}

func (s *IAMAuditSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Subset of providers to audit. Defaults to [aws, gcp, oci]."},
			"ai_provider": map[string]any{"type": "string", "description": "AI backend for remediation. Defaults to gemini."},
		},
	}
}

type iamFinding struct {
	Provider  string `json:"provider"`
	Severity  string `json:"severity"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	Principal string `json:"principal,omitempty"`
}

func (s *IAMAuditSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providers := extractProviders(input)
	aiProviderName, _ := input["ai_provider"].(string)
	if aiProviderName == "" {
		aiProviderName = iamDefaultAIProvider
	}

	var allFindings []iamFinding
	for _, name := range providers {
		auth, err := s.providers.GetCloudAuthProvider(name)
		if err != nil {
			allFindings = append(allFindings, iamFinding{
				Provider: name, Severity: "high", Category: "provider_resolution",
				Message: fmt.Sprintf("Provider %s not registered: %v", name, err),
			})
			continue
		}

		identity, err := auth.Whoami(ctx)
		if err != nil {
			allFindings = append(allFindings, iamFinding{
				Provider: name, Severity: "high", Category: "whoami",
				Message: fmt.Sprintf("Could not resolve active identity: %v", err),
			})
			continue
		}
		allFindings = append(allFindings, classifyIAMIdentity(name, identity.Principal, identity.PrincipalType)...)
	}

	remediation := ""
	if len(allFindings) > 0 {
		aiProvider, err := s.providers.GetAIProvider(aiProviderName)
		if err != nil {
			remediation = fmt.Sprintf("(AI provider %q unavailable: %v)", aiProviderName, err)
		} else {
			prompt := buildIAMRemediationPrompt(allFindings)
			resp, err := aiProvider.Generate(ctx, ai.Prompt{Text: prompt, CreatedAt: time.Now()})
			if err != nil {
				remediation = fmt.Sprintf("(AI generation failed: %v)", err)
			} else {
				remediation = resp.Text
			}
		}
	}

	severityCounts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0}
	for _, f := range allFindings {
		severityCounts[f.Severity]++
	}

	return map[string]any{
		"findings":        allFindings,
		"severity_counts": severityCounts,
		"remediation":     remediation,
		"ai_provider":     aiProviderName,
	}, nil
}

// classifyIAMIdentity applies the same rules as security.identity_posture but
// returns iam-shaped findings (with a Principal field). Keeping the logic
// duplicated rather than imported avoids cyclic skill-to-skill calls; the
// rule set is small and rarely changes.
func classifyIAMIdentity(provider, principal, principalType string) []iamFinding {
	lower := strings.ToLower(principal)
	switch {
	case strings.Contains(lower, ":root"):
		return []iamFinding{{
			Provider: provider, Severity: "critical", Category: "root_usage",
			Message:   "AWS root user is active — never use root for day-to-day operations. Move to IAM Identity Center or assumed roles.",
			Principal: principal,
		}}
	case principalType == "user" && provider == "aws":
		return []iamFinding{{
			Provider: provider, Severity: "medium", Category: "long_lived_user",
			Message:   "AWS access via IAM user (long-lived credentials). Migrate to IAM Identity Center / SSO with assumed roles.",
			Principal: principal,
		}}
	case principalType == "instance_principal":
		return []iamFinding{{
			Provider: provider, Severity: "info", Category: "instance_principal",
			Message:   "Authenticated via instance principal — preferred for in-cloud workloads.",
			Principal: principal,
		}}
	case principalType == "service_account":
		return []iamFinding{{
			Provider: provider, Severity: "info", Category: "service_account",
			Message:   "Authenticated via service account.",
			Principal: principal,
		}}
	case principalType == "" || principalType == "unknown":
		return []iamFinding{{
			Provider: provider, Severity: "low", Category: "unknown_principal_type",
			Message:   "Active principal type could not be determined; cannot assess privilege model.",
			Principal: principal,
		}}
	default:
		return []iamFinding{{
			Provider: provider, Severity: "info", Category: "active_principal",
			Message:   fmt.Sprintf("Active principal type: %s", principalType),
			Principal: principal,
		}}
	}
}

func buildIAMRemediationPrompt(findings []iamFinding) string {
	var b strings.Builder
	b.WriteString("You are a cloud IAM specialist. Given the following identity findings, propose least-privilege remediation steps. Keep it under 200 words. Findings:\n")
	sorted := make([]iamFinding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Provider != sorted[j].Provider {
			return sorted[i].Provider < sorted[j].Provider
		}
		return sorted[i].Severity < sorted[j].Severity
	})
	for _, f := range sorted {
		b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", f.Severity, f.Provider, f.Message))
	}
	return b.String()
}
