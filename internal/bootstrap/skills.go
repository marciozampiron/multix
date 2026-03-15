// File: internal/bootstrap/skills.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Registers platform-level capabilities into the central skill executor.

package bootstrap

import (
	"multix/internal/application/ai"
	"multix/internal/application/auth"
	"multix/internal/application/doctor"
	"multix/internal/application/inventory"
	"multix/internal/application/k8s"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// BuildSkillRegistry gathers all the skills into a central skill registry.
// Skills are injected with the ProviderRegistry so they can resolve providers dynamically.
func BuildSkillRegistry(providers outbound.ProviderRegistry) *skills.Registry {
	skillRegistry := skills.NewRegistry()

	skillRegistry.Register(doctor.NewCheckEnvSkill())

	skillRegistry.Register(auth.NewLoginSkill(providers))
	skillRegistry.Register(auth.NewValidateSkill(providers))
	skillRegistry.Register(auth.NewWhoamiSkill(providers))

	skillRegistry.Register(inventory.NewScanSkill(providers))
	skillRegistry.Register(inventory.NewSummarySkill(providers))

	skillRegistry.Register(k8s.NewListClustersSkill(providers))

	skillRegistry.Register(ai.NewExplainSkill(providers))

	return skillRegistry
}
