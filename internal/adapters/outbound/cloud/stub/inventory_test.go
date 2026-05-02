// File: internal/adapters/outbound/cloud/stub/inventory_test.go
// Company: Hassan
// Creator: Zamp
// Created: 02/05/2026
// Updated: 02/05/2026
// Purpose: Verifies deterministic inventory stubs for local cloud tests.

package stub

import (
	"context"
	"testing"
)

func TestInventoryProviderFixtures(t *testing.T) {
	tests := []struct {
		provider     string
		computeType  string
		storageType  string
		expectedName string
	}{
		{provider: "aws", computeType: "EC2", storageType: "S3", expectedName: "prod-web-server"},
		{provider: "gcp", computeType: "computeEngine", storageType: "cloudStorage", expectedName: "gce-prod-api"},
		{provider: "oci", computeType: "Compute", storageType: "Bucket", expectedName: "prod-web-server-oci"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			provider := MustInventoryProvider(tt.provider)

			if provider.ID() != tt.provider {
				t.Fatalf("expected provider ID %q, got %q", tt.provider, provider.ID())
			}

			compute, err := provider.List(context.Background(), "compute")
			if err != nil {
				t.Fatalf("unexpected compute list error: %v", err)
			}
			if len(compute) != 1 || compute[0].Type != tt.computeType || compute[0].Name != tt.expectedName {
				t.Fatalf("unexpected compute resources: %+v", compute)
			}

			storage, err := provider.List(context.Background(), "storage")
			if err != nil {
				t.Fatalf("unexpected storage list error: %v", err)
			}
			if len(storage) != 1 || storage[0].Type != tt.storageType {
				t.Fatalf("unexpected storage resources: %+v", storage)
			}

			summary, err := provider.Scan(context.Background())
			if err != nil {
				t.Fatalf("unexpected scan error: %v", err)
			}
			if summary.ProviderName != tt.provider || summary.Total != 2 {
				t.Fatalf("unexpected summary: %+v", summary)
			}
			if summary.CountByType[tt.computeType] != 1 || summary.CountByType[tt.storageType] != 1 {
				t.Fatalf("unexpected summary counts: %+v", summary.CountByType)
			}
		})
	}
}

func TestInventoryProviderReturnsClonedResources(t *testing.T) {
	provider := MustInventoryProvider("aws")

	resources, err := provider.List(context.Background(), "compute")
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	resources[0].Name = "mutated"
	resources[0].Tags["source"] = "mutated"

	nextResources, err := provider.List(context.Background(), "compute")
	if err != nil {
		t.Fatalf("unexpected second list error: %v", err)
	}
	if nextResources[0].Name == "mutated" || nextResources[0].Tags["source"] == "mutated" {
		t.Fatalf("expected cloned fixture resources, got %+v", nextResources[0])
	}
}

func TestNewInventoryProviderRejectsUnsupportedCloud(t *testing.T) {
	if _, err := NewInventoryProvider("azure"); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}
