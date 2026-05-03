// File: internal/adapters/inbound/runtime/inventory_e2e_test.go
// Company: Hassan
// Creator: Zamp
// Created: 03/05/2026
// Purpose: E2E coverage for inventory execution through the local runtime using cloud stubs.

package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"multix/internal/adapters/inbound/agent"
	"multix/internal/adapters/outbound/cloud/stub"
	inventoryapp "multix/internal/application/inventory"
	appSkills "multix/internal/application/skills"
	domainSkills "multix/internal/domain/skills"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"
)

func TestInventoryRuntimeE2EWithAWSAndGCPStubs(t *testing.T) {
	server := newInventoryStubRuntime(t)

	tests := []struct {
		provider string
		service  string
		wantType string
		wantName string
		want     float64
	}{
		{provider: "aws", service: "compute", wantType: "EC2", wantName: "prod-web-server", want: 2},
		{provider: "aws", service: "storage", wantType: "S3", wantName: "prod-logs", want: 2},
		{provider: "gcp", service: "compute", wantType: "computeEngine", wantName: "gce-prod-api", want: 2},
		{provider: "gcp", service: "storage", wantType: "cloudStorage", wantName: "gcs-backup-vault", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"/"+tt.service, func(t *testing.T) {
			resp := executeRuntimeSkill(t, server, map[string]any{
				"skill":    "inventory.scan",
				"provider": tt.provider,
				"params": map[string]any{
					"service": tt.service,
				},
			})

			if !resp.Ok || resp.Skill != "inventory.scan" || resp.Provider != tt.provider {
				t.Fatalf("unexpected execute envelope: %+v", resp)
			}

			result := resultMap(t, resp.Result)
			if result["count"] != tt.want {
				t.Fatalf("expected %v resources, got %+v", tt.want, result)
			}

			resources, ok := result["resources"].([]any)
			if !ok || len(resources) != int(tt.want) {
				t.Fatalf("unexpected resources payload: %+v", result["resources"])
			}

			first := resultMap(t, resources[0])
			if first["type"] != tt.wantType || first["name"] != tt.wantName {
				t.Fatalf("unexpected first resource: %+v", first)
			}
			for _, key := range []string{"id", "type", "region", "name", "status"} {
				if first[key] == "" {
					t.Fatalf("resource missing %q field: %+v", key, first)
				}
			}
		})
	}
}

func TestInventoryRuntimeSummaryE2EWithAWSAndGCPStubs(t *testing.T) {
	server := newInventoryStubRuntime(t)

	for _, provider := range []string{"aws", "gcp"} {
		t.Run(provider, func(t *testing.T) {
			resp := executeRuntimeSkill(t, server, map[string]any{
				"skill":    "inventory.summary",
				"provider": provider,
			})

			result := resultMap(t, resp.Result)
			if result["provider"] != provider || result["total_count"] != float64(4) {
				t.Fatalf("unexpected summary result: %+v", result)
			}

			counts := resultMap(t, result["count_by_type"])
			if len(counts) != 2 {
				t.Fatalf("expected two inventory type buckets, got %+v", counts)
			}
		})
	}
}

func newInventoryStubRuntime(t *testing.T) *Server {
	t.Helper()

	providers := &inventoryStubRegistry{inventory: map[string]outbound.InventoryProvider{}}
	for _, provider := range []string{"aws", "gcp"} {
		providers.inventory[provider] = stub.MustInventoryProvider(provider)
	}

	registry := domainSkills.NewRegistry()
	registry.Register(inventoryapp.NewScanSkill(providers))
	registry.Register(inventoryapp.NewSummarySkill(providers))

	executor := appSkills.NewExecutor(registry)
	adapter := agent.NewToolAdapter(registry, executor)
	return NewServer(logger.New("error"), adapter, 8080)
}

type inventoryStubRegistry struct {
	inventory map[string]outbound.InventoryProvider
}

func (r *inventoryStubRegistry) GetCloudInventoryProvider(name string) (outbound.InventoryProvider, error) {
	provider, ok := r.inventory[name]
	if !ok {
		return nil, fmt.Errorf("inventory provider not found: %s", name)
	}
	return provider, nil
}

func (r *inventoryStubRegistry) GetCloudAuthProvider(name string) (outbound.AuthProvider, error) {
	return nil, fmt.Errorf("auth provider not configured for inventory e2e: %s", name)
}

func (r *inventoryStubRegistry) GetKubernetesProvider(name string) (outbound.K8sProvider, error) {
	return nil, fmt.Errorf("kubernetes provider not configured for inventory e2e: %s", name)
}

func (r *inventoryStubRegistry) GetAIProvider(name string) (outbound.AIProvider, error) {
	return nil, fmt.Errorf("ai provider not configured for inventory e2e: %s", name)
}

func executeRuntimeSkill(t *testing.T, server *Server, payload map[string]any) executeSuccessResponse {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/execute", bytes.NewReader(body))
	req.Header.Set(RequestIDHeader, fmt.Sprintf("test-%s", t.Name()))
	rr := httptest.NewRecorder()
	server.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp executeSuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode execute response: %v", err)
	}
	return resp
}

func resultMap(t *testing.T, value any) map[string]any {
	t.Helper()

	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T: %+v", value, value)
	}
	return result
}
