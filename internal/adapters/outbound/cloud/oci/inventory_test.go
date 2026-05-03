// File: internal/adapters/outbound/cloud/oci/inventory_test.go
// Purpose: Tests OCI Compute + Object Storage inventory listing without live cloud dependencies.

package oci

import (
	"context"
	"errors"
	"testing"
	"time"

	"multix/internal/platform/logger"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func newAdapterWithInventoryFakes(instances []ociInstance, buckets []ociBucket, tenancyID string) *adapter {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.cfgProviderFunc = func() common.ConfigurationProvider {
		return &mockConfigProvider{tenancyID: tenancyID, userID: "ocid1.user.oc1..u"}
	}
	a.computeListFunc = func(ctx context.Context, cfg common.ConfigurationProvider, c string) ([]ociInstance, error) {
		return instances, nil
	}
	a.bucketListFunc = func(ctx context.Context, cfg common.ConfigurationProvider, c string) ([]ociBucket, error) {
		return buckets, nil
	}
	return a
}

func TestOCIAdapter_List_Compute(t *testing.T) {
	created := time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC)
	instances := []ociInstance{
		{ID: "ocid1.instance.oc1..a", Name: "prod-web-1", Region: "us-ashburn-1", LifecycleState: "RUNNING", FreeformTags: map[string]string{"env": "prod"}, TimeCreated: created},
		{ID: "ocid1.instance.oc1..b", Name: "dev-web-1", Region: "us-ashburn-1", LifecycleState: "STOPPED"},
	}
	a := newAdapterWithInventoryFakes(instances, nil, "ocid1.tenancy.oc1..t")

	resources, err := a.List(context.Background(), "compute")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(resources))
	}
	r := resources[0]
	if r.ID != "ocid1.instance.oc1..a" || r.Name != "prod-web-1" || r.Status != "RUNNING" || r.Tags["env"] != "prod" || !r.CreatedAt.Equal(created) {
		t.Fatalf("unexpected first resource: %+v", r)
	}
	if r.Type != "Compute" || r.Region != "us-ashburn-1" || r.AccountID != "ocid1.tenancy.oc1..t" {
		t.Fatalf("unexpected first metadata: %+v", r)
	}
	if resources[1].Status != "STOPPED" {
		t.Fatalf("expected STOPPED, got %q", resources[1].Status)
	}
}

func TestOCIAdapter_List_Buckets(t *testing.T) {
	buckets := []ociBucket{
		{Name: "artifacts-prod", Namespace: "tenancy-ns", Region: "us-ashburn-1", TimeCreated: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "logs-eu", Namespace: "tenancy-ns", Region: "eu-frankfurt-1"},
	}
	a := newAdapterWithInventoryFakes(nil, buckets, "ocid1.tenancy.oc1..t")

	resources, err := a.List(context.Background(), "storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(resources))
	}
	if resources[0].Type != "Bucket" || resources[0].Name != "artifacts-prod" || resources[0].Status != "AVAILABLE" || resources[0].Tags["namespace"] != "tenancy-ns" {
		t.Fatalf("unexpected bucket resource: %+v", resources[0])
	}
}

func TestOCIAdapter_List_All(t *testing.T) {
	a := newAdapterWithInventoryFakes(
		[]ociInstance{{ID: "i", Name: "vm-1", Region: "us-ashburn-1", LifecycleState: "RUNNING"}},
		[]ociBucket{{Name: "b-1", Region: "us-ashburn-1"}, {Name: "b-2", Region: "us-ashburn-1"}},
		"ocid1.tenancy.oc1..t",
	)
	resources, err := a.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 3 {
		t.Fatalf("expected 1 Compute + 2 Bucket = 3 resources, got %d", len(resources))
	}
}

func TestOCIAdapter_List_UnknownType(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "ocid1.tenancy.oc1..t")
	resources, err := a.List(context.Background(), "block-volume")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected empty result for unknown type, got %d", len(resources))
	}
}

func TestOCIAdapter_Scan(t *testing.T) {
	a := newAdapterWithInventoryFakes(
		[]ociInstance{{ID: "i1", Name: "vm-1", Region: "us-ashburn-1"}, {ID: "i2", Name: "vm-2", Region: "us-ashburn-1"}},
		[]ociBucket{{Name: "b-1", Region: "us-ashburn-1"}},
		"ocid1.tenancy.oc1..t",
	)
	summary, err := a.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.ProviderName != "oci" || summary.Total != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.CountByType["Compute"] != 2 || summary.CountByType["Bucket"] != 1 {
		t.Fatalf("unexpected counts: %+v", summary.CountByType)
	}
}

func TestOCIAdapter_List_ComputeError(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "ocid1.tenancy.oc1..t")
	a.computeListFunc = func(ctx context.Context, cfg common.ConfigurationProvider, c string) ([]ociInstance, error) {
		return nil, errors.New("permission denied")
	}
	if _, err := a.List(context.Background(), "compute"); err == nil {
		t.Fatal("expected error from compute list failure")
	}
}

func TestOCIAdapter_List_TenancyMissing(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.cfgProviderFunc = func() common.ConfigurationProvider {
		return &mockConfigProvider{err: errors.New("missing tenancy")}
	}
	if _, err := a.List(context.Background(), "compute"); err == nil {
		t.Fatal("expected error when tenancy OCID is missing")
	}
}
