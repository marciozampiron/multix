// File: internal/application/doctor/k8s_skill.go
// Purpose: Implements the doctor.k8s skill — reports cluster health across providers.

package doctor

import (
	"context"
	"strings"

	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// K8sHealthSkill lists clusters per provider and classifies each as healthy/unhealthy
// based on lifecycle state. It does not connect to the cluster control plane — that
// would require kubeconfig sync and is deferred to a deeper diagnostic skill.
type K8sHealthSkill struct {
	providers outbound.ProviderRegistry
}

// NewK8sHealthSkill creates a new K8sHealthSkill.
func NewK8sHealthSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &K8sHealthSkill{providers: pr}
}

func (s *K8sHealthSkill) Name() string { return "doctor.k8s" }
func (s *K8sHealthSkill) Description() string {
	return "Lists managed Kubernetes clusters per provider and classifies them as healthy or degraded based on lifecycle state."
}
func (s *K8sHealthSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
	}
}

func (s *K8sHealthSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	requested := extractProviders(input)
	report := make([]map[string]any, 0, len(requested))
	totalClusters, totalHealthy := 0, 0

	for _, name := range requested {
		entry := map[string]any{"provider": name}
		p, err := s.providers.GetKubernetesProvider(name)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		clusters, err := p.ListClusters(ctx)
		if err != nil {
			entry["ok"] = false
			entry["error"] = err.Error()
			report = append(report, entry)
			continue
		}
		clusterRows := make([]map[string]any, 0, len(clusters))
		healthy := 0
		for _, c := range clusters {
			h := isHealthyClusterStatus(c.Status)
			if h {
				healthy++
			}
			clusterRows = append(clusterRows, map[string]any{
				"name":       c.Name,
				"region":     c.Region,
				"version":    c.Version,
				"status":     c.Status,
				"healthy":    h,
				"node_count": c.NodeCount,
			})
		}
		totalClusters += len(clusters)
		totalHealthy += healthy
		entry["ok"] = true
		entry["count"] = len(clusters)
		entry["healthy"] = healthy
		entry["clusters"] = clusterRows
		report = append(report, entry)
	}

	return map[string]any{
		"checked_providers": len(requested),
		"total_clusters":    totalClusters,
		"total_healthy":     totalHealthy,
		"providers":         report,
	}, nil
}

// isHealthyClusterStatus maps provider-specific cluster statuses to a boolean.
// AWS EKS: ACTIVE; GCP GKE: RUNNING; OCI OKE: ACTIVE.
func isHealthyClusterStatus(s string) bool {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "ACTIVE", "RUNNING":
		return true
	default:
		return false
	}
}
