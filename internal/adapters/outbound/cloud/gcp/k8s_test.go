// File: internal/adapters/outbound/cloud/gcp/k8s_test.go
// Purpose: Tests GCP GKE cluster listing without live cloud dependencies.

package gcp

import (
	"context"
	"errors"
	"testing"

	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newAdapterWithGKEFake(clusters []*k8s.Cluster, projectID string, listErr error) *Adapter {
	a := NewAdapter(logger.New("info"))
	a.findCredentialsFunc = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return &google.Credentials{
			ProjectID:   projectID,
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "x"}),
		}, nil
	}
	a.execCmdFunc = mockExecFail
	a.gkeClusterListFunc = func(ctx context.Context, p string) ([]*k8s.Cluster, error) {
		if listErr != nil {
			return nil, listErr
		}
		return clusters, nil
	}
	return a
}

func TestGCPAdapter_ListClusters(t *testing.T) {
	a := newAdapterWithGKEFake([]*k8s.Cluster{
		{ID: "111", Name: "gke-prod", Region: "us-central1", Status: "RUNNING", Version: "1.30.1", NodeCount: 5},
		{ID: "222", Name: "gke-autopilot", Region: "us-east1", Status: "RUNNING", Version: "1.30.1"},
	}, "demo-project", nil)

	clusters, err := a.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 2 || clusters[0].Name != "gke-prod" || clusters[0].NodeCount != 5 {
		t.Fatalf("unexpected clusters: %+v", clusters)
	}
}

func TestGCPAdapter_ListClusters_Error(t *testing.T) {
	a := newAdapterWithGKEFake(nil, "demo-project", errors.New("access denied"))
	if _, err := a.ListClusters(context.Background()); err == nil {
		t.Fatal("expected error from GKE listing failure")
	}
}
