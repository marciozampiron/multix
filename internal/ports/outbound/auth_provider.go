package outbound

import (
	"context"

	"multix/internal/domain/auth"
)

// AuthProvider defines the capabilities needed to authenticate against a backend.
type AuthProvider interface {
	// ID returns the unique identifier for this provider.
	ID() string

	// Login initiates a connection and returns an active session.
	Login(ctx context.Context, creds auth.Credentials) (*auth.Session, error)

	// Whoami retrieves information about the current active identity.
	Whoami(ctx context.Context) (*auth.Identity, error)

	// Validate checks if the current context credentials are valid.
	Validate(ctx context.Context) (*auth.ValidationResult, error)
}
