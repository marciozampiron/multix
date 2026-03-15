package gcp

import (
	"context"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
)

type Adapter struct {
	log logger.Logger
}

func NewAdapter(log logger.Logger) *Adapter {
	return &Adapter{log: log}
}

func (a *Adapter) ID() string {
	return "gcp"
}

// AuthProvider Implementation
func (a *Adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.log.Info("Authenticating via Google Cloud SDK (gcloud stub)", "provider", "gcp")
	return &auth.Session{
		Provider: "gcp",
		IsValid:  true,
	}, nil
}

func (a *Adapter) Validate(ctx context.Context) (bool, error) {
	a.log.Info("Validating GCP application default credentials", "provider", "gcp")
	return true, nil
}

func (a *Adapter) Whoami(ctx context.Context) (*auth.Session, error) {
	a.log.Info("Retrieving GCP active account", "provider", "gcp")
	return &auth.Session{
		Provider:  "gcp",
		AccountID: "my-gcp-project-123",
		Username:  "stub-user@example.com",
		Role:      "roles/editor",
		IsValid:   true,
	}, nil
}

// InventoryProvider Implementation
func (a *Adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.log.Info("Summarizing GCP inventory", "provider", "gcp")
	return &inventory.Summary{
		ProviderName: "gcp",
		Total:        55,
		CountByType: map[string]int{
			"computeEngine": 15,
			"cloudStorage":  40,
		},
	}, nil
}

func (a *Adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.log.Info("Listing GCP inventory resources", "provider", "gcp", "type", resourceType)
	return []*inventory.Resource{
		{ID: "instance-1934", Name: "gce-prod-api", Type: "computeEngine", Region: "us-central1"},
		{ID: "bucket-8493", Name: "gcs-backup-vault", Type: "cloudStorage", Region: "us-central1"},
	}, nil
}

// K8sProvider Implementation
func (a *Adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.log.Info("Listing GKE clusters", "provider", "gcp", "region", "us-central1")
	return []*k8s.Cluster{
		{ID: "c-111", Name: "gke-autopilot-prod", Region: "us-central1", Version: "1.29", NodeCount: 0, Status: "RUNNING"},
		{ID: "c-222", Name: "gke-standard-dev", Region: "us-east4", Version: "1.28", NodeCount: 5, Status: "RUNNING"},
	}, nil
}

func (a *Adapter) SyncContext(ctx context.Context, clusterID string, region string) error {
	a.log.Info("Syncing GKE context to kubeconfig", "cluster", clusterID, "region", region)
	return nil
}
