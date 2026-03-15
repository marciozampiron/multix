// File: internal/bootstrap/registry.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Registers provider adapters by capability for runtime resolution.

package bootstrap

import (
	"multix/internal/adapters/outbound/ai/gemini"
	"multix/internal/adapters/outbound/cloud/aws"
	"multix/internal/adapters/outbound/cloud/gcp"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"
)

// BuildProviderRegistry constructs and registers the available provider adapters by capability.
func BuildProviderRegistry(log logger.Logger) outbound.ProviderRegistry {
	providers := NewBootstrapRegistry()

	// Initialize adapters.
	awsAdapter := aws.NewAdapter(log)
	gcpAdapter := gcp.NewAdapter(log)
	geminiAdapter := gemini.NewAdapter(log)

	// Register AWS capabilities.
	providers.RegisterAuth("aws", awsAdapter)
	providers.RegisterInventory("aws", awsAdapter)
	providers.RegisterK8s("aws", awsAdapter)

	// Register GCP capabilities.
	providers.RegisterAuth("gcp", gcpAdapter)
	providers.RegisterInventory("gcp", gcpAdapter)
	providers.RegisterK8s("gcp", gcpAdapter)

	// Register AI capabilities.
	providers.RegisterAI("gemini", geminiAdapter)

	return providers
}
