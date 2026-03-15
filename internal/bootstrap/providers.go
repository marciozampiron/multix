// File: internal/bootstrap/providers.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Stores and resolves provider adapters by capability at runtime.

package bootstrap

import (
	"fmt"
	"strings"

	"multix/internal/ports/outbound"
)

// BootstrapRegistry stores provider adapters grouped by capability for runtime resolution.
type BootstrapRegistry struct {
	authProviders      map[string]outbound.AuthProvider
	inventoryProviders map[string]outbound.InventoryProvider
	k8sProviders       map[string]outbound.K8sProvider
	aiProviders        map[string]outbound.AIProvider
}

// NewBootstrapRegistry creates an empty provider registry grouped by capability.
func NewBootstrapRegistry() *BootstrapRegistry {
	return &BootstrapRegistry{
		authProviders:      make(map[string]outbound.AuthProvider),
		inventoryProviders: make(map[string]outbound.InventoryProvider),
		k8sProviders:       make(map[string]outbound.K8sProvider),
		aiProviders:        make(map[string]outbound.AIProvider),
	}
}

// RegisterAuth registers an authentication provider implementation.
func (r *BootstrapRegistry) RegisterAuth(name string, p outbound.AuthProvider) {
	key := normalizeProviderName(name)
	if key == "" {
		return
	}
	r.authProviders[key] = p
}

// RegisterInventory registers an inventory provider implementation.
func (r *BootstrapRegistry) RegisterInventory(name string, p outbound.InventoryProvider) {
	key := normalizeProviderName(name)
	if key == "" {
		return
	}
	r.inventoryProviders[key] = p
}

// RegisterK8s registers a Kubernetes provider implementation.
func (r *BootstrapRegistry) RegisterK8s(name string, p outbound.K8sProvider) {
	key := normalizeProviderName(name)
	if key == "" {
		return
	}
	r.k8sProviders[key] = p
}

// RegisterAI registers an AI provider implementation.
func (r *BootstrapRegistry) RegisterAI(name string, p outbound.AIProvider) {
	key := normalizeProviderName(name)
	if key == "" {
		return
	}
	r.aiProviders[key] = p
}

// GetCloudAuthProvider resolves an authentication provider by name.
func (r *BootstrapRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	key := normalizeProviderName(name)
	if key == "" {
		return nil, fmt.Errorf("provider name is required")
	}

	p, ok := r.authProviders[key]
	if !ok {
		return nil, fmt.Errorf("auth provider %q not found", key)
	}
	return p, nil
}

// GetCloudInventoryProvider resolves an inventory provider by name.
func (r *BootstrapRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	key := normalizeProviderName(name)
	if key == "" {
		return nil, fmt.Errorf("provider name is required")
	}

	p, ok := r.inventoryProviders[key]
	if !ok {
		return nil, fmt.Errorf("inventory provider %q not found", key)
	}
	return p, nil
}

// GetKubernetesProvider resolves a Kubernetes provider by name.
func (r *BootstrapRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	key := normalizeProviderName(name)
	if key == "" {
		return nil, fmt.Errorf("provider name is required")
	}

	p, ok := r.k8sProviders[key]
	if !ok {
		return nil, fmt.Errorf("k8s provider %q not found", key)
	}
	return p, nil
}

// GetAIProvider resolves an AI provider by name.
func (r *BootstrapRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	key := normalizeProviderName(name)
	if key == "" {
		return nil, fmt.Errorf("provider name is required")
	}

	p, ok := r.aiProviders[key]
	if !ok {
		return nil, fmt.Errorf("ai provider %q not found", key)
	}
	return p, nil
}

// normalizeProviderName canonicalizes provider identifiers for stable lookups.
func normalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
