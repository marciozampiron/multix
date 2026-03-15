// File: internal/bootstrap/skills.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Registers application skills into the central runtime skill registry.

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

// BuildSkillRegistry registers all available skills and injects the provider registry
// so provider-dependent skills can resolve implementations dynamically at runtime.
func BuildSkillRegistry(providers outbound.ProviderRegistry) *skills.Registry {
	skillRegistry := skills.NewRegistry()

	// Platform diagnostics.
	skillRegistry.Register(doctor.NewCheckEnvSkill())

	// Authentication and identity.
	skillRegistry.Register(auth.NewLoginSkill(providers))
	skillRegistry.Register(auth.NewValidateSkill(providers))
	skillRegistry.Register(auth.NewWhoamiSkill(providers))

	// Cloud inventory.
	skillRegistry.Register(inventory.NewScanSkill(providers))
	skillRegistry.Register(inventory.NewSummarySkill(providers))

	// Kubernetes.
	skillRegistry.Register(k8s.NewListClustersSkill(providers))

	// AI.
	skillRegistry.Register(ai.NewExplainSkill(providers))

	return skillRegistry
}
