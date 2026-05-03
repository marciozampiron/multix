// File: internal/domain/cost/focus.go
// Company: Hassan
// Creator: Zamp
// Created: 02/05/2026
// Updated: 02/05/2026
// Purpose: Domain types for FOCUS-aligned billing reports consumed by cost.focus_report.

// Package cost defines provider-agnostic billing types aligned with the
// FinOps Open Cost & Usage Specification (FOCUS) v1.0.
package cost

// FocusBillingRow is a single line item in a FOCUS-shaped billing report.
//
// Field tags follow the FOCUS PascalCase convention so the JSON wire format
// matches consumers that already speak FOCUS (BigQuery exports, FinOps Hub,
// CloudHealth FOCUS export, etc.). The Go field names follow Go's idiomatic
// camelCase and are exported.
type FocusBillingRow struct {
	BilledCost       float64 `json:"BilledCost"`
	BillingPeriod    string  `json:"BillingPeriod"`    // YYYY-MM
	ProviderName     string  `json:"ProviderName"`     // "AWS" | "GCP" | "OCI"
	ChargeType       string  `json:"ChargeType"`       // "Usage" | "Tax" | "Purchase" | "Credit"
	ResourceCategory string  `json:"ResourceCategory"` // "Compute" | "Storage" | "Network" | "Other"
	ServiceName      string  `json:"ServiceName"`      // raw provider service ID (e.g. "AmazonEC2")
	Region           string  `json:"Region"`
}

// ProviderReport bundles per-provider billing rows with explicit support metadata.
// Providers that have no native integration yet return Supported=false plus a
// reason and tracking issue, so agents can distinguish "no spend" from
// "not yet implemented".
type ProviderReport struct {
	Provider      string            `json:"provider"`
	Supported     bool              `json:"supported"`
	Reason        string            `json:"reason,omitempty"`
	TrackedUnder  string            `json:"tracked_under,omitempty"`
	Rows          []FocusBillingRow `json:"rows,omitempty"`
	ProviderTotal float64           `json:"provider_total,omitempty"`
}

// FocusReport is the top-level payload of cost.focus_report.
type FocusReport struct {
	Currency  string           `json:"currency"`
	Period    string           `json:"period"`
	Providers []ProviderReport `json:"providers"`
	GrandTotal float64         `json:"grand_total"`
}
