package bootstrap

import (
	"fmt"
	"multix/internal/ports/outbound"
)

// BootstrapRegistry acts as an Abstract Factory for Multi-cloud adapters.
type BootstrapRegistry struct {
	authProviders      map[string]outbound.AuthProvider
	inventoryProviders map[string]outbound.InventoryProvider
	k8sProviders       map[string]outbound.K8sProvider
	aiProviders        map[string]outbound.AIProvider
}

func NewBootstrapRegistry() *BootstrapRegistry {
	return &BootstrapRegistry{
		authProviders:      make(map[string]outbound.AuthProvider),
		inventoryProviders: make(map[string]outbound.InventoryProvider),
		k8sProviders:       make(map[string]outbound.K8sProvider),
		aiProviders:        make(map[string]outbound.AIProvider),
	}
}

func (r *BootstrapRegistry) RegisterAuth(name string, p outbound.AuthProvider) {
	r.authProviders[name] = p
}
func (r *BootstrapRegistry) RegisterInventory(name string, p outbound.InventoryProvider) {
	r.inventoryProviders[name] = p
}
func (r *BootstrapRegistry) RegisterK8s(name string, p outbound.K8sProvider) {
	r.k8sProviders[name] = p
}
func (r *BootstrapRegistry) RegisterAI(name string, p outbound.AIProvider) {
	r.aiProviders[name] = p
}

// Getters dynamically retrieve the provider by string identifier
func (r *BootstrapRegistry) GetAuth(name string) (outbound.AuthProvider, error) {
	p, ok := r.authProviders[name]
	if !ok {
		return nil, fmt.Errorf("auth provider '%s' not found", name)
	}
	return p, nil
}

func (r *BootstrapRegistry) GetInventory(name string) (outbound.InventoryProvider, error) {
	p, ok := r.inventoryProviders[name]
	if !ok {
		return nil, fmt.Errorf("inventory provider '%s' not found", name)
	}
	return p, nil
}

func (r *BootstrapRegistry) GetK8s(name string) (outbound.K8sProvider, error) {
	p, ok := r.k8sProviders[name]
	if !ok {
		return nil, fmt.Errorf("k8s provider '%s' not found", name)
	}
	return p, nil
}

func (r *BootstrapRegistry) GetAI(name string) (outbound.AIProvider, error) {
	p, ok := r.aiProviders[name]
	if !ok {
		return nil, fmt.Errorf("ai provider '%s' not found", name)
	}
	return p, nil
}
