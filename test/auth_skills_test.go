// File: test/auth_skills_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Validates auth.validate and auth.whoami skills logic.

package test

import (
	"context"
	"testing"

	"multix/internal/application/auth"
	"multix/internal/bootstrap"
	domainAuth "multix/internal/domain/auth"
)

// mockAuthProvider is a test double for outbound.AuthProvider.
type mockAuthProvider struct {
	id             string
	validateResult *domainAuth.ValidationResult
	validateErr    error
	whoamiIdentity *domainAuth.Identity
	whoamiErr      error
}

func (m *mockAuthProvider) ID() string { return m.id }
func (m *mockAuthProvider) Login(ctx context.Context, creds domainAuth.Credentials) (*domainAuth.Session, error) {
	return nil, nil
}
func (m *mockAuthProvider) Validate(ctx context.Context) (*domainAuth.ValidationResult, error) {
	return m.validateResult, m.validateErr
}
func (m *mockAuthProvider) Whoami(ctx context.Context) (*domainAuth.Identity, error) {
	return m.whoamiIdentity, m.whoamiErr
}

func TestAuthSkills_Execution(t *testing.T) {
	providers := bootstrap.NewBootstrapRegistry()

	mockAWS := &mockAuthProvider{
		id: "aws",
		validateResult: &domainAuth.ValidationResult{
			Provider:  "aws",
			Valid:     true,
			AccountID: "12345",
			Principal: "arn:mock",
		},
		whoamiIdentity: &domainAuth.Identity{
			Provider:      "aws",
			AccountID:     "12345",
			Principal:     "arn:mock:user/mock",
			PrincipalType: "user",
		},
	}
	providers.RegisterAuth("aws", mockAWS)

	t.Run("auth.validate execution with explicit provider", func(t *testing.T) {
		skill := auth.NewValidateSkill(providers)
		input := map[string]any{"provider": "aws"}

		res, err := skill.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, ok := res.(*domainAuth.ValidationResult)
		if !ok || result.Provider != "aws" || !result.Valid {
			t.Errorf("expected valid aws ValidationResult, got %+v", res)
		}
	})

	t.Run("auth.whoami execution with explicit provider", func(t *testing.T) {
		skill := auth.NewWhoamiSkill(providers)
		input := map[string]any{"provider": "aws"}

		res, err := skill.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		identity, ok := res.(*domainAuth.Identity)
		if !ok || identity.Provider != "aws" || identity.PrincipalType != "user" {
			t.Errorf("expected valid aws Identity, got %+v", res)
		}
	})

	t.Run("auth.validate falls back to default provider when omitted", func(t *testing.T) {
		skill := auth.NewValidateSkill(providers)

		res, err := skill.Execute(context.Background(), map[string]any{})
		if err != nil {
			t.Fatalf("unexpected error with default provider: %v", err)
		}

		result := res.(*domainAuth.ValidationResult)
		if result.Provider != "aws" {
			t.Fatalf("expected default provider aws, got %q", result.Provider)
		}
	})

	t.Run("auth.validate with unknown provider returns error", func(t *testing.T) {
		skill := auth.NewValidateSkill(providers)
		input := map[string]any{"provider": "unknown"}

		_, err := skill.Execute(context.Background(), input)
		if err == nil {
			t.Fatal("expected error for unknown provider, got nil")
		}
	})

	t.Run("auth.whoami with unknown provider returns error", func(t *testing.T) {
		skill := auth.NewWhoamiSkill(providers)
		input := map[string]any{"provider": "unknown"}

		_, err := skill.Execute(context.Background(), input)
		if err == nil {
			t.Fatal("expected error for unknown provider, got nil")
		}
	})
}
