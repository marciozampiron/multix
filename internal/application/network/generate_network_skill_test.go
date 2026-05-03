// File: internal/application/network/generate_network_skill_test.go
// Purpose: Unit tests for infra.generate_network — happy path, parser tolerance,
// validation guardrails, AI failure propagation.

package network

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/ai"
	netdom "multix/internal/domain/network"
	"multix/internal/ports/outbound"
)

type fakeAI struct {
	id   string
	resp *ai.Response
	err  error
}

func (f *fakeAI) ID() string { return f.id }
func (f *fakeAI) Generate(ctx context.Context, p ai.Prompt) (*ai.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}
func (f *fakeAI) SuggestCommand(ctx context.Context, intent string) (string, error) {
	return "", nil
}

type stubRegistry struct {
	ai    map[string]outbound.AIProvider
	aiErr map[string]error
}

func (s *stubRegistry) GetCloudAuthProvider(string) (outbound.AuthProvider, error) {
	return nil, errors.New("unused")
}
func (s *stubRegistry) GetCloudInventoryProvider(string) (outbound.InventoryProvider, error) {
	return nil, errors.New("unused")
}
func (s *stubRegistry) GetKubernetesProvider(string) (outbound.K8sProvider, error) {
	return nil, errors.New("unused")
}
func (s *stubRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	if err, ok := s.aiErr[name]; ok {
		return nil, err
	}
	if p, ok := s.ai[name]; ok {
		return p, nil
	}
	return nil, errors.New("unknown AI provider: " + name)
}

const validJSON = `{
  "vpc_cidr": "10.0.0.0/16",
  "subnets": [
    {"name": "public-a", "cidr": "10.0.1.0/24", "availability_zone": "us-east-1a", "tier": "public"},
    {"name": "private-a", "cidr": "10.0.10.0/24", "availability_zone": "us-east-1a", "tier": "private"}
  ],
  "route_rules": [
    {"from": "private-a", "to": "nat", "action": "nat"}
  ],
  "rationale": "Two-tier baseline with NAT egress"
}`

func newSkill(text string, err error) *GenerateNetworkSkill {
	reg := &stubRegistry{
		ai: map[string]outbound.AIProvider{
			"gemini": &fakeAI{id: "gemini", resp: &ai.Response{Text: text, ProviderName: "gemini"}, err: err},
		},
	}
	return NewGenerateNetworkSkill(reg).(*GenerateNetworkSkill)
}

func TestGenerateNetwork_HappyPath(t *testing.T) {
	skill := newSkill(validJSON, nil)
	out, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws",
		"intent":   "two-tier baseline",
		"region":   "us-east-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	if m["ai_provider"] != "gemini" {
		t.Errorf("ai_provider mismatch: %v", m["ai_provider"])
	}
	if m["rationale"] != "Two-tier baseline with NAT egress" {
		t.Errorf("rationale mismatch: %v", m["rationale"])
	}
	spec := m["spec"].(netdom.NetworkSpec)
	if spec.ProviderName != "aws" || spec.Region != "us-east-1" || len(spec.Subnets) != 2 {
		t.Fatalf("spec wrong: %+v", spec)
	}
}

func TestGenerateNetwork_ToleratesMarkdownFence(t *testing.T) {
	wrapped := "```json\n" + validJSON + "\n```"
	skill := newSkill(wrapped, nil)
	if _, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws", "intent": "x", "region": "us-east-1",
	}); err != nil {
		t.Fatalf("unexpected parse error on fenced output: %v", err)
	}
}

func TestGenerateNetwork_RejectsNonJSON(t *testing.T) {
	skill := newSkill("Sure! Here's a network for you...", nil)
	_, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws", "intent": "x", "region": "us-east-1",
	})
	if err == nil {
		t.Fatal("expected error on non-JSON output")
	}
	// raw output must be in the error message so the agent can iterate
	if !contains(err.Error(), "Sure! Here's a network") {
		t.Errorf("error must include raw output, got: %v", err)
	}
}

func TestGenerateNetwork_AIFailurePropagates(t *testing.T) {
	skill := newSkill("", errors.New("rate-limited"))
	_, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws", "intent": "x", "region": "us-east-1",
	})
	if err == nil {
		t.Fatal("expected AI error to propagate")
	}
}

func TestGenerateNetwork_RejectsInvalidTier(t *testing.T) {
	bad := `{"vpc_cidr":"10.0.0.0/16","subnets":[{"name":"x","cidr":"10.0.1.0/24","availability_zone":"a","tier":"unknown"}]}`
	skill := newSkill(bad, nil)
	_, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws", "intent": "x", "region": "us-east-1",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid tier")
	}
}

func TestGenerateNetwork_RejectsDuplicateSubnetName(t *testing.T) {
	bad := `{"vpc_cidr":"10.0.0.0/16","subnets":[
		{"name":"a","cidr":"10.0.1.0/24","availability_zone":"a","tier":"public"},
		{"name":"a","cidr":"10.0.2.0/24","availability_zone":"b","tier":"private"}
	]}`
	skill := newSkill(bad, nil)
	_, err := skill.Execute(context.Background(), map[string]any{
		"provider": "aws", "intent": "x", "region": "us-east-1",
	})
	if err == nil || !contains(err.Error(), "duplicate subnet name") {
		t.Fatalf("expected duplicate-subnet error, got %v", err)
	}
}

func TestGenerateNetwork_RequiresMandatoryInput(t *testing.T) {
	skill := newSkill(validJSON, nil)
	_, err := skill.Execute(context.Background(), map[string]any{"provider": "aws"})
	if err == nil {
		t.Fatal("expected error when intent/region missing")
	}
}

func TestGenerateNetwork_StampsAuthoritativeFields(t *testing.T) {
	// LLM lies about provider/region — skill must overwrite with input values.
	lying := `{"vpc_cidr":"10.0.0.0/16","subnets":[{"name":"a","cidr":"10.0.1.0/24","availability_zone":"a","tier":"public"}]}`
	skill := newSkill(lying, nil)
	out, err := skill.Execute(context.Background(), map[string]any{
		"provider": "oci", "intent": "x", "region": "sa-saopaulo-1",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	spec := out.(map[string]any)["spec"].(netdom.NetworkSpec)
	if spec.ProviderName != "oci" || spec.Region != "sa-saopaulo-1" {
		t.Fatalf("authoritative fields not stamped: %+v", spec)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
