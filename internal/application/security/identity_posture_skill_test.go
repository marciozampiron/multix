// File: internal/application/security/identity_posture_skill_test.go
package security

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/auth"
	"multix/internal/ports/outbound"
)

type fakeAuth struct {
	identity *auth.Identity
	err      error
}

func (f *fakeAuth) ID() string { return "" }
func (f *fakeAuth) Login(ctx context.Context, c auth.Credentials) (*auth.Session, error) {
	return nil, nil
}
func (f *fakeAuth) Whoami(ctx context.Context) (*auth.Identity, error) {
	return f.identity, f.err
}
func (f *fakeAuth) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	return nil, nil
}

type fakeRegistry struct {
	auth    map[string]outbound.AuthProvider
	authErr map[string]error
}

func (f *fakeRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	if err, ok := f.authErr[name]; ok {
		return nil, err
	}
	if p, ok := f.auth[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown")
}
func (f *fakeRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	return nil, errors.New("not used")
}
func (f *fakeRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	return nil, errors.New("not used")
}
func (f *fakeRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	return nil, errors.New("not used")
}

func TestIdentityPostureSkill_RootDetected(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]outbound.AuthProvider{
		"aws": &fakeAuth{identity: &auth.Identity{
			Provider:      "aws",
			Principal:     "arn:aws:iam::123:root",
			PrincipalType: "user",
		}},
	}}
	skill := NewIdentityPostureSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]map[string]any)
	if findings[0]["severity"] != "critical" || findings[0]["category"] != "root_usage" {
		t.Fatalf("expected critical root_usage, got %+v", findings[0])
	}
}

func TestIdentityPostureSkill_LongLivedAWSUser(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]outbound.AuthProvider{
		"aws": &fakeAuth{identity: &auth.Identity{
			Provider:      "aws",
			Principal:     "arn:aws:iam::123:user/marcio",
			PrincipalType: "user",
		}},
	}}
	skill := NewIdentityPostureSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]map[string]any)
	if findings[0]["severity"] != "medium" || findings[0]["category"] != "long_lived_user" {
		t.Fatalf("expected medium long_lived_user, got %+v", findings[0])
	}
}

func TestIdentityPostureSkill_InfoForRole(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]outbound.AuthProvider{
		"aws": &fakeAuth{identity: &auth.Identity{Provider: "aws", Principal: "arn:aws:sts::123:assumed-role/Admin/sess", PrincipalType: "role"}},
	}}
	skill := NewIdentityPostureSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]map[string]any)
	if findings[0]["severity"] != "info" {
		t.Fatalf("expected info, got %+v", findings[0])
	}
}

func TestIdentityPostureSkill_WhoamiError(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]outbound.AuthProvider{
		"aws": &fakeAuth{err: errors.New("expired token")},
	}}
	skill := NewIdentityPostureSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]map[string]any)
	if findings[0]["severity"] != "high" || findings[0]["category"] != "whoami" {
		t.Fatalf("expected high whoami finding, got %+v", findings[0])
	}
}

func TestIdentityPostureSkill_SeverityCounts(t *testing.T) {
	reg := &fakeRegistry{auth: map[string]outbound.AuthProvider{
		"aws": &fakeAuth{identity: &auth.Identity{Principal: "arn:aws:iam::123:root", PrincipalType: "user"}},
		"gcp": &fakeAuth{identity: &auth.Identity{Provider: "gcp", PrincipalType: "service_account"}},
	}}
	skill := NewIdentityPostureSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws", "gcp"}})
	counts := out.(map[string]any)["severity_counts"].(map[string]int)
	if counts["critical"] != 1 || counts["info"] != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}
