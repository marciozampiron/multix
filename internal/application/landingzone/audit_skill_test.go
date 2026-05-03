// File: internal/application/landingzone/audit_skill_test.go
package landingzone

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/ports/outbound"
)

type fakeAuth struct {
	result *auth.ValidationResult
	err    error
}

func (f *fakeAuth) ID() string { return "" }
func (f *fakeAuth) Login(ctx context.Context, c auth.Credentials) (*auth.Session, error) {
	return nil, nil
}
func (f *fakeAuth) Whoami(ctx context.Context) (*auth.Identity, error) { return nil, nil }
func (f *fakeAuth) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	return f.result, f.err
}

type fakeInventory struct {
	resources []*inventory.Resource
	err       error
}

func (f *fakeInventory) ID() string { return "" }
func (f *fakeInventory) List(ctx context.Context, t string) ([]*inventory.Resource, error) {
	return f.resources, f.err
}
func (f *fakeInventory) Scan(ctx context.Context) (*inventory.Summary, error) {
	return nil, nil
}

type fakeRegistry struct {
	auth map[string]outbound.AuthProvider
	inv  map[string]outbound.InventoryProvider
}

func (f *fakeRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	if p, ok := f.auth[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown")
}
func (f *fakeRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	if p, ok := f.inv[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown")
}
func (f *fakeRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	return nil, errors.New("not used")
}
func (f *fakeRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	return nil, errors.New("not used")
}

func TestAuditSkill_HappyPath(t *testing.T) {
	reg := &fakeRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuth{result: &auth.ValidationResult{AccountID: "123", Principal: "arn:..."}},
		},
		inv: map[string]outbound.InventoryProvider{
			"aws": &fakeInventory{resources: []*inventory.Resource{
				{Region: "us-east-1"}, {Region: "us-east-1"}, {Region: "eu-west-1"},
			}},
		},
	}
	skill := NewAuditSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := out.(map[string]any)["providers"].([]map[string]any)
	entry := report[0]
	if entry["account"] != "123" || entry["resource_count"].(int) != 3 || entry["distinct_regions"].(int) != 2 || entry["multi_region"] != true {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestAuditSkill_AuthError(t *testing.T) {
	reg := &fakeRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuth{err: errors.New("expired")},
		},
	}
	skill := NewAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	entry := out.(map[string]any)["providers"].([]map[string]any)[0]
	if entry["ok"] != false {
		t.Fatalf("expected ok=false on auth error, got %+v", entry)
	}
}
