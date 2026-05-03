// File: internal/application/security/k8s_audit_skill_test.go
// Purpose: Unit tests for security.k8s_audit — finding classification, AI failure tolerance.

package security

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/ai"
	"multix/internal/domain/k8s"
	"multix/internal/ports/outbound"
)

type fakeK8s struct {
	clusters []*k8s.Cluster
	err      error
}

func (f *fakeK8s) ID() string { return "fake" }
func (f *fakeK8s) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	return f.clusters, f.err
}
func (f *fakeK8s) SyncContext(ctx context.Context, name, region string) error { return nil }

type fakeAIK8s struct {
	text string
	err  error
}

func (f *fakeAIK8s) ID() string { return "gemini" }
func (f *fakeAIK8s) Generate(ctx context.Context, p ai.Prompt) (*ai.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &ai.Response{Text: f.text, ProviderName: "gemini"}, nil
}
func (f *fakeAIK8s) SuggestCommand(ctx context.Context, intent string) (string, error) {
	return "", nil
}

type k8sAuditRegistry struct {
	k8s   map[string]outbound.K8sProvider
	ai    map[string]outbound.AIProvider
	aiErr map[string]error
}

func (r *k8sAuditRegistry) GetCloudAuthProvider(string) (outbound.AuthProvider, error) {
	return nil, errors.New("unused")
}
func (r *k8sAuditRegistry) GetCloudInventoryProvider(string) (outbound.InventoryProvider, error) {
	return nil, errors.New("unused")
}
func (r *k8sAuditRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	if p, ok := r.k8s[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown k8s provider: " + name)
}
func (r *k8sAuditRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	if err, ok := r.aiErr[name]; ok {
		return nil, err
	}
	if p, ok := r.ai[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown AI provider: " + name)
}

func TestK8sAudit_HappyPath_FlagsCreatingState(t *testing.T) {
	reg := &k8sAuditRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{clusters: []*k8s.Cluster{
				{Name: "prod", Region: "us-east-1", Status: "ACTIVE", Version: "1.30", NodeCount: 5},
				{Name: "dev", Region: "us-east-1", Status: "CREATING", Version: "1.29", NodeCount: 2},
			}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIK8s{text: "Wait for CREATING cluster to finish."}},
	}
	skill := NewK8sAuditSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	findings := m["findings"].([]k8sFinding)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (CREATING dev cluster), got %d: %+v", len(findings), findings)
	}
	if findings[0].Severity != "high" || findings[0].Category != "lifecycle" {
		t.Errorf("unexpected finding: %+v", findings[0])
	}
	if m["remediation"] == "" {
		t.Error("expected remediation text from AI")
	}
}

func TestK8sAudit_NoFindings_SkipsAI(t *testing.T) {
	reg := &k8sAuditRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{clusters: []*k8s.Cluster{
				{Name: "prod", Region: "us-east-1", Status: "ACTIVE", Version: "1.30", NodeCount: 5},
			}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIK8s{text: "should not be called"}},
	}
	skill := NewK8sAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	m := out.(map[string]any)
	if m["remediation"] != "" {
		t.Errorf("expected empty remediation when no findings, got %q", m["remediation"])
	}
}

func TestK8sAudit_AIFailure_StillReturnsFindings(t *testing.T) {
	reg := &k8sAuditRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{clusters: []*k8s.Cluster{{Name: "broken", Status: "DELETING", Version: "1.30", Region: "us-east-1"}}},
		},
		aiErr: map[string]error{"gemini": errors.New("rate-limited")},
	}
	skill := NewK8sAuditSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	rem := m["remediation"].(string)
	if rem == "" {
		t.Error("expected non-empty remediation explaining AI failure")
	}
	findings := m["findings"].([]k8sFinding)
	if len(findings) == 0 {
		t.Error("findings must still be returned even when AI fails")
	}
}

func TestK8sAudit_ListClustersError_Captured(t *testing.T) {
	reg := &k8sAuditRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{err: errors.New("access denied")},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIK8s{text: "check IAM"}},
	}
	skill := NewK8sAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	m := out.(map[string]any)
	findings := m["findings"].([]k8sFinding)
	if len(findings) != 1 || findings[0].Category != "list_clusters" {
		t.Fatalf("expected list_clusters finding, got %+v", findings)
	}
}

func TestK8sAudit_ZeroNodesIsInfo(t *testing.T) {
	reg := &k8sAuditRegistry{
		k8s: map[string]outbound.K8sProvider{
			"gcp": &fakeK8s{clusters: []*k8s.Cluster{{Name: "autopilot", Status: "RUNNING", Version: "1.30", NodeCount: 0, Region: "us-central1"}}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIK8s{text: "ok"}},
	}
	skill := NewK8sAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"gcp"}})
	findings := out.(map[string]any)["findings"].([]k8sFinding)
	if len(findings) != 1 || findings[0].Severity != "info" || findings[0].Category != "zero_nodes" {
		t.Fatalf("expected single info zero_nodes finding, got %+v", findings)
	}
}
