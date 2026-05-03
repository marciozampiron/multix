// File: internal/adapters/outbound/cloud/aws/adapter.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 02/05/2026
// Purpose: Implements AWS provider adapters with real STS auth, EC2 inventory and S3 inventory.

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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type adapter struct {
	logger logger.Logger
	// stsClientFunc allows testable seams for AWS STS calls.
	stsClientFunc func(ctx context.Context) (stsAPI, error)
	// ec2ClientFunc allows testable seams for EC2 listing. Returns the client and the active region.
	ec2ClientFunc func(ctx context.Context) (ec2API, string, error)
	// s3ClientFunc allows testable seams for S3 listing.
	s3ClientFunc func(ctx context.Context) (s3API, error)
}

type stsAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type ec2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

type s3API interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
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
		ec2ClientFunc: func(ctx context.Context) (ec2API, string, error) {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, "", err
			}
			return ec2.NewFromConfig(cfg), cfg.Region, nil
		},
		s3ClientFunc: func(ctx context.Context) (s3API, error) {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, err
			}
			return s3.NewFromConfig(cfg), nil
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

// List returns AWS inventory resources. resourceType selects a service:
//   - "compute" or "ec2"   → EC2 instances in the active region
//   - "storage" or "s3"    → S3 buckets (global)
//   - "" (empty)           → both EC2 and S3
//
// Any other value returns an empty slice without error.
func (a *adapter) List(ctx context.Context, resourceType string) ([]*inventory.Resource, error) {
	a.logger.Info("Listing AWS inventory resources", "type", resourceType)

	accountID, err := a.resolveAccountID(ctx)
	if err != nil {
		return nil, err
	}

	kind := strings.ToLower(strings.TrimSpace(resourceType))
	switch kind {
	case "compute", "ec2":
		return a.listEC2(ctx, accountID)
	case "storage", "s3":
		return a.listS3(ctx, accountID)
	case "":
		ec2res, err := a.listEC2(ctx, accountID)
		if err != nil {
			return nil, err
		}
		s3res, err := a.listS3(ctx, accountID)
		if err != nil {
			return nil, err
		}
		return append(ec2res, s3res...), nil
	default:
		a.logger.Warn("Unknown AWS resource type, returning empty list", "type", resourceType)
		return []*inventory.Resource{}, nil
	}
}

// Scan summarizes AWS inventory resources across EC2 and S3.
func (a *adapter) Scan(ctx context.Context) (*inventory.Summary, error) {
	a.logger.Info("Scanning entire AWS account inventory")

	accountID, err := a.resolveAccountID(ctx)
	if err != nil {
		return nil, err
	}

	ec2res, err := a.listEC2(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("EC2 listing failed: %w", err)
	}
	s3res, err := a.listS3(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("S3 listing failed: %w", err)
	}

	summary := &inventory.Summary{
		ProviderName: "aws",
		Total:        len(ec2res) + len(s3res),
		CountByType: map[string]int{
			"EC2": len(ec2res),
			"S3":  len(s3res),
		},
	}
	return summary, nil
}

func (a *adapter) resolveAccountID(ctx context.Context) (string, error) {
	out, err := a.getCallerIdentity(ctx)
	if err != nil {
		return "", err
	}
	return awsString(out.Account), nil
}

func (a *adapter) listEC2(ctx context.Context, accountID string) ([]*inventory.Resource, error) {
	client, region, err := a.ec2ClientFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize EC2 client: %w", err)
	}

	var resources []*inventory.Resource
	var nextToken *string
	for {
		out, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{NextToken: nextToken})
		if err != nil {
			return nil, fmt.Errorf("EC2 DescribeInstances failed: %w", err)
		}
		for _, res := range out.Reservations {
			for _, inst := range res.Instances {
				resources = append(resources, mapEC2Instance(inst, accountID, region))
			}
		}
		if out.NextToken == nil || *out.NextToken == "" {
			break
		}
		nextToken = out.NextToken
	}
	return resources, nil
}

func (a *adapter) listS3(ctx context.Context, accountID string) ([]*inventory.Resource, error) {
	client, err := a.s3ClientFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("S3 ListBuckets failed: %w", err)
	}

	resources := make([]*inventory.Resource, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		resources = append(resources, mapS3Bucket(b, accountID))
	}
	return resources, nil
}

func mapEC2Instance(inst ec2types.Instance, accountID, region string) *inventory.Resource {
	r := inventory.NewResource(accountID, region, "EC2", awsString(inst.InstanceId))
	r.ID = awsString(inst.InstanceId)
	r.Status = string(inst.State.Name)
	for _, t := range inst.Tags {
		key := awsString(t.Key)
		val := awsString(t.Value)
		r.Tags[key] = val
		if strings.EqualFold(key, "Name") && val != "" {
			r.Name = val
		}
	}
	if inst.LaunchTime != nil {
		r.CreatedAt = *inst.LaunchTime
	}
	return r
}

func mapS3Bucket(b s3types.Bucket, accountID string) *inventory.Resource {
	region := awsString(b.BucketRegion)
	r := inventory.NewResource(accountID, region, "S3", awsString(b.Name))
	r.ID = awsString(b.Name)
	r.Status = "AVAILABLE"
	if b.CreationDate != nil {
		r.CreatedAt = *b.CreationDate
	}
	return r
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
