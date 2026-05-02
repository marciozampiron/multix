// File: internal/application/landingzone/audit_skill.go
// Purpose: Implements the landingzone.audit skill — baseline checks for multi-account/multi-region landing zone health.

package landingzone

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// AuditSkill performs first-pass landing zone validation by combining auth identity
// (account/tenancy/project ID) with inventory presence (multi-region resource spread).
// Deeper org-level checks (AWS Organizations, GCP Folders, OCI compartment trees) are
// scoped to follow-up issues; this skill establishes the contract and surface.
type AuditSkill struct {
	providers outbound.ProviderRegistry
}

// NewAuditSkill creates a new AuditSkill.
func NewAuditSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &AuditSkill{providers: pr}
}

func (s *AuditSkill) Name() string { return "landingzone.audit" }
func (s *AuditSkill) Description() string {
	return "Baseline landing zone audit: confirms identity context and reports multi-region resource spread per provider. Deeper org/folder/compartment trees are deferred to follow-up checks."
}
func (s *AuditSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func (s *AuditSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	requested := extractProviders(input)
	report := make([]map[string]any, 0, len(requested))
	for _, name := range requested {
		entry := map[string]any{"provider": name}

		auth, err := s.providers.GetCloudAuthProvider(name)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		validation, err := auth.Validate(ctx)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		entry["account"] = validation.AccountID
		entry["principal"] = validation.Principal

		// Inventory spread: count distinct regions across resources.
		inv, err := s.providers.GetCloudInventoryProvider(name)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		resources, err := inv.List(ctx, "")
		if err != nil {
			entry["ok"] = true
			entry["inventory_error"] = err.Error()
			report = append(report, entry)
			continue
		}
		regions := map[string]struct{}{}
		for _, r := range resources {
			if r.Region != "" {
				regions[r.Region] = struct{}{}
			}
		}
		entry["ok"] = true
		entry["resource_count"] = len(resources)
		entry["distinct_regions"] = len(regions)
		entry["multi_region"] = len(regions) > 1
		entry["regions"] = mapKeys(regions)
		report = append(report, entry)
	}
	return map[string]any{"providers": report}, nil
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

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
