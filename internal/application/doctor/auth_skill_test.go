// File: internal/application/doctor/auth_skill_test.go
package doctor

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/auth"
	"multix/internal/ports/outbound"
)

type fakeAuthProvider struct {
	id     string
	result *auth.ValidationResult
	err    error
}

func (f *fakeAuthProvider) ID() string { return f.id }
func (f *fakeAuthProvider) Login(ctx context.Context, c auth.Credentials) (*auth.Session, error) {
	return nil, nil
}
func (f *fakeAuthProvider) Whoami(ctx context.Context) (*auth.Identity, error) {
	return &auth.Identity{Provider: f.id}, nil
}
func (f *fakeAuthProvider) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	return f.result, f.err
}

type fakeRegistry struct {
	auth    map[string]*fakeAuthProvider
	authErr map[string]error
	k8s     map[string]outbound.K8sProvider
	k8sErr  map[string]error
}

func (f *fakeRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	if err, ok := f.authErr[name]; ok {
		return nil, err
	}
	if p, ok := f.auth[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown provider: " + name)
}
func (f *fakeRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	return nil, errors.New("not used in this test")
}
func (f *fakeRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	if err, ok := f.k8sErr[name]; ok {
		return nil, err
	}
	if p, ok := f.k8s[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown k8s provider: " + name)
}
func (f *fakeRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	return nil, errors.New("not used in this test")
}

func TestAuthSkill_Defaults(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]*fakeAuthProvider{
		"aws": {id: "aws", result: &auth.ValidationResult{Provider: "aws", Valid: true, AccountID: "123", Principal: "arn:..."}},
		"gcp": {id: "gcp", result: &auth.ValidationResult{Provider: "gcp", Valid: true, AccountID: "demo-project"}},
		"oci": {id: "oci", err: errors.New("not configured")},
	}}
	skill := NewAuthSkill(reg)
	out, err := skill.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	if m["checked"].(int) != 3 || m["healthy"].(int) != 2 {
		t.Fatalf("unexpected counts: %+v", m)
	}
	findings := m["findings"].([]map[string]any)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	// Last (oci) should be ok=false with the error
	last := findings[2]
	if last["provider"] != "oci" || last["ok"] != false || last["error"] == "" {
		t.Fatalf("unexpected oci finding: %+v", last)
	}
}

func TestAuthSkill_FilteredProviders(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]*fakeAuthProvider{
		"aws": {id: "aws", result: &auth.ValidationResult{Valid: true}},
	}}
	skill := NewAuthSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	if out.(map[string]any)["checked"].(int) != 1 {
		t.Fatalf("expected single provider check, got %+v", out)
	}
}

func TestAuthSkill_ProviderUnregistered(t *testing.T) {
	reg := &fakeRegistry{authErr: map[string]error{"aws": errors.New("not registered")}}
	skill := NewAuthSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]map[string]any)
	if findings[0]["ok"] != false || findings[0]["error"] != "not registered" {
		t.Fatalf("expected not-registered error, got %+v", findings[0])
	}
}
