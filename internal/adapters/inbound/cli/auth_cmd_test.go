// File: internal/adapters/inbound/cli/auth_cmd_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Verifies auth CLI command provider wiring and renderer compatibility.

package cli

import (
	"bytes"
	"context"
	"testing"

	"multix/cmd/root"
	appSkills "multix/internal/application/skills"
	domainAuth "multix/internal/domain/auth"
	domainSkills "multix/internal/domain/skills"
)

type captureSkill struct {
	name       string
	result     any
	lastInput  map[string]any
	execCalled bool
}

func (s *captureSkill) Name() string        { return s.name }
func (s *captureSkill) Description() string { return s.name }
func (s *captureSkill) InputSchema() any    { return map[string]any{} }
func (s *captureSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	s.execCalled = true
	s.lastInput = input
	return s.result, nil
}

func TestAuthCLIProviderFlagWiring(t *testing.T) {
	rootCmd := root.NewRootCmd()
	registry := domainSkills.NewRegistry()
	validateSkill := &captureSkill{name: "auth.validate", result: &domainAuth.ValidationResult{Provider: "gcp", Valid: true}}
	whoamiSkill := &captureSkill{name: "auth.whoami", result: &domainAuth.Identity{Provider: "gcp", ProjectID: "demo"}}
	registry.Register(validateSkill)
	registry.Register(whoamiSkill)

	executor := appSkills.NewExecutor(registry)
	NewAuthHandler(rootCmd, executor).Register()

	t.Run("auth validate forwards --provider", func(t *testing.T) {
		rootCmd.SetArgs([]string{"auth", "validate", "--provider", "gcp", "--output", "json"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("unexpected command error: %v", err)
		}
		if !validateSkill.execCalled || validateSkill.lastInput["provider"] != "gcp" {
			t.Fatalf("expected provider gcp passed to skill, got %+v", validateSkill.lastInput)
		}
	})

	t.Run("auth whoami forwards --provider", func(t *testing.T) {
		rootCmd.SetArgs([]string{"auth", "whoami", "--provider", "aws", "--output", "table"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("unexpected command error: %v", err)
		}
		if !whoamiSkill.execCalled || whoamiSkill.lastInput["provider"] != "aws" {
			t.Fatalf("expected provider aws passed to skill, got %+v", whoamiSkill.lastInput)
		}
	})
}

func TestRenderAuthOutputs_JSONAndTable(t *testing.T) {
	payload := &domainAuth.ValidationResult{Provider: "aws", Valid: true, Principal: "arn:aws:iam::123:user/demo"}

	jsonBuf := &bytes.Buffer{}
	if err := renderTo(jsonBuf, payload, "json"); err != nil {
		t.Fatalf("json render failed: %v", err)
	}
	if !bytes.Contains(jsonBuf.Bytes(), []byte(`"provider": "aws"`)) {
		t.Fatalf("unexpected json output: %s", jsonBuf.String())
	}

	tableBuf := &bytes.Buffer{}
	if err := renderTo(tableBuf, payload, "table"); err != nil {
		t.Fatalf("table render failed: %v", err)
	}
	if !bytes.Contains(tableBuf.Bytes(), []byte("provider")) {
		t.Fatalf("unexpected table output: %s", tableBuf.String())
	}
}
