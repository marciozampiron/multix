package test

import (
	"multix/internal/bootstrap"
	"multix/internal/platform/logger"
	"testing"
)

func TestProviderRegistryResolution(t *testing.T) {
	log := logger.New("none")
	providers := bootstrap.BuildProviderRegistry(log)

	t.Run("Valid AWS Provider", func(t *testing.T) {
		p, err := providers.GetCloudAuthProvider("aws")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if p == nil {
			t.Fatal("expected a valid provider, got nil")
		}
	})

	t.Run("Valid GCP Provider", func(t *testing.T) {
		p, err := providers.GetCloudInventoryProvider("gcp")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if p == nil {
			t.Fatal("expected a valid provider, got nil")
		}
	})

	t.Run("Unknown Provider", func(t *testing.T) {
		_, err := providers.GetKubernetesProvider("unknown-cloud")
		if err == nil {
			t.Fatal("expected error for unknown provider, got nil")
		}
	})
}
