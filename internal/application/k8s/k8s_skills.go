package k8s

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound" // Dynamic Provider Resolution
)

type ListClustersSkill struct {
	providers outbound.ProviderRegistry
}

func NewListClustersSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &ListClustersSkill{providers: pr}
}

func (s *ListClustersSkill) Name() string { return "k8s.list_clusters" }
func (s *ListClustersSkill) Description() string {
	return "Lists all managed Kubernetes clusters."
}
func (s *ListClustersSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{"type": "string"},
		},
		"required": []string{"provider"},
	}
}
func (s *ListClustersSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providerName, _ := input["provider"].(string)
	p, err := s.providers.GetK8s(providerName)
	if err != nil {
		return nil, err
	}

	clusters, err := p.ListClusters(ctx)
	if err != nil {
		return nil, err
	}
	var res []map[string]any
	for _, c := range clusters {
		res = append(res, map[string]any{
			"id":         c.ID,
			"name":       c.Name,
			"region":     c.Region,
			"version":    c.Version,
			"node_count": c.NodeCount,
			"status":     c.Status,
		})
	}
	return map[string]any{"clusters": res}, nil
}
