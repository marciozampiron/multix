package agent_test

import (
	"multix/internal/adapters/inbound/agent"
	"multix/internal/application/doctor"
	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
	"testing"
)

func TestToolAdapterManifests(t *testing.T) {
	registry := domainSkills.NewRegistry()
	registry.Register(doctor.NewCheckEnvSkill())

	executor := appSkills.NewExecutor(registry)
	adapter := agent.NewToolAdapter(registry, executor)

	manifests := adapter.Manifests()

	if len(manifests) == 0 {
		t.Fatal("expected manifests, got empty list")
	}

	if manifests[0].Name != "doctor.run" {
		t.Fatalf("expected doctor.run manifest, got %s", manifests[0].Name)
	}
}
