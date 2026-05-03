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
	"multix/internal/application/cost"
	"multix/internal/application/doctor"
	"multix/internal/application/inventory"
	"multix/internal/application/k8s"
	"multix/internal/application/landingzone"
	"multix/internal/application/network"
	"multix/internal/application/security"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

// BuildSkillRegistry registers all available skills and injects the provider registry
// so provider-dependent skills can resolve implementations dynamically at runtime.
func BuildSkillRegistry(providers outbound.ProviderRegistry) *skills.Registry {
	skillRegistry := skills.NewRegistry()

	// Platform diagnostics.
	skillRegistry.Register(doctor.NewCheckEnvSkill())
	skillRegistry.Register(doctor.NewAuthSkill(providers))
	skillRegistry.Register(doctor.NewK8sHealthSkill(providers))

	// Landing zone & governance.
	skillRegistry.Register(landingzone.NewAuditSkill(providers))

	// Security posture + AI-augmented audits.
	skillRegistry.Register(security.NewIdentityPostureSkill(providers))
	skillRegistry.Register(security.NewK8sAuditSkill(providers))
	skillRegistry.Register(security.NewIAMAuditSkill(providers))

	// Cost: quick_scan is a resource-count proxy; focus_report is FOCUS-aligned billing.
	skillRegistry.Register(cost.NewQuickScanSkill(providers))
	skillRegistry.Register(cost.NewFocusReportSkill(providers))

	// Network — AI-driven topology generation.
	skillRegistry.Register(network.NewGenerateNetworkSkill(providers))

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
