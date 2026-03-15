package auth

import "time"

// Credentials represents generalized login data for any provider.
type Credentials struct {
	Token     string
	AccessKey string
	SecretKey string
	ExpiresAt time.Time
}

// Session represents an active validated connection to a provider.
type Session struct {
	AccountID string
	Username  string
	Role      string
	IsValid   bool
	Provider  string
}
