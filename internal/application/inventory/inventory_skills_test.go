// File: internal/application/inventory/inventory_skills_test.go
// Company: Hassan
// Creator: Zamp
// Created: 02/05/2026
// Updated: 02/05/2026
// Purpose: Tests inventory skills against local cloud stubs.

package inventory_test

import (
	"context"
	"testing"

	"multix/internal/adapters/outbound/cloud/stub"
	inventoryapp "multix/internal/application/inventory"
	"multix/internal/bootstrap"
)

func TestScanSkillWithCloudStubs(t *testing.T) {
	providers := newStubInventoryRegistry()
	skill := inventoryapp.NewScanSkill(providers)

	tests := []struct {
		provider string
		service  string
		wantType string
		wantName string
	}{
		{provider: "aws", service: "compute", wantType: "EC2", wantName: "prod-web-server"},
		{provider: "gcp", service: "compute", wantType: "computeEngine", wantName: "gce-prod-api"},
		{provider: "oci", service: "compute", wantType: "Compute", wantName: "prod-web-server-oci"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result, err := skill.Execute(context.Background(), map[string]any{
				"provider": tt.provider,
				"service":  tt.service,
			})
			if err != nil {
				t.Fatalf("unexpected scan error: %v", err)
			}

			payload := result.(map[string]any)
			if payload["count"] != 1 {
				t.Fatalf("expected one scanned resource, got %+v", payload)
			}

			resources := payload["resources"].([]map[string]any)
			if resources[0]["type"] != tt.wantType || resources[0]["name"] != tt.wantName {
				t.Fatalf("unexpected scanned resource: %+v", resources[0])
			}
		})
	}
}

func TestSummarySkillWithCloudStubs(t *testing.T) {
	providers := newStubInventoryRegistry()
	skill := inventoryapp.NewSummarySkill(providers)

	for _, provider := range stub.SupportedInventoryProviders() {
		t.Run(provider, func(t *testing.T) {
			result, err := skill.Execute(context.Background(), map[string]any{"provider": provider})
			if err != nil {
				t.Fatalf("unexpected summary error: %v", err)
			}

			payload := result.(map[string]any)
			if payload["provider"] != provider || payload["total_count"] != 2 {
				t.Fatalf("unexpected summary payload: %+v", payload)
			}

			counts := payload["count_by_type"].(map[string]int)
			if len(counts) != 2 {
				t.Fatalf("expected two resource type counts, got %+v", counts)
			}
		})
	}
}

func newStubInventoryRegistry() *bootstrap.BootstrapRegistry {
	registry := bootstrap.NewBootstrapRegistry()
	for _, provider := range stub.SupportedInventoryProviders() {
		registry.RegisterInventory(provider, stub.MustInventoryProvider(provider))
	}
	return registry
}
