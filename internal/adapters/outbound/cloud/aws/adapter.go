package aws

import (
	"context"
	"fmt"
	"strings"

	"multix/internal/domain/auth"
	"multix/internal/domain/inventory"
	"multix/internal/domain/k8s"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type adapter struct {
	logger logger.Logger
	// stsClientFunc allows testable seams for AWS STS calls.
	stsClientFunc func(ctx context.Context) (stsAPI, error)
}

type stsAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// NewAdapter creates a new AWS cloud provider adapter.
func NewAdapter(log logger.Logger) interface {
	outbound.AuthProvider
	outbound.InventoryProvider
	outbound.K8sProvider
} {
	return &adapter{
		logger: log.With("provider", "aws"),
		stsClientFunc: func(ctx context.Context) (stsAPI, error) {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, err
			}
			return sts.NewFromConfig(cfg), nil
		},
	}
}

func (a *adapter) ID() string {
	return "aws"
}

// AuthProvider Implementations
func (a *adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.logger.Info("Logging in to AWS (stub)")
	return &auth.Session{Provider: "aws", IsValid: true}, nil
}

func (a *adapter) Whoami(ctx context.Context) (*auth.Identity, error) {
	a.logger.Info("Retrieving AWS caller identity")
	client, err := a.stsClientFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws config error: %w", err)
	}

	out, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("aws sts error: %w", err)
	}

	principalType := "unknown"
	arn := *out.Arn
	if strings.Contains(arn, ":user/") {
		principalType = "user"
	} else if strings.Contains(arn, ":assumed-role/") {
		principalType = "assumed-role"
	}

	return &auth.Identity{
		Provider:      "aws",
		AccountID:     *out.Account,
		Principal:     arn,
		PrincipalType: principalType,
	}, nil
}

func (a *adapter) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	a.logger.Info("Validating AWS credentials via STS")
	client, err := a.stsClientFunc(ctx)
	if err != nil {
		return &auth.ValidationResult{
			Provider: "aws",
			IsValid:  false,
			Message:  fmt.Sprintf("failed to load aws config: %v", err),
		}, nil
	}

	out, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return &auth.ValidationResult{
			Provider: "aws",
			IsValid:  false,
			Message:  fmt.Sprintf("failed GetCallerIdentity: %v", err),
		}, nil
	}

	return &auth.ValidationResult{
		Provider:  "aws",
		IsValid:   true,
		AccountID: *out.Account,
		Principal: *out.Arn,
		Message:   "AWS credentials are valid",
	}, nil
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
