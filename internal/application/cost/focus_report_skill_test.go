// File: internal/application/cost/focus_report_skill_test.go
// Purpose: Unit tests for cost.focus_report — AWS happy path, GCP/OCI not_supported envelopes,
// Cost Explorer mapping, period resolution, and Cost Explorer error propagation.

package cost

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"multix/internal/domain/cost"
	"multix/internal/ports/outbound"
)

type stubRegistry struct{}

func (stubRegistry) GetCloudAuthProvider(string) (outbound.AuthProvider, error) {
	return nil, errors.New("unused")
}
func (stubRegistry) GetCloudInventoryProvider(string) (outbound.InventoryProvider, error) {
	return nil, errors.New("unused")
}
func (stubRegistry) GetKubernetesProvider(string) (outbound.K8sProvider, error) {
	return nil, errors.New("unused")
}
func (stubRegistry) GetAIProvider(string) (outbound.AIProvider, error) {
	return nil, errors.New("unused")
}

func newSkillWithFakeAWS(rows []cost.FocusBillingRow, awsErr error) *FocusReportSkill {
	s := NewFocusReportSkill(stubRegistry{}).(*FocusReportSkill)
	s.awsCostFetchFunc = func(ctx context.Context, period string) ([]cost.FocusBillingRow, error) {
		return rows, awsErr
	}
	return s
}

func newSkillWithFakeAWSAndGCP(awsRows, gcpRows []cost.FocusBillingRow, awsErr, gcpErr error) *FocusReportSkill {
	s := newSkillWithFakeAWS(awsRows, awsErr)
	s.gcpBillingFetchFunc = func(ctx context.Context, period string) ([]cost.FocusBillingRow, error) {
		return gcpRows, gcpErr
	}
	return s
}

func TestFocusReport_AWSHappyPath(t *testing.T) {
	rows := []cost.FocusBillingRow{
		{BilledCost: 100, BillingPeriod: "2026-05", ProviderName: "AWS", ChargeType: "Usage", ResourceCategory: "Compute", ServiceName: "AmazonEC2", Region: "us-east-1"},
		{BilledCost: 25, BillingPeriod: "2026-05", ProviderName: "AWS", ChargeType: "Usage", ResourceCategory: "Storage", ServiceName: "AmazonS3", Region: "us-east-1"},
	}
	s := newSkillWithFakeAWS(rows, nil)

	out, err := s.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := out.(cost.FocusReport)
	if report.GrandTotal != 125 {
		t.Fatalf("expected GrandTotal=125, got %v", report.GrandTotal)
	}
	if len(report.Providers) != 1 || !report.Providers[0].Supported || report.Providers[0].ProviderTotal != 125 {
		t.Fatalf("unexpected aws provider report: %+v", report.Providers[0])
	}
}

func TestFocusReport_GCPRequiresBigQueryExportAndOCIReturnsNotSupported(t *testing.T) {
	t.Setenv("MULTIX_GCP_BILLING_DATASET", "")
	s := newSkillWithFakeAWS(nil, nil)

	out, err := s.Execute(context.Background(), map[string]any{"providers": []any{"gcp", "oci"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := out.(cost.FocusReport)
	if len(report.Providers) != 2 {
		t.Fatalf("expected 2 provider reports, got %d", len(report.Providers))
	}
	gcp := report.Providers[0]
	if !gcp.Supported {
		t.Fatalf("gcp should remain supported but unconfigured, got %+v", gcp)
	}
	if !strings.Contains(gcp.Reason, "BigQuery export not configured") {
		t.Fatalf("unexpected gcp configuration reason: %+v", gcp)
	}

	oci := report.Providers[1]
	if oci.Supported {
		t.Fatalf("oci must remain unsupported until OCI Usage API is wired, got %+v", oci)
	}
	if oci.Reason == "" || oci.TrackedUnder == "" {
		t.Fatalf("oci missing reason/tracked_under: %+v", oci)
	}

	if report.GrandTotal != 0 {
		t.Fatalf("GrandTotal should be 0 when only unsupported providers requested, got %v", report.GrandTotal)
	}
}

func TestFocusReport_GCPHappyPath(t *testing.T) {
	t.Setenv("MULTIX_GCP_BILLING_DATASET", "demo-project.billing_export.gcp_billing")
	rows := []cost.FocusBillingRow{
		{BilledCost: 80, BillingPeriod: "2026-05", ProviderName: "GCP", ChargeType: "Usage", ResourceCategory: "Compute", ServiceName: "Compute Engine", Region: "us-central1"},
		{BilledCost: 20, BillingPeriod: "2026-05", ProviderName: "GCP", ChargeType: "Usage", ResourceCategory: "Storage", ServiceName: "Cloud Storage", Region: "us"},
	}
	s := newSkillWithFakeAWSAndGCP(nil, rows, nil, nil)

	out, err := s.Execute(context.Background(), map[string]any{"providers": []any{"gcp"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report := out.(cost.FocusReport)
	if report.GrandTotal != 100 {
		t.Fatalf("expected GrandTotal=100, got %v", report.GrandTotal)
	}
	if len(report.Providers) != 1 || !report.Providers[0].Supported || report.Providers[0].ProviderTotal != 100 {
		t.Fatalf("unexpected gcp provider report: %+v", report.Providers)
	}
}

func TestFocusReport_GCPBigQueryError(t *testing.T) {
	t.Setenv("MULTIX_GCP_BILLING_DATASET", "demo-project.billing_export.gcp_billing")
	s := newSkillWithFakeAWSAndGCP(nil, nil, nil, errors.New("permission denied"))

	out, _ := s.Execute(context.Background(), map[string]any{"providers": []any{"gcp"}})
	report := out.(cost.FocusReport)
	gcp := report.Providers[0]
	if !gcp.Supported {
		t.Fatal("gcp.Supported should remain true on BigQuery API errors")
	}
	if !strings.Contains(gcp.Reason, "permission denied") {
		t.Fatalf("expected BigQuery error in reason, got %+v", gcp)
	}
	if len(gcp.Rows) != 0 {
		t.Fatalf("expected zero rows on error, got %d", len(gcp.Rows))
	}
}

func TestFocusReport_AWSError(t *testing.T) {
	s := newSkillWithFakeAWS(nil, errors.New("AccessDenied: ce:GetCostAndUsage not allowed"))

	out, _ := s.Execute(context.Background(), map[string]any{"providers": []any{"aws"}})
	report := out.(cost.FocusReport)
	aws := report.Providers[0]
	if !aws.Supported {
		t.Fatal("aws.Supported should remain true on transient API errors")
	}
	if aws.Reason == "" {
		t.Fatal("aws report should expose the Cost Explorer error message")
	}
	if len(aws.Rows) != 0 {
		t.Fatalf("expected zero rows on error, got %d", len(aws.Rows))
	}
}

func TestFocusReport_DefaultProvidersAndPeriod(t *testing.T) {
	s := newSkillWithFakeAWS(nil, nil)
	out, _ := s.Execute(context.Background(), map[string]any{})
	report := out.(cost.FocusReport)

	if report.Period != "current_month" {
		t.Errorf("expected default period current_month, got %q", report.Period)
	}
	if len(report.Providers) != 3 {
		t.Errorf("expected default 3 providers, got %d", len(report.Providers))
	}
	got := []string{}
	for _, p := range report.Providers {
		got = append(got, p.Provider)
	}
	want := []string{"aws", "gcp", "oci"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("provider[%d] = %q; want %q", i, got[i], w)
		}
	}
}

func TestFocusReport_UnknownProviderFlagged(t *testing.T) {
	s := newSkillWithFakeAWS(nil, nil)
	out, _ := s.Execute(context.Background(), map[string]any{"providers": []any{"alibaba"}})
	report := out.(cost.FocusReport)
	if len(report.Providers) != 1 || report.Providers[0].Supported || report.Providers[0].Reason != "unknown provider" {
		t.Fatalf("unexpected report for unknown provider: %+v", report.Providers)
	}
}

func TestCategorizeAWSService(t *testing.T) {
	cases := map[string]string{
		"Amazon Elastic Compute Cloud - Compute": "Compute",
		"Amazon Simple Storage Service":          "Storage",
		"Amazon CloudFront":                      "Network",
		"AWS Lambda":                             "Compute",
		"Amazon EBS":                             "Storage",
		"AWS Data Transfer":                      "Network",
		"Amazon Quantum Ledger":                  "Other",
	}
	for input, want := range cases {
		if got := categorizeAWSService(input); got != want {
			t.Errorf("categorizeAWSService(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestCategorizeGCPService(t *testing.T) {
	cases := map[string]string{
		"Compute Engine":       "Compute",
		"Kubernetes Engine":    "Compute",
		"Cloud Storage":        "Storage",
		"Persistent Disk":      "Storage",
		"Cloud Load Balancing": "Network",
		"Cloud DNS":            "Network",
		"Cloud SQL":            "Database",
		"BigQuery":             "Database",
		"Secret Manager":       "Other",
	}
	for input, want := range cases {
		if got := categorizeGCPService(input); got != want {
			t.Errorf("categorizeGCPService(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestParseBigQueryTable(t *testing.T) {
	project, table, err := parseBigQueryTable("demo-project.billing_export.gcp_billing")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if project != "demo-project" || table != "demo-project.billing_export.gcp_billing" {
		t.Fatalf("unexpected parsed table: project=%q table=%q", project, table)
	}

	if _, _, err := parseBigQueryTable("demo-project.billing_export"); err == nil {
		t.Fatal("expected error for incomplete table ID")
	}
	if _, _, err := parseBigQueryTable("demo-project.billing_export.gcp_billing;DROP"); err == nil {
		t.Fatal("expected error for unsafe table ID")
	}
}

func TestNormalizeGCPBillingPeriod(t *testing.T) {
	if got := normalizeGCPBillingPeriod("202605"); got != "2026-05" {
		t.Fatalf("expected 2026-05, got %q", got)
	}
	if got := normalizeGCPBillingPeriod("2026-05"); got != "2026-05" {
		t.Fatalf("expected existing period to pass through, got %q", got)
	}
}

func TestResolvePeriod(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)

	t.Run("current_month", func(t *testing.T) {
		start, end, err := resolvePeriod("current_month", now)
		if err != nil {
			t.Fatal(err)
		}
		if start.Format("2006-01-02") != "2026-05-01" {
			t.Errorf("start = %s", start)
		}
		// end should include today (so >= 2026-05-16)
		if !end.After(now) {
			t.Errorf("end %s should be after now %s", end, now)
		}
	})

	t.Run("last_month", func(t *testing.T) {
		start, end, err := resolvePeriod("last_month", now)
		if err != nil {
			t.Fatal(err)
		}
		if start.Format("2006-01-02") != "2026-04-01" || end.Format("2006-01-02") != "2026-05-01" {
			t.Errorf("expected 2026-04-01..2026-05-01, got %s..%s", start, end)
		}
	})

	t.Run("last_30_days", func(t *testing.T) {
		_, end, err := resolvePeriod("last_30_days", now)
		if err != nil {
			t.Fatal(err)
		}
		if !end.After(now) {
			t.Errorf("end should be in the future relative to now")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		if _, _, err := resolvePeriod("ytd", now); err == nil {
			t.Error("expected error for unsupported period")
		}
	})
}
