// File: internal/application/cost/quick_scan_skill_test.go
package cost

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/inventory"
	"multix/internal/ports/outbound"
)

type fakeInv struct {
	summary *inventory.Summary
	err     error
}

func (f *fakeInv) ID() string { return "" }
func (f *fakeInv) List(ctx context.Context, t string) ([]*inventory.Resource, error) {
	return nil, nil
}
func (f *fakeInv) Scan(ctx context.Context) (*inventory.Summary, error) {
	return f.summary, f.err
}

type fakeRegistry struct {
	inv map[string]outbound.InventoryProvider
}

func (f *fakeRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	return nil, errors.New("not used")
}
func (f *fakeRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	if p, ok := f.inv[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown provider")
}
func (f *fakeRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	return nil, errors.New("not used")
}
func (f *fakeRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	return nil, errors.New("not used")
}

func TestQuickScanSkill_Aggregates(t *testing.T) {
	reg := &fakeRegistry{
		inv: map[string]outbound.InventoryProvider{
			"aws": &fakeInv{summary: &inventory.Summary{ProviderName: "aws", Total: 10, CountByType: map[string]int{"EC2": 7, "S3": 3}}},
			"gcp": &fakeInv{summary: &inventory.Summary{ProviderName: "gcp", Total: 5, CountByType: map[string]int{"computeEngine": 2, "cloudStorage": 3}}},
		},
	}
	skill := NewQuickScanSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws", "gcp"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	if m["grand_total"].(int) != 15 {
		t.Fatalf("expected grand_total=15, got %+v", m["grand_total"])
	}
	if m["note"] == "" {
		t.Fatal("expected disclaimer note about resource-count proxy")
	}
}

func TestQuickScanSkill_ScanError(t *testing.T) {
	reg := &fakeRegistry{
		inv: map[string]outbound.InventoryProvider{
			"aws": &fakeInv{err: errors.New("permission denied")},
		},
	}
	skill := NewQuickScanSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	report := out.(map[string]any)["providers"].([]map[string]any)
	if report[0]["ok"] != false {
		t.Fatalf("expected ok=false on scan error, got %+v", report[0])
	}
}
