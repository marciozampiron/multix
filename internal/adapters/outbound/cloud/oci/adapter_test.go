// File: internal/adapters/outbound/cloud/oci/adapter_test.go
package oci

import (
	"context"
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"multix/internal/platform/logger"
)

type mockConfigProvider struct {
	tenancyID string
	userID    string
	err       error
}

func (m *mockConfigProvider) PrivateRSAKey() (key *rsa.PrivateKey, err error) { return nil, nil }
func (m *mockConfigProvider) KeyID() (keyID string, err error)                { return "", nil }
func (m *mockConfigProvider) TenancyOCID() (tenancy string, err error) {
	return m.tenancyID, m.err
}
func (m *mockConfigProvider) UserOCID() (user string, err error) {
	return m.userID, m.err
}
func (m *mockConfigProvider) KeyFingerprint() (fingerprint string, err error) { return "", nil }
func (m *mockConfigProvider) Region() (region string, err error)              { return "us-ashburn-1", nil }
func (m *mockConfigProvider) AuthType() (common.AuthConfig, error)            { return common.AuthConfig{}, nil }

type mockIdentityClient struct {
	err error
}

func (m *mockIdentityClient) GetUser(ctx context.Context, request identity.GetUserRequest) (response identity.GetUserResponse, err error) {
	if m.err != nil {
		return identity.GetUserResponse{}, m.err
	}
	return identity.GetUserResponse{}, nil
}

func TestWhoami(t *testing.T) {
	log := logger.New("error")
	a := &adapter{
		logger: log,
		cfgProviderFunc: func() common.ConfigurationProvider {
			return &mockConfigProvider{
				tenancyID: "ocid1.tenancy.oc1..abc",
				userID:    "ocid1.user.oc1..xyz",
			}
		},
	}

	id, err := a.Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id.AccountID != "ocid1.tenancy.oc1..abc" {
		t.Errorf("got account ID %q, want ocid1.tenancy.oc1..abc", id.AccountID)
	}
	if id.Principal != "ocid1.user.oc1..xyz" {
		t.Errorf("got principal %q, want ocid1.user.oc1..xyz", id.Principal)
	}
}

func TestValidate_Success(t *testing.T) {
	log := logger.New("error")
	a := &adapter{
		logger: log,
		cfgProviderFunc: func() common.ConfigurationProvider {
			return &mockConfigProvider{
				tenancyID: "ocid1.tenancy.oc1..abc",
				userID:    "ocid1.user.oc1..xyz",
			}
		},
		identityClientFunc: func(cfg common.ConfigurationProvider) (identityAPI, error) {
			return &mockIdentityClient{}, nil
		},
	}

	res, err := a.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Valid {
		t.Errorf("expected validation to be true")
	}
	if res.AccountID != "ocid1.tenancy.oc1..abc" {
		t.Errorf("expected account to match tenancy ocid")
	}
}

func TestValidate_APIError(t *testing.T) {
	log := logger.New("error")
	a := &adapter{
		logger: log,
		cfgProviderFunc: func() common.ConfigurationProvider {
			return &mockConfigProvider{
				tenancyID: "ocid1.tenancy.oc1..abc",
				userID:    "ocid1.user.oc1..xyz",
			}
		},
		identityClientFunc: func(cfg common.ConfigurationProvider) (identityAPI, error) {
			return &mockIdentityClient{err: fmt.Errorf("api rejected")}, nil
		},
	}

	_, err := a.Validate(context.Background())
	if err == nil {
		t.Fatalf("expected error from API failure")
	}
}
