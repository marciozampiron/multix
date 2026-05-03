// File: internal/application/security/k8s_audit_skill.go
// Company: Hassan
// Creator: Zamp
// Created: 03/05/2026
// Updated: 03/05/2026
// Purpose: security.k8s_audit — heuristic posture findings + AI-generated remediation.

package security

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"multix/internal/domain/ai"
	"multix/internal/domain/k8s"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

const k8sDefaultAIProvider = "gemini"

// K8sAuditSkill produces severity-tagged posture findings about managed
// Kubernetes clusters and asks an AIProvider for plain-English remediation
// suggestions. v1 is **structural** — it does NOT connect to the cluster
// control plane (no kubeconfig sync, no Trivy scan). Real CVE scanning is a
// follow-up captured in the skill description and the catalog.
type K8sAuditSkill struct {
	providers outbound.ProviderRegistry
}

// NewK8sAuditSkill creates the skill bound to the registry.
func NewK8sAuditSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &K8sAuditSkill{providers: pr}
}

func (s *K8sAuditSkill) Name() string { return "security.k8s_audit" }

func (s *K8sAuditSkill) Description() string {
	return "Posture audit for managed Kubernetes clusters (EKS/GKE/OKE) with AI-generated remediation. v1 uses ListClusters signals (version drift, lifecycle state, node count); real CVE scanning via Trivy is a follow-up."
}

func (s *K8sAuditSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Subset of providers to audit. Defaults to [aws, gcp, oci]."},
			"ai_provider": map[string]any{"type": "string", "description": "AI backend for remediation. Defaults to gemini."},
		},
	}
}

type k8sFinding struct {
	Provider string `json:"provider"`
	Cluster  string `json:"cluster"`
	Region   string `json:"region"`
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

func (s *K8sAuditSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providers := extractProviders(input)
	aiProviderName, _ := input["ai_provider"].(string)
	if aiProviderName == "" {
		aiProviderName = k8sDefaultAIProvider
	}

	var allFindings []k8sFinding
	clustersByProvider := map[string]int{}

	for _, name := range providers {
		k8sProvider, err := s.providers.GetKubernetesProvider(name)
		if err != nil {
			allFindings = append(allFindings, k8sFinding{
				Provider: name, Severity: "high", Category: "provider_resolution",
				Message: fmt.Sprintf("Provider %s not registered: %v", name, err),
			})
			continue
		}
		clusters, err := k8sProvider.ListClusters(ctx)
		if err != nil {
			allFindings = append(allFindings, k8sFinding{
				Provider: name, Severity: "high", Category: "list_clusters",
				Message: fmt.Sprintf("ListClusters failed: %v", err),
			})
			continue
		}
		clustersByProvider[name] = len(clusters)
		for _, c := range clusters {
			allFindings = append(allFindings, classifyCluster(name, c)...)
		}
	}

	// Ask AI for remediation only when we have findings to remediate.
	remediation := ""
	if len(allFindings) > 0 {
		aiProvider, err := s.providers.GetAIProvider(aiProviderName)
		if err != nil {
			// AI failure is not fatal — surface findings without remediation.
			remediation = fmt.Sprintf("(AI provider %q unavailable: %v)", aiProviderName, err)
		} else {
			prompt := buildK8sRemediationPrompt(allFindings)
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
		"findings":             allFindings,
		"severity_counts":      severityCounts,
		"clusters_by_provider": clustersByProvider,
		"remediation":          remediation,
		"ai_provider":          aiProviderName,
	}, nil
}

// classifyCluster applies posture heuristics to a single cluster and returns
// zero or more findings.
func classifyCluster(provider string, c *k8s.Cluster) []k8sFinding {
	var out []k8sFinding
	state := strings.ToUpper(strings.TrimSpace(c.Status))

	if state != "ACTIVE" && state != "RUNNING" {
		out = append(out, k8sFinding{
			Provider: provider, Cluster: c.Name, Region: c.Region,
			Severity: "high", Category: "lifecycle",
			Message: fmt.Sprintf("Cluster lifecycle state %q is not steady-state; control-plane may be mid-operation.", c.Status),
		})
	}

	if strings.TrimSpace(c.Version) == "" {
		out = append(out, k8sFinding{
			Provider: provider, Cluster: c.Name, Region: c.Region,
			Severity: "medium", Category: "version_unknown",
			Message: "Cluster version not reported by provider; cannot evaluate freshness.",
		})
	}

	// Autopilot/managed-node-group clusters legitimately report node_count=0;
	// flag only as info so the agent can decide whether to investigate.
	if c.NodeCount == 0 {
		out = append(out, k8sFinding{
			Provider: provider, Cluster: c.Name, Region: c.Region,
			Severity: "info", Category: "zero_nodes",
			Message: "Cluster reports node_count=0 (autopilot/serverless or empty cluster).",
		})
	}

	return out
}

func buildK8sRemediationPrompt(findings []k8sFinding) string {
	var b strings.Builder
	b.WriteString("You are a Kubernetes platform engineer. Given the following posture findings, propose concrete remediation steps. Keep it under 200 words. Findings:\n")
	// Stable ordering for deterministic prompts.
	sorted := make([]k8sFinding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Provider != sorted[j].Provider {
			return sorted[i].Provider < sorted[j].Provider
		}
		return sorted[i].Cluster < sorted[j].Cluster
	})
	for _, f := range sorted {
		b.WriteString(fmt.Sprintf("- [%s] %s/%s: %s\n", f.Severity, f.Provider, f.Cluster, f.Message))
	}
	return b.String()
}
