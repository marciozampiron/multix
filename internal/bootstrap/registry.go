package bootstrap

import (
	"multix/internal/adapters/outbound/ai/gemini"
	"multix/internal/adapters/outbound/cloud/aws"
	"multix/internal/adapters/outbound/cloud/gcp"
	"multix/internal/platform/logger"
)

// BuildProviderRegistry constructs the multi-cloud abstract factory.
func BuildProviderRegistry(log logger.Logger) *BootstrapRegistry {
	providers := NewBootstrapRegistry()

	// Init adapters
	awsAdapter := aws.NewAdapter(log)
	gcpAdapter := gcp.NewAdapter(log)
	geminiAdapter := gemini.NewAdapter(log)

	// Register AWS
	providers.RegisterAuth("aws", awsAdapter)
	providers.RegisterInventory("aws", awsAdapter)
	providers.RegisterK8s("aws", awsAdapter)

	// Register GCP
	providers.RegisterAuth("gcp", gcpAdapter)
	providers.RegisterInventory("gcp", gcpAdapter)
	providers.RegisterK8s("gcp", gcpAdapter)

	// Register AI
	providers.RegisterAI("gemini", geminiAdapter)

	return providers
}
