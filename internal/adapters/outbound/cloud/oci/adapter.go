// File: internal/adapters/outbound/cloud/oci/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 02/05/2026
// Purpose: Implements OCI provider adapters, including real auth validation, identity, Compute and Object Storage inventory.

package oci

import (
	"context"
	"fmt"
	"strings"
	"time"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// ociInstance is a minimal projection of a Compute instance for inventory output.
type ociInstance struct {
	ID            string
	Name          string
	Region        string
	LifecycleState string
	FreeformTags  map[string]string
	TimeCreated   time.Time
}

// ociBucket is a minimal projection of an Object Storage bucket for inventory output.
type ociBucket struct {
	Name        string
	Namespace   string
	Region      string
	TimeCreated time.Time
}

type adapter struct {
	logger logger.Logger
	// cfgProviderFunc allows testable seams for OCI configuration resolution.
	cfgProviderFunc func() common.ConfigurationProvider
	// identityClientFunc allows testable seams for OCI API calls.
	identityClientFunc func(cfg common.ConfigurationProvider) (identityAPI, error)
	// computeListFunc lists Compute instances under a compartment (typically tenancy OCID).
	computeListFunc func(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]ociInstance, error)
	// bucketListFunc lists Object Storage buckets under a compartment.
	bucketListFunc func(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]ociBucket, error)
	// okeClusterListFunc lists OKE clusters under a compartment.
	okeClusterListFunc func(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]*k8s.Cluster, error)
}

// identityAPI defines the interface we need from OCI's identity client to make testing easier.
type identityAPI interface {
	GetUser(ctx context.Context, request identity.GetUserRequest) (response identity.GetUserResponse, err error)
}

// NewAdapter creates a new OCI cloud provider adapter.
func NewAdapter(log logger.Logger) interface {
	outbound.AuthProvider
	outbound.InventoryProvider
	outbound.K8sProvider
} {
	a := &adapter{
		logger: log.With("provider", "oci"),
		cfgProviderFunc: func() common.ConfigurationProvider {
			return common.DefaultConfigProvider()
		},
		identityClientFunc: func(cfg common.ConfigurationProvider) (identityAPI, error) {
			client, err := identity.NewIdentityClientWithConfigurationProvider(cfg)
			if err != nil {
				return nil, err
			}
			return &client, nil
		},
	}
	a.computeListFunc = defaultComputeList
	a.bucketListFunc = defaultBucketList
	a.okeClusterListFunc = defaultOKEList
	return a
}

func defaultOKEList(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]*k8s.Cluster, error) {
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OKE client: %w", err)
	}
	region, _ := cfg.Region()

	var out []*k8s.Cluster
	var page *string
	for {
		req := containerengine.ListClustersRequest{CompartmentId: common.String(compartmentID), Page: page}
		resp, err := client.ListClusters(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("OKE ListClusters failed: %w", err)
		}
		for _, c := range resp.Items {
			out = append(out, &k8s.Cluster{
				ID:      stringValue(c.Id),
				Name:    stringValue(c.Name),
				Region:  region,
				Status:  string(c.LifecycleState),
				Version: stringValue(c.KubernetesVersion),
			})
		}
		if resp.OpcNextPage == nil || *resp.OpcNextPage == "" {
			break
		}
		page = resp.OpcNextPage
	}
	return out, nil
}

func defaultComputeList(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]ociInstance, error) {
	client, err := core.NewComputeClientWithConfigurationProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OCI Compute client: %w", err)
	}
	region, _ := cfg.Region()

	var out []ociInstance
	var page *string
	for {
		req := core.ListInstancesRequest{CompartmentId: common.String(compartmentID), Page: page}
		resp, err := client.ListInstances(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("OCI ListInstances failed: %w", err)
		}
		for _, inst := range resp.Items {
			out = append(out, ociInstance{
				ID:             stringValue(inst.Id),
				Name:           stringValue(inst.DisplayName),
				Region:         region,
				LifecycleState: string(inst.LifecycleState),
				FreeformTags:   inst.FreeformTags,
				TimeCreated:    timeValue(inst.TimeCreated),
			})
		}
		if resp.OpcNextPage == nil || *resp.OpcNextPage == "" {
			break
		}
		page = resp.OpcNextPage
	}
	return out, nil
}

func defaultBucketList(ctx context.Context, cfg common.ConfigurationProvider, compartmentID string) ([]ociBucket, error) {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OCI Object Storage client: %w", err)
	}
	region, _ := cfg.Region()

	nsResp, err := client.GetNamespace(ctx, objectstorage.GetNamespaceRequest{})
	if err != nil {
		return nil, fmt.Errorf("OCI GetNamespace failed: %w", err)
	}
	namespace := stringValue(nsResp.Value)

	var out []ociBucket
	var page *string
	for {
		req := objectstorage.ListBucketsRequest{
			NamespaceName: common.String(namespace),
			CompartmentId: common.String(compartmentID),
			Page:          page,
		}
		resp, err := client.ListBuckets(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("OCI ListBuckets failed: %w", err)
		}
		for _, b := range resp.Items {
			out = append(out, ociBucket{
				Name:        stringValue(b.Name),
				Namespace:   stringValue(b.Namespace),
				Region:      region,
				TimeCreated: timeValue(b.TimeCreated),
			})
		}
		if resp.OpcNextPage == nil || *resp.OpcNextPage == "" {
			break
		}
		page = resp.OpcNextPage
	}
	return out, nil
}

func stringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func timeValue(t *common.SDKTime) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Time
}

func (a *adapter) ID() string {
	return "oci"
}

// Login implements the AuthProvider contract for legacy login compatibility.
func (a *adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.logger.Info("Logging in to OCI (stub)")
	return &auth.Session{Provider: "oci", IsValid: true}, nil
}

// Whoami returns the active OCI identity by inspecting the configuration provider.
func (a *adapter) Whoami(ctx context.Context) (*auth.Identity, error) {
	a.logger.Info("Retrieving OCI caller identity")
	
	cfg := a.cfgProviderFunc()
	
	tenancyID, err := cfg.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OCI tenancy OCID: %w", err)
	}
	
	userID, err := cfg.UserOCID()
	if err != nil {
		// In instance principal or resource principal scenarios, this might be empty
		// We can handle fallback gracefully if necessary, but for v0.4 assume local user cfg
		return nil, fmt.Errorf("failed to retrieve OCI user OCID: %w", err)
	}

	return &auth.Identity{
		Provider:      "oci",
		AccountID:     tenancyID,
		Principal:     userID,
		PrincipalType: inferOCIPrincipalType(userID),
		UserID:        userID,
	}, nil
}

// Validate validates OCI credentials by making a real API call to GetUser.
func (a *adapter) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	a.logger.Info("Validating OCI credentials via Identity API")
	
	cfg := a.cfgProviderFunc()
	client, err := a.identityClientFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OCI identity client: %w", err)
	}

	tenancyID, err := cfg.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("invalid OCI configuration (missing tenancy): %w", err)
	}

	userID, err := cfg.UserOCID()
	if err != nil {
		return nil, fmt.Errorf("invalid OCI configuration (missing user): %w", err)
	}

	// Make an actual API call to OCI to prove credentials are valid
	req := identity.GetUserRequest{UserId: common.String(userID)}
	_, err = client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate OCI credentials; API call rejected: %w", err)
	}

	return &auth.ValidationResult{
		Provider:  "oci",
		Valid:     true,
		AccountID: tenancyID,
		Principal: userID,
		Message:   "OCI credentials are valid and active",
		Details: map[string]string{
			"tenancy_ocid": tenancyID,
			"user_ocid":    userID,
		},
	}, nil
}

func inferOCIPrincipalType(ocid string) string {
	if strings.Contains(ocid, "ocid1.user.") {
		return "user"
	}
	if strings.Contains(ocid, "ocid1.instance.") {
		return "instance_principal"
	}
	return "unknown"
}

// List returns OCI inventory resources. resourceType selects a service:
//   - "compute", "instance"     → Compute instances under tenancy
//   - "storage", "bucket", "objectstorage" → Object Storage buckets under tenancy
//   - "" (empty)                → both
func (a *adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.logger.Info("Listing OCI inventory resources", "type", resourceType)

	cfg := a.cfgProviderFunc()
	tenancyID, err := cfg.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("invalid OCI configuration (missing tenancy): %w", err)
	}

	kind := strings.ToLower(strings.TrimSpace(resourceType))
	switch kind {
	case "compute", "instance":
		return a.listCompute(ctx, cfg, tenancyID)
	case "storage", "bucket", "objectstorage":
		return a.listBuckets(ctx, cfg, tenancyID)
	case "":
		compute, err := a.listCompute(ctx, cfg, tenancyID)
		if err != nil {
			return nil, err
		}
		buckets, err := a.listBuckets(ctx, cfg, tenancyID)
		if err != nil {
			return nil, err
		}
		return append(compute, buckets...), nil
	default:
		a.logger.Warn("Unknown OCI resource type, returning empty list", "type", resourceType)
		return []*inventory.Resource{}, nil
	}
}

// Scan summarizes OCI inventory resources across Compute and Object Storage.
func (a *adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.logger.Info("Scanning entire OCI tenancy inventory")

	cfg := a.cfgProviderFunc()
	tenancyID, err := cfg.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("invalid OCI configuration (missing tenancy): %w", err)
	}

	instances, err := a.computeListFunc(ctx, cfg, tenancyID)
	if err != nil {
		return nil, fmt.Errorf("Compute listing failed: %w", err)
	}
	buckets, err := a.bucketListFunc(ctx, cfg, tenancyID)
	if err != nil {
		return nil, fmt.Errorf("Object Storage listing failed: %w", err)
	}

	return &inventory.Summary{
		ProviderName: "oci",
		Total:        len(instances) + len(buckets),
		CountByType: map[string]int{
			"Compute": len(instances),
			"Bucket":  len(buckets),
		},
	}, nil
}

func (a *adapter) listCompute(ctx context.Context, cfg common.ConfigurationProvider, tenancyID string) ([]*inventory.Resource, error) {
	instances, err := a.computeListFunc(ctx, cfg, tenancyID)
	if err != nil {
		return nil, err
	}
	resources := make([]*inventory.Resource, 0, len(instances))
	for _, inst := range instances {
		r := inventory.NewResource(tenancyID, inst.Region, "Compute", inst.Name)
		if inst.ID != "" {
			r.ID = inst.ID
		}
		r.Status = inst.LifecycleState
		for k, v := range inst.FreeformTags {
			r.Tags[k] = v
		}
		if !inst.TimeCreated.IsZero() {
			r.CreatedAt = inst.TimeCreated
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (a *adapter) listBuckets(ctx context.Context, cfg common.ConfigurationProvider, tenancyID string) ([]*inventory.Resource, error) {
	buckets, err := a.bucketListFunc(ctx, cfg, tenancyID)
	if err != nil {
		return nil, err
	}
	resources := make([]*inventory.Resource, 0, len(buckets))
	for _, b := range buckets {
		r := inventory.NewResource(tenancyID, b.Region, "Bucket", b.Name)
		r.ID = b.Name
		r.Status = "AVAILABLE"
		if b.Namespace != "" {
			r.Tags["namespace"] = b.Namespace
		}
		if !b.TimeCreated.IsZero() {
			r.CreatedAt = b.TimeCreated
		}
		resources = append(resources, r)
	}
	return resources, nil
}

// ListClusters returns OKE clusters under the active tenancy compartment.
func (a *adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.logger.Info("Listing OKE clusters")
	cfg := a.cfgProviderFunc()
	tenancyID, err := cfg.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("invalid OCI configuration (missing tenancy): %w", err)
	}
	return a.okeClusterListFunc(ctx, cfg, tenancyID)
}

// SyncContext syncs OKE context to kubeconfig (stub).
func (a *adapter) SyncContext(ctx context.Context, clusterName, region string) error {
	a.logger.Info("Generating kubeconfig for OKE cluster", "cluster", clusterName)
	return nil
}
