// File: internal/domain/cost/focus.go
// Creator: Gemini
// Purpose: Defines the Go structures mapped to the FinOps Open Cost & Usage Specification (FOCUS).

package cost

// FocusBillingRow represents a single standardized line item in a cloud billing report.
type FocusBillingRow struct {
	BilledCost       float64 `json:"BilledCost"`
	BillingPeriod    string  `json:"BillingPeriod"` // e.g. "2026-05"
	ProviderName     string  `json:"ProviderName"`  // "AWS", "GCP", "OCI"
	ChargeType       string  `json:"ChargeType"`    // "Usage", "Tax", "Purchase"
	ResourceCategory string  `json:"ResourceCategory"` // "Compute", "Storage", "Network"
	Region           string  `json:"Region"`
}

// FocusReport is the aggregated payload ready for AI analysis.
type FocusReport struct {
	Rows       []FocusBillingRow `json:"Rows"`
	TotalCost  float64           `json:"TotalCost"`
	Currency   string            `json:"Currency"`
}
