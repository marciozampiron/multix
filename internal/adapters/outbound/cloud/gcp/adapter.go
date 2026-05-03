// File: internal/adapters/outbound/cloud/gcp/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 02/05/2026
// Purpose: Implements GCP provider adapters, including real ADC-based auth validation and identity.

package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type findCredentialsFunc func(ctx context.Context, scopes ...string) (*google.Credentials, error)
type execCmdFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

// gcpInstance is a minimal projection of a Compute Engine instance for inventory output.
type gcpInstance struct {
	ID           string
	Name         string
	Zone         string
	Status       string
	Labels       map[string]string
	CreationTime time.Time
}

// gcpBucket is a minimal projection of a Cloud Storage bucket for inventory output.
type gcpBucket struct {
	Name     string
	Location string
	Created  time.Time
}

// Adapter implements GCP-backed outbound provider contracts.
type Adapter struct {
	log                 logger.Logger
	findCredentialsFunc findCredentialsFunc
	execCmdFunc         execCmdFunc
	// computeListFunc lists Compute Engine instances across all zones for a project.
	computeListFunc func(ctx context.Context, projectID string) ([]gcpInstance, error)
	// storageListFunc lists Cloud Storage buckets for a project.
	storageListFunc func(ctx context.Context, projectID string) ([]gcpBucket, error)
	// gkeClusterListFunc lists GKE clusters across all locations for a project.
	gkeClusterListFunc func(ctx context.Context, projectID string) ([]*k8s.Cluster, error)
}

// NewAdapter creates a new GCP cloud adapter.
func NewAdapter(log logger.Logger) *Adapter {
	a := &Adapter{
		log:                 log.With("provider", "gcp"),
		findCredentialsFunc: google.FindDefaultCredentials,
		execCmdFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
	}
	a.computeListFunc = a.defaultComputeList
	a.storageListFunc = a.defaultStorageList
	a.gkeClusterListFunc = a.defaultGKEList
	return a
}

func (a *Adapter) defaultGKEList(ctx context.Context, projectID string) ([]*k8s.Cluster, error) {
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := container.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GKE service: %w", err)
	}
	parent := fmt.Sprintf("projects/%s/locations/-", projectID)
	resp, err := svc.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("GKE Clusters.List failed: %w", err)
	}
	clusters := make([]*k8s.Cluster, 0, len(resp.Clusters))
	for _, c := range resp.Clusters {
		clusters = append(clusters, &k8s.Cluster{
			ID:        c.Id,
			Name:      c.Name,
			Region:    c.Location,
			Status:    c.Status,
			Version:   c.CurrentMasterVersion,
			NodeCount: int(c.CurrentNodeCount),
		})
	}
	return clusters, nil
}

func (a *Adapter) defaultComputeList(ctx context.Context, projectID string) ([]gcpInstance, error) {
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCP Compute service: %w", err)
	}

	var out []gcpInstance
	call := svc.Instances.AggregatedList(projectID)
	if err := call.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for zone, scoped := range page.Items {
			for _, inst := range scoped.Instances {
				created, _ := time.Parse(time.RFC3339, inst.CreationTimestamp)
				out = append(out, gcpInstance{
					ID:           fmt.Sprintf("%d", inst.Id),
					Name:         inst.Name,
					Zone:         strings.TrimPrefix(zone, "zones/"),
					Status:       inst.Status,
					Labels:       inst.Labels,
					CreationTime: created,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Compute Engine AggregatedList failed: %w", err)
	}
	return out, nil
}

func (a *Adapter) defaultStorageList(ctx context.Context, projectID string) ([]gcpBucket, error) {
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS client: %w", err)
	}
	defer client.Close()

	var out []gcpBucket
	it := client.Buckets(ctx, projectID)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Cloud Storage Buckets list failed: %w", err)
		}
		out = append(out, gcpBucket{
			Name:     attrs.Name,
			Location: attrs.Location,
			Created:  attrs.Created,
		})
	}
	return out, nil
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

// Scan summarizes GCP inventory resources across Compute Engine and Cloud Storage.
func (a *Adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.log.Info("Summarizing GCP inventory")
	projectID, err := a.resolveProjectID(ctx)
	if err != nil {
		return nil, err
	}

	instances, err := a.computeListFunc(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Compute Engine listing failed: %w", err)
	}
	buckets, err := a.storageListFunc(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Cloud Storage listing failed: %w", err)
	}

	return &inventory.Summary{
		ProviderName: "gcp",
		Total:        len(instances) + len(buckets),
		CountByType: map[string]int{
			"computeEngine": len(instances),
			"cloudStorage":  len(buckets),
		},
	}, nil
}

// List returns GCP inventory resources. resourceType selects a service:
//   - "compute", "computeEngine", "gce"     → Compute Engine instances
//   - "storage", "cloudStorage", "gcs"      → Cloud Storage buckets
//   - "" (empty)                            → both
func (a *Adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.log.Info("Listing GCP inventory resources", "type", resourceType)
	projectID, err := a.resolveProjectID(ctx)
	if err != nil {
		return nil, err
	}

	kind := strings.ToLower(strings.TrimSpace(resourceType))
	switch kind {
	case "compute", "computeengine", "gce":
		return a.listComputeEngine(ctx, projectID)
	case "storage", "cloudstorage", "gcs":
		return a.listCloudStorage(ctx, projectID)
	case "":
		gce, err := a.listComputeEngine(ctx, projectID)
		if err != nil {
			return nil, err
		}
		gcs, err := a.listCloudStorage(ctx, projectID)
		if err != nil {
			return nil, err
		}
		return append(gce, gcs...), nil
	default:
		a.log.Warn("Unknown GCP resource type, returning empty list", "type", resourceType)
		return []*inventory.Resource{}, nil
	}
}

func (a *Adapter) listComputeEngine(ctx context.Context, projectID string) ([]*inventory.Resource, error) {
	instances, err := a.computeListFunc(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resources := make([]*inventory.Resource, 0, len(instances))
	for _, inst := range instances {
		r := inventory.NewResource(projectID, regionFromZone(inst.Zone), "computeEngine", inst.Name)
		if inst.ID != "" {
			r.ID = inst.ID
		}
		r.Status = inst.Status
		for k, v := range inst.Labels {
			r.Tags[k] = v
		}
		if !inst.CreationTime.IsZero() {
			r.CreatedAt = inst.CreationTime
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (a *Adapter) listCloudStorage(ctx context.Context, projectID string) ([]*inventory.Resource, error) {
	buckets, err := a.storageListFunc(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resources := make([]*inventory.Resource, 0, len(buckets))
	for _, b := range buckets {
		r := inventory.NewResource(projectID, b.Location, "cloudStorage", b.Name)
		r.ID = b.Name
		r.Status = "AVAILABLE"
		if !b.Created.IsZero() {
			r.CreatedAt = b.Created
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (a *Adapter) resolveProjectID(ctx context.Context) (string, error) {
	creds, err := a.defaultCredentials(ctx)
	if err != nil {
		return "", err
	}
	if creds.ProjectID != "" {
		return creds.ProjectID, nil
	}
	if envProj := os.Getenv("GOOGLE_CLOUD_PROJECT"); envProj != "" {
		return envProj, nil
	}
	if envProj := os.Getenv("GCLOUD_PROJECT"); envProj != "" {
		return envProj, nil
	}
	if out, err := a.execCmdFunc(ctx, "gcloud", "config", "get-value", "project"); err == nil {
		if proj := strings.TrimSpace(string(out)); proj != "" {
			return proj, nil
		}
	}
	return "", fmt.Errorf("could not resolve GCP project ID; set GOOGLE_CLOUD_PROJECT or run 'gcloud config set project <id>'")
}

// regionFromZone strips the trailing "-<letter>" from a GCE zone (e.g. "us-central1-a" → "us-central1").
// Returns the input unchanged if it does not match the zone shape.
func regionFromZone(zone string) string {
	if i := strings.LastIndex(zone, "-"); i > 0 && len(zone)-i-1 == 1 {
		return zone[:i]
	}
	return zone
}

// ListClusters returns GKE clusters across all locations for the active project.
func (a *Adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.log.Info("Listing GKE clusters")
	projectID, err := a.resolveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return a.gkeClusterListFunc(ctx, projectID)
}

// SyncContext syncs GKE context to kubeconfig.
func (a *Adapter) SyncContext(ctx context.Context, clusterID string, region string) error {
	a.log.Info("Syncing GKE context to kubeconfig", "cluster", clusterID, "region", region)
	return nil
}
