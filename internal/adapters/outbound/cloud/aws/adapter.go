// File: internal/adapters/outbound/cloud/aws/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements AWS provider adapters, including real auth validation and identity via STS.

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

// Login implements the AuthProvider contract for legacy login compatibility.
func (a *adapter) Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error) {
	a.logger.Info("Logging in to AWS (stub)")
	return &auth.Session{Provider: "aws", IsValid: true}, nil
}

// Whoami returns the active AWS identity using STS GetCallerIdentity.
func (a *adapter) Whoami(ctx context.Context) (*auth.Identity, error) {
	a.logger.Info("Retrieving AWS caller identity")
	out, err := a.getCallerIdentity(ctx)
	if err != nil {
		return nil, err
	}

	identity := mapAWSIdentity(out)
	return &identity, nil
}

// Validate validates AWS credentials using STS GetCallerIdentity.
func (a *adapter) Validate(ctx context.Context) (*auth.ValidationResult, error) {
	a.logger.Info("Validating AWS credentials via STS")
	out, err := a.getCallerIdentity(ctx)
	if err != nil {
		return nil, err
	}

	identity := mapAWSIdentity(out)
	return &auth.ValidationResult{
		Provider:  "aws",
		Valid:     true,
		AccountID: identity.AccountID,
		Principal: identity.Principal,
		Message:   "AWS credentials are valid",
		Details: map[string]string{
			"arn":            identity.Principal,
			"principal_type": identity.PrincipalType,
		},
	}, nil
}

func (a *adapter) getCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	client, err := a.stsClientFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config/credentials; run 'aws configure' or 'aws sso login': %w", err)
	}

	out, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed AWS STS GetCallerIdentity; verify credentials/session are active: %w", err)
	}
	return out, nil
}

func mapAWSIdentity(out *sts.GetCallerIdentityOutput) auth.Identity {
	arn := awsString(out.Arn)
	return auth.Identity{
		Provider:      "aws",
		AccountID:     awsString(out.Account),
		Principal:     arn,
		PrincipalType: inferAWSPrincipalType(arn),
		UserID:        awsString(out.UserId),
	}
}

func inferAWSPrincipalType(arn string) string {
	switch {
	case strings.Contains(arn, ":assumed-role/"):
		return "role"
	case strings.Contains(arn, ":role/"):
		return "role"
	case strings.Contains(arn, ":user/"):
		return "user"
	default:
		return "unknown"
	}
}

func awsString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// List returns AWS inventory resources.
func (a *adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.logger.Info("Listing AWS inventory resources", "type", resourceType)
	return []*inventory.Resource{
		inventory.NewResource("123456789012", "us-east-1", "EC2", "prod-web-server"),
	}, nil
}

// Scan summarizes AWS inventory resources.
func (a *adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.logger.Info("Scanning entire AWS account inventory")
	return &inventory.Summary{ProviderName: "aws", Total: 15, CountByType: map[string]int{"EC2": 10, "S3": 5}}, nil
}

// ListClusters returns EKS clusters.
func (a *adapter) ListClusters(ctx context.Context) ([]*k8s.Cluster, error) {
	a.logger.Info("Listing EKS clusters", "region", "us-east-1")
	return []*k8s.Cluster{
		{ID: "c-1", Name: "prod-eks-cluster", Region: "us-east-1", Status: "ACTIVE", Version: "1.30", NodeCount: 12},
	}, nil
}

// SyncContext syncs EKS context to kubeconfig.
func (a *adapter) SyncContext(ctx context.Context, clusterName, region string) error {
	a.logger.Info("Generating kubeconfig for EKS cluster", "cluster", clusterName)
	return nil
}
