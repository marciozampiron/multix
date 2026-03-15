package aws

import (
	"context"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"
)

type adapter struct {
	logger logger.Logger
}

func NewAdapter(log logger.Logger) interface {
	outbound.AuthProvider
	outbound.InventoryProvider
	outbound.K8sProvider
} {
	return &adapter{
		logger: log.With("provider", "aws"),
	}
}

func (a *adapter) ID() string {
	return "aws"
}

// AuthProvider Implementations
func (a *adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.logger.Info("Logging in to AWS (stub)")
	return &auth.Session{AccountID: "123456789012", Username: "admin", Role: "AdminRole", IsValid: true, Provider: "aws"}, nil
}

func (a *adapter) Whoami(ctx context.Context) (*auth.Session, error) {
	return &auth.Session{AccountID: "123456789012", Username: "admin", Role: "AdminRole", IsValid: true, Provider: "aws"}, nil
}

func (a *adapter) Validate(ctx context.Context) (bool, error) {
	return true, nil
}

// InventoryProvider Implementations
func (a *adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.logger.Info("Listing AWS inventory resources", "type", resourceType)
	return []*inventory.Resource{
		inventory.NewResource("123456789012", "us-east-1", "EC2", "prod-web-server"),
	}, nil
}

func (a *adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.logger.Info("Scanning entire AWS account inventory")
	return &inventory.Summary{ProviderName: "aws", Total: 15, CountByType: map[string]int{"EC2": 10, "S3": 5}}, nil
}

// K8sProvider Implementations
func (a *adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.logger.Info("Listing EKS clusters", "region", "us-east-1")
	return []*k8s.Cluster{
		{ID: "c-1", Name: "prod-eks-cluster", Region: "us-east-1", Status: "ACTIVE", Version: "1.30", NodeCount: 12},
	}, nil
}

func (a *adapter) SyncContext(ctx context.Context, clusterName, region string) error {
	a.logger.Info("Generating kubeconfig for EKS cluster", "cluster", clusterName)
	return nil
}
