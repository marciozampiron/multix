// File: internal/application/k8s/k8s_skills.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements skills to manage and list Kubernetes clusters.

package k8s

import (
	"context"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound" // Dynamic Provider Resolution
)

// ListClustersSkill lists all managed Kubernetes clusters.
type ListClustersSkill struct {
	providers outbound.ProviderRegistry
}

// NewListClustersSkill creates a new ListClustersSkill.
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
	p, err := s.providers.GetKubernetesProvider(providerName)
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
