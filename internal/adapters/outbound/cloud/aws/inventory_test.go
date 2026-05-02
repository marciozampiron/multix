// File: internal/adapters/outbound/cloud/aws/inventory_test.go
// Purpose: Tests AWS EC2 + S3 inventory listing without live cloud dependencies.

package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"multix/internal/platform/logger"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type fakeEC2Client struct {
	pages []*ec2.DescribeInstancesOutput
	err   error
	calls int
}

func (f *fakeEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.calls >= len(f.pages) {
		return &ec2.DescribeInstancesOutput{}, nil
	}
	out := f.pages[f.calls]
	f.calls++
	return out, nil
}

type fakeS3Client struct {
	output *s3.ListBucketsOutput
	err    error
}

func (f *fakeS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

func newAdapterWithFakes(stsOut *sts.GetCallerIdentityOutput, ec2Pages []*ec2.DescribeInstancesOutput, s3Out *s3.ListBucketsOutput) *adapter {
	a := NewAdapter(logger.New("info")).(*adapter)
	a.stsClientFunc = func(ctx context.Context) (stsAPI, error) {
		return &fakeSTSClient{output: stsOut}, nil
	}
	a.ec2ClientFunc = func(ctx context.Context) (ec2API, string, error) {
		return &fakeEC2Client{pages: ec2Pages}, "us-east-1", nil
	}
	a.s3ClientFunc = func(ctx context.Context) (s3API, error) {
		return &fakeS3Client{output: s3Out}, nil
	}
	return a
}

func ptr[T any](v T) *T { return &v }

func TestAWSAdapter_List_EC2(t *testing.T) {
	account := "123456789012"
	stsOut := &sts.GetCallerIdentityOutput{Account: &account}

	page1 := &ec2.DescribeInstancesOutput{
		NextToken: ptr("token-2"),
		Reservations: []ec2types.Reservation{{
			Instances: []ec2types.Instance{{
				InstanceId: ptr("i-aaa"),
				State:      &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning},
				LaunchTime: ptr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
				Tags:       []ec2types.Tag{{Key: ptr("Name"), Value: ptr("web-1")}, {Key: ptr("Env"), Value: ptr("prod")}},
			}},
		}},
	}
	page2 := &ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{{
			Instances: []ec2types.Instance{{
				InstanceId: ptr("i-bbb"),
				State:      &ec2types.InstanceState{Name: ec2types.InstanceStateNameStopped},
			}},
		}},
	}

	a := newAdapterWithFakes(stsOut, []*ec2.DescribeInstancesOutput{page1, page2}, nil)

	resources, err := a.List(context.Background(), "compute")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 instances across pages, got %d", len(resources))
	}
	first := resources[0]
	if first.ID != "i-aaa" || first.Name != "web-1" || first.Status != "running" || first.Tags["Env"] != "prod" {
		t.Fatalf("unexpected first resource: %+v", first)
	}
	if first.AccountID != account || first.Region != "us-east-1" || first.Type != "EC2" {
		t.Fatalf("unexpected metadata on first resource: %+v", first)
	}
	if resources[1].Status != "stopped" {
		t.Fatalf("expected second instance status 'stopped', got %q", resources[1].Status)
	}
}

func TestAWSAdapter_List_S3(t *testing.T) {
	account := "123456789012"
	stsOut := &sts.GetCallerIdentityOutput{Account: &account}
	s3Out := &s3.ListBucketsOutput{
		Buckets: []s3types.Bucket{
			{Name: ptr("artifacts-prod"), BucketRegion: ptr("us-east-1"), CreationDate: ptr(time.Now())},
			{Name: ptr("logs-prod"), BucketRegion: ptr("eu-west-1")},
		},
	}

	a := newAdapterWithFakes(stsOut, nil, s3Out)

	resources, err := a.List(context.Background(), "storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(resources))
	}
	if resources[0].Type != "S3" || resources[0].Name != "artifacts-prod" || resources[0].Status != "AVAILABLE" {
		t.Fatalf("unexpected bucket resource: %+v", resources[0])
	}
	if resources[1].Region != "eu-west-1" {
		t.Fatalf("expected region eu-west-1, got %q", resources[1].Region)
	}
}

func TestAWSAdapter_List_All(t *testing.T) {
	account := "123456789012"
	stsOut := &sts.GetCallerIdentityOutput{Account: &account}
	ec2Out := &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{
		Instances: []ec2types.Instance{{InstanceId: ptr("i-1"), State: &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}}},
	}}}
	s3Out := &s3.ListBucketsOutput{Buckets: []s3types.Bucket{{Name: ptr("b-1"), BucketRegion: ptr("us-east-1")}}}

	a := newAdapterWithFakes(stsOut, []*ec2.DescribeInstancesOutput{ec2Out}, s3Out)

	resources, err := a.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 1 EC2 + 1 S3 = 2 resources, got %d", len(resources))
	}
}

func TestAWSAdapter_List_UnknownType(t *testing.T) {
	stsOut := &sts.GetCallerIdentityOutput{Account: ptr("123")}
	a := newAdapterWithFakes(stsOut, nil, nil)
	resources, err := a.List(context.Background(), "rds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected empty result for unknown type, got %d", len(resources))
	}
}

func TestAWSAdapter_Scan(t *testing.T) {
	stsOut := &sts.GetCallerIdentityOutput{Account: ptr("123")}
	ec2Out := &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{
		Instances: []ec2types.Instance{
			{InstanceId: ptr("i-1"), State: &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}},
			{InstanceId: ptr("i-2"), State: &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}},
		},
	}}}
	s3Out := &s3.ListBucketsOutput{Buckets: []s3types.Bucket{{Name: ptr("b-1")}, {Name: ptr("b-2")}, {Name: ptr("b-3")}}}

	a := newAdapterWithFakes(stsOut, []*ec2.DescribeInstancesOutput{ec2Out}, s3Out)

	summary, err := a.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.ProviderName != "aws" || summary.Total != 5 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.CountByType["EC2"] != 2 || summary.CountByType["S3"] != 3 {
		t.Fatalf("unexpected counts: %+v", summary.CountByType)
	}
}

func TestAWSAdapter_List_EC2_Error(t *testing.T) {
	stsOut := &sts.GetCallerIdentityOutput{Account: ptr("123")}
	a := NewAdapter(logger.New("info")).(*adapter)
	a.stsClientFunc = func(ctx context.Context) (stsAPI, error) {
		return &fakeSTSClient{output: stsOut}, nil
	}
	a.ec2ClientFunc = func(ctx context.Context) (ec2API, string, error) {
		return &fakeEC2Client{err: errors.New("api throttled")}, "us-east-1", nil
	}

	_, err := a.List(context.Background(), "ec2")
	if err == nil {
		t.Fatal("expected error from EC2 listing failure")
	}
}
