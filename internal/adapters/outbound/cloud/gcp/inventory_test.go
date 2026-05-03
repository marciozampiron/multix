// File: internal/adapters/outbound/cloud/gcp/inventory_test.go
// Purpose: Tests GCP Compute Engine + Cloud Storage inventory listing without live cloud dependencies.

package gcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"multix/internal/platform/logger"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newAdapterWithInventoryFakes(instances []gcpInstance, buckets []gcpBucket, projectID string) *Adapter {
	a := NewAdapter(logger.New("info"))
	a.findCredentialsFunc = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return &google.Credentials{
			ProjectID:   projectID,
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "x"}),
		}, nil
	}
	a.execCmdFunc = mockExecFail
	a.computeListFunc = func(ctx context.Context, p string) ([]gcpInstance, error) { return instances, nil }
	a.storageListFunc = func(ctx context.Context, p string) ([]gcpBucket, error) { return buckets, nil }
	return a
}

func TestGCPAdapter_List_Compute(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	instances := []gcpInstance{
		{ID: "111", Name: "gce-prod-api", Zone: "us-central1-a", Status: "RUNNING", Labels: map[string]string{"env": "prod"}, CreationTime: created},
		{ID: "222", Name: "gce-dev-api", Zone: "us-east1-b", Status: "TERMINATED"},
	}
	a := newAdapterWithInventoryFakes(instances, nil, "demo-project")

	resources, err := a.List(context.Background(), "compute")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(resources))
	}
	r := resources[0]
	if r.ID != "111" || r.Name != "gce-prod-api" || r.Region != "us-central1" || r.Status != "RUNNING" || r.Tags["env"] != "prod" || !r.CreatedAt.Equal(created) {
		t.Fatalf("unexpected first resource: %+v", r)
	}
	if r.AccountID != "demo-project" || r.Type != "computeEngine" {
		t.Fatalf("unexpected first metadata: %+v", r)
	}
	if resources[1].Region != "us-east1" {
		t.Fatalf("expected region us-east1, got %q", resources[1].Region)
	}
}

func TestGCPAdapter_List_Storage(t *testing.T) {
	created := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	buckets := []gcpBucket{
		{Name: "artifacts-prod", Location: "US", Created: created},
		{Name: "logs-eu", Location: "EU"},
	}
	a := newAdapterWithInventoryFakes(nil, buckets, "demo-project")

	resources, err := a.List(context.Background(), "storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(resources))
	}
	if resources[0].Type != "cloudStorage" || resources[0].Name != "artifacts-prod" || resources[0].Status != "AVAILABLE" || resources[0].Region != "US" {
		t.Fatalf("unexpected bucket resource: %+v", resources[0])
	}
}

func TestGCPAdapter_List_All(t *testing.T) {
	a := newAdapterWithInventoryFakes(
		[]gcpInstance{{ID: "1", Name: "gce-1", Zone: "us-central1-a", Status: "RUNNING"}},
		[]gcpBucket{{Name: "b-1", Location: "US"}, {Name: "b-2", Location: "EU"}},
		"demo-project",
	)
	resources, err := a.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 3 {
		t.Fatalf("expected 1 GCE + 2 GCS = 3 resources, got %d", len(resources))
	}
}

func TestGCPAdapter_List_UnknownType(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "demo-project")
	resources, err := a.List(context.Background(), "bigquery")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected empty result for unknown type, got %d", len(resources))
	}
}

func TestGCPAdapter_Scan(t *testing.T) {
	a := newAdapterWithInventoryFakes(
		[]gcpInstance{{ID: "1", Name: "gce-1", Zone: "us-central1-a"}, {ID: "2", Name: "gce-2", Zone: "us-east1-c"}},
		[]gcpBucket{{Name: "b-1"}},
		"demo-project",
	)
	summary, err := a.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.ProviderName != "gcp" || summary.Total != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.CountByType["computeEngine"] != 2 || summary.CountByType["cloudStorage"] != 1 {
		t.Fatalf("unexpected counts: %+v", summary.CountByType)
	}
}

func TestGCPAdapter_List_ComputeError(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "demo-project")
	a.computeListFunc = func(ctx context.Context, p string) ([]gcpInstance, error) {
		return nil, errors.New("api permission denied")
	}
	if _, err := a.List(context.Background(), "compute"); err == nil {
		t.Fatal("expected error from compute list failure")
	}
}

func TestGCPAdapter_RegionFromZone(t *testing.T) {
	cases := map[string]string{
		"us-central1-a": "us-central1",
		"europe-west4-b": "europe-west4",
		"asia-east1-c":  "asia-east1",
		"us-central1":   "us-central1", // not zone-shaped → unchanged
		"":              "",
	}
	for input, want := range cases {
		if got := regionFromZone(input); got != want {
			t.Errorf("regionFromZone(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestGCPAdapter_ResolveProjectID_FromCreds(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "explicit-project")
	got, err := a.resolveProjectID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "explicit-project" {
		t.Fatalf("expected explicit-project, got %q", got)
	}
}

func TestGCPAdapter_ResolveProjectID_FromEnv(t *testing.T) {
	a := newAdapterWithInventoryFakes(nil, nil, "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "env-project")
	got, err := a.resolveProjectID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "env-project" {
		t.Fatalf("expected env-project, got %q", got)
	}
}
