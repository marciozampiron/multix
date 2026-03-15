package inventory

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound" // Dynamic Provider Resolution
)

type ScanSkill struct {
	providers outbound.ProviderRegistry
}

func NewScanSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &ScanSkill{providers: pr}
}

func (s *ScanSkill) Name() string { return "inventory.scan" }
func (s *ScanSkill) Description() string {
	return "Scans the cloud provider to discover active infrastructure resources filtered by type."
}
func (s *ScanSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{"type": "string"},
			"service":  map[string]any{"type": "string", "description": "e.g., compute, storage, rds"},
		},
		"required": []string{"provider", "service"},
	}
}

func (s *ScanSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetInventory(providerName)
	if err != nil {
		return nil, err
	}

	service, _ := input["service"].(string)
	resources, err := p.List(ctx, service)
	if err != nil {
		return nil, err
	}

	var parsed []map[string]any
	for _, r := range resources {
		parsed = append(parsed, map[string]any{
			"id":     r.ID,
			"type":   r.Type,
			"region": r.Region,
			"name":   r.Name,
			"status": r.Status,
		})
	}

	return map[string]any{"resources": parsed, "count": len(parsed)}, nil
}

type SummarySkill struct {
	providers outbound.ProviderRegistry
}

func NewSummarySkill(pr outbound.ProviderRegistry) skills.Skill {
	return &SummarySkill{providers: pr}
}
func (s *SummarySkill) Name() string        { return "inventory.summary" }
func (s *SummarySkill) Description() string { return "Returns an aggregated count of all discoveries." }
func (s *SummarySkill) InputSchema() any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"provider": map[string]any{"type": "string"}},
		"required":   []string{"provider"},
	}
}
func (s *SummarySkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetInventory(providerName)
	if err != nil {
		return nil, err
	}

	summary, err := p.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"provider":      summary.ProviderName,
		"total_count":   summary.Total,
		"count_by_type": summary.CountByType,
	}, nil
}
