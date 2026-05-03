// File: internal/application/doctor/k8s_skill_test.go
package doctor

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/k8s"
	"multix/internal/ports/outbound"
)

type fakeK8s struct {
	id       string
	clusters []*k8s.Cluster
	err      error
}

func (f *fakeK8s) ID() string { return f.id }
func (f *fakeK8s) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	return f.clusters, f.err
}
func (f *fakeK8s) SyncContext(ctx context.Context, name, region string) error { return nil }

func TestK8sHealthSkill_HappyPath(t *testing.T) {
	reg := &fakeRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{id: "aws", clusters: []*k8s.Cluster{
				{Name: "prod", Status: "ACTIVE", Version: "1.30"},
				{Name: "dev", Status: "CREATING", Version: "1.30"},
			}},
			"gcp": &fakeK8s{id: "gcp", clusters: []*k8s.Cluster{{Name: "g1", Status: "RUNNING"}}},
			"oci": &fakeK8s{id: "oci", clusters: nil},
		},
	}
	skill := NewK8sHealthSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	if m["total_clusters"].(int) != 3 || m["total_healthy"].(int) != 2 {
		t.Fatalf("unexpected aggregates: %+v", m)
	}
}

func TestK8sHealthSkill_ProviderError(t *testing.T) {
	reg := &fakeRegistry{
		k8s: map[string]outbound.K8sProvider{
			"aws": &fakeK8s{err: errors.New("api unreachable")},
		},
	}
	skill := NewK8sHealthSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	report := out.(map[string]any)["providers"].([]map[string]any)
	if report[0]["ok"] != false || report[0]["error"] == "" {
		t.Fatalf("expected error-flagged provider entry, got %+v", report[0])
	}
}

func TestIsHealthyClusterStatus(t *testing.T) {
	cases := map[string]bool{
		"ACTIVE":   true,
		"RUNNING":  true,
		"running":  true,
		" Active":  true,
		"CREATING": false,
		"DELETING": false,
		"":         false,
	}
	for input, want := range cases {
		if got := isHealthyClusterStatus(input); got != want {
			t.Errorf("isHealthyClusterStatus(%q) = %v; want %v", input, got, want)
		}
	}
}
