// File: internal/ports/outbound/provider_registry.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Defines the outbound abstract factory interface for managing integrations.

package outbound

// ProviderRegistry resolves cloud and AI providers by stable logical name.
type ProviderRegistry interface {
	GetCloudAuthProvider(name string) (AuthProvider, error)
	GetCloudInventoryProvider(name string) (InventoryProvider, error)
	GetKubernetesProvider(name string) (K8sProvider, error)
	GetAIProvider(name string) (AIProvider, error)
}
