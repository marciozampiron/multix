package gcp

import (
	"context"

	"encoding/json"
	"fmt"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"

	"golang.org/x/oauth2/google"
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

func (a *Adapter) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	a.log.Info("Validating GCP application default credentials", "provider", "gcp")
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return &auth.ValidationResult{
			Provider: "gcp",
			IsValid:  false,
			Message:  fmt.Sprintf("ADC not found or invalid: %v", err),
		}, nil
	}

	return &auth.ValidationResult{
		Provider:  "gcp",
		IsValid:   true,
		AccountID: creds.ProjectID,
		Message:   "Valid application default credentials found",
	}, nil
}

func (a *Adapter) Whoami(ctx context.Context) (*auth.Identity, error) {
	a.log.Info("Retrieving GCP active credentials context", "provider", "gcp")
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to find default credentials: %w", err)
	}

	principal := "unknown"
	authSource := "application_default_credentials"

	var credJSON map[string]any
	if len(creds.JSON) > 0 {
		if err := json.Unmarshal(creds.JSON, &credJSON); err == nil {
			if email, ok := credJSON["client_email"].(string); ok {
				principal = email
				authSource = "service_account_key"
			}
		}
	}

	return &auth.Identity{
		Provider:   "gcp",
		AccountID:  creds.ProjectID,
		ProjectID:  creds.ProjectID,
		Principal:  principal,
		AuthSource: authSource,
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
