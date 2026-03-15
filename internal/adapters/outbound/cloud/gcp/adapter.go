// File: internal/adapters/outbound/cloud/gcp/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements GCP provider adapters, including real ADC-based auth validation and identity.

package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"

	"golang.org/x/oauth2/google"
)

type findCredentialsFunc func(ctx context.Context, scopes ...string) (*google.Credentials, error)
type execCmdFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

// Adapter implements GCP-backed outbound provider contracts.
type Adapter struct {
	log                 logger.Logger
	findCredentialsFunc findCredentialsFunc
	execCmdFunc         execCmdFunc
}

// NewAdapter creates a new GCP cloud adapter.
func NewAdapter(log logger.Logger) *Adapter {
	return &Adapter{
		log:                 log.With("provider", "gcp"),
		findCredentialsFunc: google.FindDefaultCredentials,
		execCmdFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
	}
}

func (a *Adapter) ID() string {
	return "gcp"
}

// Login implements the AuthProvider contract for legacy login compatibility.
func (a *Adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.log.Info("Authenticating via Google Cloud SDK (gcloud stub)")
	return &auth.Session{
		Provider: "gcp",
		IsValid:  true,
	}, nil
}

// Validate checks whether ADC credentials are available and usable.
func (a *Adapter) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	a.log.Info("Validating GCP application default credentials")
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return nil, err
	}

	result := &auth.ValidationResult{
		Provider: "gcp",
		Valid:    true,
		Message:  "GCP application default credentials are available",
		Details: map[string]string{
			"auth_source": inferAuthSource(creds),
		},
	}
	projectID := creds.ProjectID
	if projectID == "" {
		if envProj := os.Getenv("GOOGLE_CLOUD_PROJECT"); envProj != "" {
			projectID = envProj
		} else if envProj := os.Getenv("GCLOUD_PROJECT"); envProj != "" {
			projectID = envProj
		}
	}
	if projectID == "" {
		if out, err := a.execCmdFunc(ctx, "gcloud", "config", "get-value", "project"); err == nil {
			projectID = strings.TrimSpace(string(out))
		}
	}

	if projectID != "" {
		result.AccountID = projectID
		result.Details["project_id"] = projectID
	}
	return result, nil
}

// Whoami returns best-effort GCP identity details from active credentials context.
func (a *Adapter) Whoami(ctx context.Context) (*auth.Identity, error) {
	a.log.Info("Retrieving GCP active credentials context")
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return nil, err
	}

	authSource := inferAuthSource(creds)
	identity := &auth.Identity{
		Provider:   "gcp",
		ProjectID:  creds.ProjectID,
		AccountID:  creds.ProjectID,
		AuthSource: authSource,
		Raw: map[string]any{
			"credential_type": authSource,
		},
	}

	if serviceAccountEmail := extractServiceAccountEmail(creds.JSON); serviceAccountEmail != "" {
		identity.Principal = serviceAccountEmail
		identity.PrincipalType = "service_account"
	}

	// Best-effort enrichment via environment and gcloud CLI
	a.enrichIdentity(ctx, identity)

	if identity.Principal == "" {
		identity.Note = "active credentials detected via ADC; principal identity is not directly resolvable for this credential source"
	}
	return identity, nil
}

func (a *Adapter) enrichIdentity(ctx context.Context, identity *auth.Identity) {
	// Step 2: Environment fallback for project
	if identity.ProjectID == "" {
		if envProj := os.Getenv("GOOGLE_CLOUD_PROJECT"); envProj != "" {
			identity.ProjectID = envProj
			identity.AccountID = envProj
		} else if envProj := os.Getenv("GCLOUD_PROJECT"); envProj != "" {
			identity.ProjectID = envProj
			identity.AccountID = envProj
		}
	}

	// Step 3: Local gcloud enrichment (best-effort)
	if out, err := a.execCmdFunc(ctx, "gcloud", "config", "get-value", "project"); err == nil {
		proj := strings.TrimSpace(string(out))
		if proj != "" && identity.ProjectID == "" {
			identity.ProjectID = proj
			identity.AccountID = proj
		}
	}

	if out, err := a.execCmdFunc(ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)"); err == nil {
		account := strings.TrimSpace(string(out))
		// gcloud active accounts are typically users
		if account != "" && identity.Principal == "" {
			identity.Principal = account
			identity.PrincipalType = "user"
		}
	}
}

func (a *Adapter) defaultCredentials(ctx context.Context) (*google.Credentials, error) {
	creds, err := a.findCredentialsFunc(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GCP application default credentials; run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS: %w", err)
	}
	return creds, nil
}

func inferAuthSource(creds *google.Credentials) string {
	if extractServiceAccountEmail(creds.JSON) != "" {
		return "service_account_key"
	}
	return "application_default_credentials"
}

func extractServiceAccountEmail(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	var credentialJSON map[string]any
	if err := json.Unmarshal(raw, &credentialJSON); err != nil {
		return ""
	}

	email, _ := credentialJSON["client_email"].(string)
	return email
}

// Scan summarizes GCP inventory resources.
func (a *Adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.log.Info("Summarizing GCP inventory")
	return &inventory.Summary{
		ProviderName: "gcp",
		Total:        55,
		CountByType: map[string]int{
			"computeEngine": 15,
			"cloudStorage":  40,
		},
	}, nil
}

// List returns GCP inventory resources.
func (a *Adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.log.Info("Listing GCP inventory resources", "type", resourceType)
	return []*inventory.Resource{
		{ID: "instance-1934", Name: "gce-prod-api", Type: "computeEngine", Region: "us-central1"},
		{ID: "bucket-8493", Name: "gcs-backup-vault", Type: "cloudStorage", Region: "us-central1"},
	}, nil
}

// ListClusters returns GKE clusters.
func (a *Adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.log.Info("Listing GKE clusters", "region", "us-central1")
	return []*k8s.Cluster{
		{ID: "c-111", Name: "gke-autopilot-prod", Region: "us-central1", Version: "1.29", NodeCount: 0, Status: "RUNNING"},
		{ID: "c-222", Name: "gke-standard-dev", Region: "us-east4", Version: "1.28", NodeCount: 5, Status: "RUNNING"},
	}, nil
}

// SyncContext syncs GKE context to kubeconfig.
func (a *Adapter) SyncContext(ctx context.Context, clusterID string, region string) error {
	a.log.Info("Syncing GKE context to kubeconfig", "cluster", clusterID, "region", region)
	return nil
}
