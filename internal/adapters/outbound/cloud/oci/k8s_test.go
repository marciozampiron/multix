// File: internal/adapters/outbound/cloud/oci/k8s_test.go
// Purpose: Tests OCI OKE cluster listing without live cloud dependencies.

package oci

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func TestOCIAdapter_ListClusters(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.cfgProviderFunc = func() common.ConfigurationProvider {
		return &mockConfigProvider{tenancyID: "ocid1.tenancy.oc1..t"}
	}
	a.okeClusterListFunc = func(ctx context.Context, cfg common.ConfigurationProvider, c string) ([]*k8s.Cluster, error) {
		return []*k8s.Cluster{
			{ID: "ocid1.cluster.oc1..a", Name: "prod-oke", Region: "us-ashburn-1", Status: "ACTIVE", Version: "v1.30.1"},
			{ID: "ocid1.cluster.oc1..b", Name: "dev-oke", Region: "us-ashburn-1", Status: "CREATING"},
		}, nil
	}

	clusters, err := a.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 2 || clusters[0].Name != "prod-oke" || clusters[1].Status != "CREATING" {
		t.Fatalf("unexpected clusters: %+v", clusters)
	}
}

func TestOCIAdapter_ListClusters_Error(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.cfgProviderFunc = func() common.ConfigurationProvider {
		return &mockConfigProvider{tenancyID: "ocid1.tenancy.oc1..t"}
	}
	a.okeClusterListFunc = func(ctx context.Context, cfg common.ConfigurationProvider, c string) ([]*k8s.Cluster, error) {
		return nil, errors.New("not authorized")
	}
	if _, err := a.ListClusters(context.Background()); err == nil {
		t.Fatal("expected error from OKE listing failure")
	}
}

func TestOCIAdapter_ListClusters_TenancyMissing(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.cfgProviderFunc = func() common.ConfigurationProvider {
		return &mockConfigProvider{err: errors.New("missing tenancy")}
	}
	if _, err := a.ListClusters(context.Background()); err == nil {
		t.Fatal("expected error when tenancy OCID is missing")
	}
}
