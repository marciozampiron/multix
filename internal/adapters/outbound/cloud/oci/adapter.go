// File: internal/adapters/outbound/cloud/oci/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Implements OCI provider adapters, including real auth validation and identity via OCI SDK.

package oci

import (
	"context"
	"fmt"
	"strings"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type adapter struct {
	logger logger.Logger
	// cfgProviderFunc allows testable seams for OCI configuration resolution.
	cfgProviderFunc func() common.ConfigurationProvider
	// identityClientFunc allows testable seams for OCI API calls.
	identityClientFunc func(cfg common.ConfigurationProvider) (identityAPI, error)
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
	return &adapter{
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

// List returns OCI inventory resources (stub).
func (a *adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.logger.Info("Listing OCI inventory resources", "type", resourceType)
	return []*inventory.Resource{
		inventory.NewResource("ocid1.instance.oc1..123", "us-ashburn-1", "Compute", "prod-web-server-oci"),
	}, nil
}

// Scan summarizes OCI inventory resources (stub).
func (a *adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.logger.Info("Scanning entire OCI tenancy inventory")
	return &inventory.Summary{ProviderName: "oci", Total: 8, CountByType: map[string]int{"Compute": 5, "Bucket": 3}}, nil
}

// ListClusters returns OKE clusters (stub).
func (a *adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.logger.Info("Listing OKE clusters", "region", "us-ashburn-1")
	return []*k8s.Cluster{
		{ID: "ocid1.cluster.oc1..abc", Name: "prod-oke-cluster", Region: "us-ashburn-1", Status: "ACTIVE", Version: "v1.30.1", NodeCount: 6},
	}, nil
}

// SyncContext syncs OKE context to kubeconfig (stub).
func (a *adapter) SyncContext(ctx context.Context, clusterName, region string) error {
	a.logger.Info("Generating kubeconfig for OKE cluster", "cluster", clusterName)
	return nil
}
