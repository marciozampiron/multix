// File: test/skill_contract_test.go
// Purpose: v1.0 contract test — every registered Skill MUST satisfy the
// universal Skill contract (well-formed name, non-empty description, valid
// JSON-Schema-shaped input). Acts as a tripwire for #19 stability promise.

package test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"multix/internal/bootstrap"
)

var skillNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

func TestSkillContract_AllRegisteredSkills(t *testing.T) {
	providers := bootstrap.NewBootstrapRegistry()
	registry := bootstrap.BuildSkillRegistry(providers)
	skills := registry.ListAll()

	if len(skills) == 0 {
		t.Fatal("registry produced zero skills — bootstrap is broken")
	}

	seen := map[string]bool{}
	for _, s := range skills {
		t.Run(s.Name(), func(t *testing.T) {
			name := s.Name()
			if name == "" {
				t.Fatal("Name() returned empty string")
			}
			if !skillNameRE.MatchString(name) {
				t.Errorf("Name %q violates <domain>.<verb> convention (regex %s)", name, skillNameRE)
			}
			if seen[name] {
				t.Errorf("duplicate skill name %q in registry", name)
			}
			seen[name] = true

			if desc := strings.TrimSpace(s.Description()); desc == "" {
				t.Error("Description() returned empty string — required for LLM tool selection")
			}

			schema := s.InputSchema()
			if schema == nil {
				t.Fatal("InputSchema() returned nil — required for LLM tool calling")
			}
			// Must marshal to JSON cleanly (the agent feeds this straight into
			// the LLM API as a JSON-schema object).
			raw, err := json.Marshal(schema)
			if err != nil {
				t.Fatalf("InputSchema() not JSON-marshalable: %v", err)
			}
			// Top-level must be a JSON object with a "type" field.
			var top map[string]any
			if err := json.Unmarshal(raw, &top); err != nil {
				t.Fatalf("InputSchema() does not marshal to a JSON object: %v", err)
			}
			if top["type"] != "object" {
				t.Errorf("InputSchema() top-level type must be \"object\", got %v", top["type"])
			}
		})
	}
}
