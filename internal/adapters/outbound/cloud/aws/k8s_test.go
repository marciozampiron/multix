// File: internal/adapters/outbound/cloud/aws/k8s_test.go
// Purpose: Tests AWS EKS cluster listing without live cloud dependencies.

package aws

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
)

func TestAWSAdapter_ListClusters(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.eksClusterListFunc = func(ctx context.Context) ([]*k8s.Cluster, error) {
		return []*k8s.Cluster{
			{ID: "arn:aws:eks:us-east-1:123:cluster/prod", Name: "prod", Region: "us-east-1", Status: "ACTIVE", Version: "1.30"},
			{ID: "arn:aws:eks:us-east-1:123:cluster/dev", Name: "dev", Region: "us-east-1", Status: "CREATING", Version: "1.29"},
		}, nil
	}

	clusters, err := a.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 2 || clusters[0].Name != "prod" || clusters[1].Status != "CREATING" {
		t.Fatalf("unexpected clusters: %+v", clusters)
	}
}

func TestAWSAdapter_ListClusters_Error(t *testing.T) {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.eksClusterListFunc = func(ctx context.Context) ([]*k8s.Cluster, error) {
		return nil, errors.New("access denied")
	}
	if _, err := a.ListClusters(context.Background()); err == nil {
		t.Fatal("expected error from EKS listing failure")
	}
}
