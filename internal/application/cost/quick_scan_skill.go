// File: internal/application/cost/quick_scan_skill.go
// Purpose: Implements the cost.quick-scan skill — resource-count proxy as a v1 cost signal.

package cost

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// QuickScanSkill produces a fast cost signal by aggregating inventory counts
// per provider and resource type. It is explicitly a *resource-count proxy*,
// not a billing query — full Cost Explorer / Billing API integration is
// scoped to a follow-up issue (FOCUS-aligned cost.focus_report skill).
type QuickScanSkill struct {
	providers outbound.ProviderRegistry
}

// NewQuickScanSkill creates a new QuickScanSkill.
func NewQuickScanSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &QuickScanSkill{providers: pr}
}

func (s *QuickScanSkill) Name() string { return "cost.quick_scan" }
func (s *QuickScanSkill) Description() string {
	return "Fast cost signal: aggregates inventory resource counts per provider as a proxy. Real billing queries are scoped to cost.focus_report (#46)."
}
func (s *QuickScanSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func (s *QuickScanSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	requested := extractProviders(input)
	report := make([]map[string]any, 0, len(requested))
	grandTotal := 0
	for _, name := range requested {
		entry := map[string]any{"provider": name}
		p, err := s.providers.GetCloudInventoryProvider(name)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		summary, err := p.Scan(ctx)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		entry["ok"] = true
		entry["total"] = summary.Total
		entry["count_by_type"] = summary.CountByType
		grandTotal += summary.Total
		report = append(report, entry)
	}
	return map[string]any{
		"note":             "resource-count proxy; not a billing query (see #46 for cost.focus_report)",
		"grand_total":      grandTotal,
		"providers":        report,
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
