// File: internal/application/network/generate_network_skill.go
// Company: Hassan
// Creator: Zamp
// Created: 03/05/2026
// Updated: 03/05/2026
// Purpose: infra.generate_network skill — AI-driven network topology generator.

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"multix/internal/domain/ai"
	netdom "multix/internal/domain/network"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

const defaultAIProvider = "gemini"

// GenerateNetworkSkill asks an AIProvider to design a NetworkSpec from a
// natural-language intent. The skill is honest about its dependency on the
// model: if the LLM returns non-JSON or violates the schema, the error
// surfaces the raw output so the agent loop can retry with a tighter prompt.
type GenerateNetworkSkill struct {
	providers outbound.ProviderRegistry
}

// NewGenerateNetworkSkill creates the skill bound to the registry.
func NewGenerateNetworkSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &GenerateNetworkSkill{providers: pr}
}

func (s *GenerateNetworkSkill) Name() string { return "infra.generate_network" }

func (s *GenerateNetworkSkill) Description() string {
	return "Designs a multi-cloud network topology (VPC + subnets + routes) from a natural-language intent. Output is a provider-agnostic NetworkSpec; rendering to Terraform/CloudFormation is a downstream skill."
}

func (s *GenerateNetworkSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider":    map[string]any{"type": "string", "description": "Target cloud (aws|gcp|oci)."},
			"intent":      map[string]any{"type": "string", "description": "Natural-language description of the topology."},
			"region":      map[string]any{"type": "string", "description": "Target region."},
			"cidr":        map[string]any{"type": "string", "description": "VPC CIDR. Defaults to 10.0.0.0/16."},
			"ai_provider": map[string]any{"type": "string", "description": "AI backend. Defaults to gemini."},
		},
		"required": []string{"provider", "intent", "region"},
	}
}

func (s *GenerateNetworkSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	provider, _ := input["provider"].(string)
	intent, _ := input["intent"].(string)
	region, _ := input["region"].(string)
	cidr, _ := input["cidr"].(string)
	if cidr == "" {
		cidr = "10.0.0.0/16"
	}
	aiProviderName, _ := input["ai_provider"].(string)
	if aiProviderName == "" {
		aiProviderName = defaultAIProvider
	}

	if provider == "" || intent == "" || region == "" {
		return nil, fmt.Errorf("provider, intent and region are required")
	}

	aiProvider, err := s.providers.GetAIProvider(aiProviderName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve AI provider %q: %w", aiProviderName, err)
	}

	prompt := ai.Prompt{
		Text:      buildPrompt(provider, intent, region, cidr),
		CreatedAt: time.Now(),
	}
	resp, err := aiProvider.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	spec, rationale, err := parseModelOutput(resp.Text)
	if err != nil {
		return nil, fmt.Errorf("could not parse AI output as NetworkSpec (raw: %q): %w", resp.Text, err)
	}

	// Stamp the input-supplied identity onto the spec — the LLM may omit
	// or paraphrase these fields, but they are authoritative.
	spec.ProviderName = provider
	spec.Region = region
	if spec.VPCCidr == "" {
		spec.VPCCidr = cidr
	}

	if err := validateSpec(spec); err != nil {
		return nil, fmt.Errorf("AI output failed validation: %w", err)
	}

	return map[string]any{
		"spec":        spec,
		"rationale":   rationale,
		"ai_provider": aiProviderName,
	}, nil
}

func buildPrompt(provider, intent, region, cidr string) string {
	return fmt.Sprintf(`Design a network topology for the following request and return ONLY a JSON object.

Provider: %s
Region: %s
VPC CIDR: %s
Intent: %s

Schema (return exactly this shape, no prose, no markdown fences):
{
  "vpc_cidr": "<CIDR>",
  "subnets": [
    {"name": "<id>", "cidr": "<sub-CIDR>", "availability_zone": "<AZ>", "tier": "public|private|database"}
  ],
  "route_rules": [
    {"from": "<subnet or igw/nat>", "to": "<subnet or igw/nat>", "action": "allow|deny|nat"}
  ],
  "rationale": "<one short paragraph explaining the design>"
}`, provider, region, cidr, intent)
}

// parseModelOutput accepts the raw text from the LLM and returns the parsed
// spec plus the rationale string. It tolerates surrounding whitespace and
// markdown code fences.
func parseModelOutput(raw string) (netdom.NetworkSpec, string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var envelope struct {
		netdom.NetworkSpec
		Rationale string `json:"rationale"`
	}
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return netdom.NetworkSpec{}, "", err
	}
	return envelope.NetworkSpec, envelope.Rationale, nil
}

func validateSpec(spec netdom.NetworkSpec) error {
	if spec.VPCCidr == "" {
		return fmt.Errorf("vpc_cidr is empty")
	}
	if len(spec.Subnets) == 0 {
		return fmt.Errorf("at least one subnet required")
	}
	seen := map[string]struct{}{}
	for _, sub := range spec.Subnets {
		if sub.Name == "" || sub.CIDR == "" {
			return fmt.Errorf("subnet missing name or CIDR: %+v", sub)
		}
		if _, dup := seen[sub.Name]; dup {
			return fmt.Errorf("duplicate subnet name %q", sub.Name)
		}
		seen[sub.Name] = struct{}{}
		switch sub.Tier {
		case "public", "private", "database":
		default:
			return fmt.Errorf("subnet %q has invalid tier %q (expected public|private|database)", sub.Name, sub.Tier)
		}
	}
	return nil
}
