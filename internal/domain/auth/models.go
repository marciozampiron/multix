package auth

import "time"

// Credentials represents generalized login data for any provider.
type Credentials struct {
	Token     string
	AccessKey string
	SecretKey string
	ExpiresAt time.Time
}

// Session represents an active connection to a provider (kept for legacy/Login compat).
type Session struct {
	Provider  string `json:"provider"`
	AccountID string `json:"account_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"role,omitempty"`
	IsValid   bool   `json:"is_valid"`
}

// ValidationResult represents the outcome of a provider validation check.
type ValidationResult struct {
	Provider  string            `json:"provider"`
	IsValid   bool              `json:"is_valid"`
	AccountID string            `json:"account_id,omitempty"`
	Principal string            `json:"principal,omitempty"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

// Identity represents detailed metadata about the current authenticated principal.
type Identity struct {
	Provider      string         `json:"provider"`
	AccountID     string         `json:"account_id,omitempty"`
	Principal     string         `json:"principal,omitempty"`
	PrincipalType string         `json:"principal_type,omitempty"`
	ProjectID     string         `json:"project_id,omitempty"`
	AuthSource    string         `json:"auth_source,omitempty"`
	Raw           map[string]any `json:"raw,omitempty"`
}
