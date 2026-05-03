// File: internal/application/security/iam_audit_skill_test.go
// Purpose: Unit tests for security.iam_audit — finding classification + AI remediation glue.

package security

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/ai"
	"multix/internal/domain/auth"
	"multix/internal/ports/outbound"
)

type fakeAuthIAM struct {
	identity *auth.Identity
	err      error
}

func (f *fakeAuthIAM) ID() string                                                        { return "" }
func (f *fakeAuthIAM) Login(ctx context.Context, c auth.Credentials) (*auth.Session, error) {
	return nil, nil
}
func (f *fakeAuthIAM) Whoami(ctx context.Context) (*auth.Identity, error) {
	return f.identity, f.err
}
func (f *fakeAuthIAM) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	return nil, nil
}

type iamAuditRegistry struct {
	auth  map[string]outbound.AuthProvider
	ai    map[string]outbound.AIProvider
	aiErr map[string]error
}

func (r *iamAuditRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	if p, ok := r.auth[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown")
}
func (r *iamAuditRegistry) GetCloudInventoryProvider(string) (outbound.InventoryProvider, error) {
	return nil, errors.New("unused")
}
func (r *iamAuditRegistry) GetKubernetesProvider(string) (outbound.K8sProvider, error) {
	return nil, errors.New("unused")
}
func (r *iamAuditRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	if err, ok := r.aiErr[name]; ok {
		return nil, err
	}
	if p, ok := r.ai[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown AI provider: " + name)
}

type fakeAIIAM struct {
	text string
	err  error
}

func (f *fakeAIIAM) ID() string { return "gemini" }
func (f *fakeAIIAM) Generate(ctx context.Context, p ai.Prompt) (*ai.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &ai.Response{Text: f.text, ProviderName: "gemini"}, nil
}
func (f *fakeAIIAM) SuggestCommand(ctx context.Context, intent string) (string, error) {
	return "", nil
}

func TestIAMAudit_RootDetected_HasRemediation(t *testing.T) {
	reg := &iamAuditRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuthIAM{identity: &auth.Identity{Principal: "arn:aws:iam::123:root", PrincipalType: "user"}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIIAM{text: "Disable root, enable IAM Identity Center."}},
	}
	skill := NewIAMAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	m := out.(map[string]any)
	findings := m["findings"].([]iamFinding)
	if len(findings) != 1 || findings[0].Severity != "critical" || findings[0].Category != "root_usage" {
		t.Fatalf("expected critical root_usage finding, got %+v", findings)
	}
	if m["remediation"] == "" {
		t.Error("expected remediation text")
	}
}

func TestIAMAudit_LongLivedAWSUser(t *testing.T) {
	reg := &iamAuditRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuthIAM{identity: &auth.Identity{Principal: "arn:aws:iam::123:user/marcio", PrincipalType: "user"}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIIAM{text: "use SSO"}},
	}
	skill := NewIAMAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	findings := out.(map[string]any)["findings"].([]iamFinding)
	if findings[0].Severity != "medium" || findings[0].Category != "long_lived_user" {
		t.Fatalf("expected medium long_lived_user, got %+v", findings[0])
	}
}

func TestIAMAudit_ServiceAccountIsInfo(t *testing.T) {
	reg := &iamAuditRegistry{
		auth: map[string]outbound.AuthProvider{
			"gcp": &fakeAuthIAM{identity: &auth.Identity{Principal: "bot@demo.iam.gserviceaccount.com", PrincipalType: "service_account"}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIIAM{text: "rotate keys"}},
	}
	skill := NewIAMAuditSkill(reg)
	findings := mustExecute(t, skill, []string{"gcp"})
	if findings[0].Severity != "info" || findings[0].Category != "service_account" {
		t.Fatalf("expected info service_account, got %+v", findings[0])
	}
}

func TestIAMAudit_AIFailure_RemediationCarriesError(t *testing.T) {
	reg := &iamAuditRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuthIAM{identity: &auth.Identity{Principal: "arn:aws:iam::123:root", PrincipalType: "user"}},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIIAM{err: errors.New("rate-limited")}},
	}
	skill := NewIAMAuditSkill(reg)
	out, _ := skill.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	rem := out.(map[string]any)["remediation"].(string)
	if rem == "" || !contains(rem, "rate-limited") {
		t.Errorf("expected remediation to mention AI failure, got %q", rem)
	}
}

func TestIAMAudit_WhoamiError_StillReturnsFinding(t *testing.T) {
	reg := &iamAuditRegistry{
		auth: map[string]outbound.AuthProvider{
			"aws": &fakeAuthIAM{err: errors.New("expired token")},
		},
		ai: map[string]outbound.AIProvider{"gemini": &fakeAIIAM{text: "renew creds"}},
	}
	skill := NewIAMAuditSkill(reg)
	findings := mustExecute(t, skill, []string{"aws"})
	if findings[0].Category != "whoami" || findings[0].Severity != "high" {
		t.Fatalf("expected high whoami finding, got %+v", findings[0])
	}
}

func mustExecute(t *testing.T, skill interface {
	Execute(ctx context.Context, input map[string]any) (any, error)
}, providers []string) []iamFinding {
	t.Helper()
	var raw []any
	for _, p := range providers {
		raw = append(raw, p)
	}
	out, err := skill.Execute(context.Background(), map[string]any{"providers": raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return out.(map[string]any)["findings"].([]iamFinding)
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
