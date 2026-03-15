// File: internal/adapters/outbound/cloud/aws/adapter_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Tests AWS auth adapter normalization behavior without live cloud dependencies.

package aws

import (
	"context"
	"errors"
	"testing"

	"multix/internal/platform/logger"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type fakeSTSClient struct {
	output *sts.GetCallerIdentityOutput
	err    error
}

func (f *fakeSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

func TestAWSAdapterValidateAndWhoami(t *testing.T) {
	log := logger.New("info")
	a := NewAdapter(log).(*adapter)

	account := "123456789012"
	arn := "arn:aws:sts::123456789012:assumed-role/Admin/session"
	userID := "AIDAXXXXX"
	a.stsClientFunc = func(ctx context.Context) (stsAPI, error) {
		return &fakeSTSClient{output: &sts.GetCallerIdentityOutput{Account: &account, Arn: &arn, UserId: &userID}}, nil
	}

	result, err := a.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if !result.Valid || result.AccountID != account || result.Principal != arn {
		t.Fatalf("unexpected validate payload: %+v", result)
	}

	identity, err := a.Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected whoami error: %v", err)
	}
	if identity.PrincipalType != "role" || identity.UserID != userID {
		t.Fatalf("unexpected identity payload: %+v", identity)
	}
}

func TestAWSAdapterValidateErrors(t *testing.T) {
	log := logger.New("info")
	a := NewAdapter(log).(*adapter)
	a.stsClientFunc = func(ctx context.Context) (stsAPI, error) {
		return &fakeSTSClient{err: errors.New("expired token")}, nil
	}

	_, err := a.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error when sts call fails")
	}
}
