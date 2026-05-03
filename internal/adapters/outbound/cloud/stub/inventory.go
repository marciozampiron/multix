// File: internal/adapters/outbound/cloud/stub/inventory.go
// Company: Hassan
// Creator: Zamp
// Created: 02/05/2026
// Updated: 02/05/2026
// Purpose: Provides deterministic cloud inventory stubs for local tests.

package stub

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"multix/internal/domain/inventory"
)

var errUnsupportedProvider = errors.New("unsupported inventory stub provider")

var fixedInventoryTime = time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)

// InventoryProvider serves deterministic cloud inventory responses for tests.
type InventoryProvider struct {
	provider  string
	resources []*inventory.Resource
}

// NewInventoryProvider creates an inventory provider backed by local cloud fixtures.
func NewInventoryProvider(provider string) (*InventoryProvider, error) {
	key := normalizeProvider(provider)
	resources, ok := inventoryFixtures()[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errUnsupportedProvider, provider)
	}

	return &InventoryProvider{
		provider:  key,
		resources: cloneResources(resources),
	}, nil
}

// MustInventoryProvider creates a fixture-backed provider and panics on invalid test setup.
func MustInventoryProvider(provider string) *InventoryProvider {
	p, err := NewInventoryProvider(provider)
	if err != nil {
		panic(err)
	}
	return p
}

// SupportedInventoryProviders returns the cloud names backed by local fixtures.
func SupportedInventoryProviders() []string {
	providers := make([]string, 0, len(inventoryFixtures()))
	for provider := range inventoryFixtures() {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	return providers
}

// ID returns the provider identifier.
func (p *InventoryProvider) ID() string {
	return p.provider
}

// List returns fixture resources filtered by service/resource type.
func (p *InventoryProvider) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	filter := normalizeProvider(resourceType)
	if filter == "" {
		return cloneResources(p.resources), nil
	}

	var resources []*inventory.Resource
	for _, resource := range p.resources {
		if matchesResourceType(resource.Type, filter) {
			resources = append(resources, cloneResource(resource))
		}
	}

	return resources, nil
}

// Scan summarizes all fixture resources for the provider.
func (p *InventoryProvider) Scan(ctx context.Context) (*inventory.Summary, error) {
	counts := make(map[string]int)
	for _, resource := range p.resources {
		counts[resource.Type]++
	}

	return &inventory.Summary{
		ProviderName: p.provider,
		Total:        len(p.resources),
		CountByType:  counts,
	}, nil
}

func inventoryFixtures() map[string][]*inventory.Resource {
	return map[string][]*inventory.Resource{
		"aws": {
			newFixtureResource("i-0abc123", "123456789012", "us-east-1", "EC2", "prod-web-server", "RUNNING"),
			newFixtureResource("i-0def456", "123456789012", "us-west-2", "EC2", "batch-worker-a", "STOPPED"),
			newFixtureResource("bucket-prod-logs", "123456789012", "us-east-1", "S3", "prod-logs", "ACTIVE"),
			newFixtureResource("bucket-audit-archive", "123456789012", "us-west-2", "S3", "audit-archive", "ACTIVE"),
		},
		"gcp": {
			newFixtureResource("gce-prod-api", "demo-project", "us-central1", "computeEngine", "gce-prod-api", "RUNNING"),
			newFixtureResource("gce-batch-worker", "demo-project", "us-east1", "computeEngine", "gce-batch-worker", "TERMINATED"),
			newFixtureResource("gcs-backup-vault", "demo-project", "us-central1", "cloudStorage", "gcs-backup-vault", "ACTIVE"),
			newFixtureResource("gcs-audit-archive", "demo-project", "us-east1", "cloudStorage", "gcs-audit-archive", "ACTIVE"),
		},
		"oci": {
			newFixtureResource("ocid1.instance.oc1..stub", "ocid1.tenancy.oc1..stub", "us-ashburn-1", "Compute", "prod-web-server-oci", "RUNNING"),
			newFixtureResource("ocid1.bucket.oc1..stub", "ocid1.tenancy.oc1..stub", "us-ashburn-1", "Bucket", "prod-object-archive", "ACTIVE"),
		},
	}
}

func newFixtureResource(id, accountID, region, resourceType, name, status string) *inventory.Resource {
	return &inventory.Resource{
		ID:        id,
		AccountID: accountID,
		Region:    region,
		Type:      resourceType,
		Name:      name,
		Status:    status,
		Tags: map[string]string{
			"source": "stub",
		},
		CreatedAt: fixedInventoryTime,
	}
}

func matchesResourceType(resourceType, filter string) bool {
	normalizedType := normalizeProvider(resourceType)
	if normalizedType == filter {
		return true
	}

	for _, alias := range resourceTypeAliases(filter) {
		if normalizedType == alias {
			return true
		}
	}
	return false
}

func resourceTypeAliases(filter string) []string {
	switch filter {
	case "compute":
		return []string{"ec2", "computeengine"}
	case "storage":
		return []string{"s3", "cloudstorage", "bucket"}
	default:
		return nil
	}
}

func normalizeProvider(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func cloneResources(resources []*inventory.Resource) []*inventory.Resource {
	clones := make([]*inventory.Resource, 0, len(resources))
	for _, resource := range resources {
		clones = append(clones, cloneResource(resource))
	}
	return clones
}

func cloneResource(resource *inventory.Resource) *inventory.Resource {
	if resource == nil {
		return nil
	}

	clone := *resource
	clone.Tags = make(map[string]string, len(resource.Tags))
	for key, value := range resource.Tags {
		clone.Tags[key] = value
	}
	return &clone
}
