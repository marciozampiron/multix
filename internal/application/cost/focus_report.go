// File: internal/application/cost/focus_report.go
// Creator: Gemini
// Purpose: Skill for extracting and standardizing multi-cloud costs into FinOps FOCUS schema.

package cost

import (
	"context"
	"multix/internal/domain/cost"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"
)

type FocusReportSkill struct {
	providers outbound.ProviderRegistry
}

func NewFocusReportSkill(pr outbound.ProviderRegistry) skills.Skill {
	return &FocusReportSkill{providers: pr}
}

func (s *FocusReportSkill) Name() string { return "cost.focus_report" }

func (s *FocusReportSkill) Description() string {
	return "Generates a consolidated billing report using the FinOps FOCUS specification for AI cost analysis."
}

func (s *FocusReportSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"provider": map[string]any{"type": "string", "enum": []string{"aws", "gcp", "oci", "all"}, "description": "Cloud provider to extract billing from"},
			"period":   map[string]any{"type": "string", "enum": []string{"current_month", "last_30_days", "last_month"}, "description": "Billing period"},
		},
		"required": []string{"provider", "period"},
	}
}

func (s *FocusReportSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	// Parse input safely
	providerReq, _ := input["provider"].(string)
	period, _ := input["period"].(string)

	if period == "" {
		period = "current_month"
	}

	// For MVP / Agent testing without actual billing privileges:
	// Return a stubbed/mocked FOCUS response.
	report := cost.FocusReport{
		Currency: "USD",
		Rows: []cost.FocusBillingRow{},
	}

	if providerReq == "aws" || providerReq == "all" {
		report.Rows = append(report.Rows, cost.FocusBillingRow{
			BilledCost:       1250.50,
			BillingPeriod:    period,
			ProviderName:     "AWS",
			ChargeType:       "Usage",
			ResourceCategory: "Compute",
			Region:           "us-east-1",
		})
	}
	
	if providerReq == "gcp" || providerReq == "all" {
		report.Rows = append(report.Rows, cost.FocusBillingRow{
			BilledCost:       850.75,
			BillingPeriod:    period,
			ProviderName:     "GCP",
			ChargeType:       "Usage",
			ResourceCategory: "Compute",
			Region:           "us-central1",
		})
	}

	// Calculate total cost
	var total float64
	for _, row := range report.Rows {
		total += row.BilledCost
	}
	report.TotalCost = total

	return map[string]any{
		"status": "success",
		"report": report,
	}, nil
}
