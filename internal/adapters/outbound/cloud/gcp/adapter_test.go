// File: internal/adapters/outbound/cloud/gcp/adapter_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Tests GCP auth adapter best-effort identity normalization without live cloud dependencies.

package gcp

import (
	"context"
	"errors"
	"testing"

	"multix/internal/platform/logger"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func TestGCPAdapterValidateAndWhoami_ServiceAccount(t *testing.T) {
	log := logger.New("info")
	a := NewAdapter(log)
	a.findCredentialsFunc = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return &google.Credentials{
			ProjectID: "demo-project",
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: "token",
			}),
			JSON: []byte(`{"client_email":"bot@demo-project.iam.gserviceaccount.com"}`),
		}, nil
	}

	result, err := a.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if !result.Valid || result.AccountID != "demo-project" {
		t.Fatalf("unexpected validate result: %+v", result)
	}

	identity, err := a.Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected whoami error: %v", err)
	}
	if identity.Principal != "bot@demo-project.iam.gserviceaccount.com" || identity.PrincipalType != "service_account" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
}

func TestGCPAdapterWhoami_BestEffortFallback(t *testing.T) {
	log := logger.New("info")
	a := NewAdapter(log)
	a.findCredentialsFunc = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return &google.Credentials{ProjectID: "demo-project"}, nil
	}

	identity, err := a.Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Principal != "" || identity.Note == "" || identity.AuthSource == "" {
		t.Fatalf("expected best-effort fallback identity fields, got %+v", identity)
	}
}

func TestGCPAdapterValidateErrors(t *testing.T) {
	log := logger.New("info")
	a := NewAdapter(log)
	a.findCredentialsFunc = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return nil, errors.New("no adc")
	}

	_, err := a.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
