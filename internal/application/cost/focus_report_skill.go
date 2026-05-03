// File: internal/application/cost/focus_report_skill.go
// Company: Hassan
// Creator: Zamp
// Created: 02/05/2026
// Updated: 02/05/2026
// Purpose: cost.focus_report skill — multi-cloud FOCUS-aligned billing aggregation.

package cost

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"multix/internal/domain/cost"
	"multix/internal/domain/skills"
	"multix/internal/ports/outbound"

	"cloud.google.com/go/bigquery"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"google.golang.org/api/iterator"
)

const (
	periodCurrentMonth = "current_month"
	periodLast30Days   = "last_30_days"
	periodLastMonth    = "last_month"
)

// FocusReportSkill is the v1 implementation of cost.focus_report:
//
//   - AWS: real Cost Explorer GetCostAndUsage call grouped by SERVICE + REGION.
//   - GCP / OCI: structured `supported: false` envelope pointing at the
//     follow-up issues that will wire their billing APIs.
//
// The skill is intentionally explicit about the partial coverage so the
// agent doesn't conflate "no spend" with "not yet implemented".
type FocusReportSkill struct {
	providers outbound.ProviderRegistry
	// awsCostFetchFunc is a testable seam for the Cost Explorer call.
	awsCostFetchFunc func(ctx context.Context, period string) ([]cost.FocusBillingRow, error)
	// gcpBillingFetchFunc is a testable seam for the BigQuery billing export call.
	gcpBillingFetchFunc func(ctx context.Context, period string) ([]cost.FocusBillingRow, error)
}

// NewFocusReportSkill creates the skill wired with the default AWS Cost
// Explorer client.
func NewFocusReportSkill(pr outbound.ProviderRegistry) skills.Skill {
	s := &FocusReportSkill{providers: pr}
	s.awsCostFetchFunc = s.defaultAWSCostFetch
	s.gcpBillingFetchFunc = s.defaultGCPBillingFetch
	return s
}

func (s *FocusReportSkill) Name() string { return "cost.focus_report" }

func (s *FocusReportSkill) Description() string {
	return "Generates a FOCUS-aligned billing report. AWS uses Cost Explorer, GCP uses Cloud Billing BigQuery export, and OCI returns a structured pending-integration envelope."
}

func (s *FocusReportSkill) InputSchema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"providers": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Subset of providers to include. Defaults to [\"aws\",\"gcp\",\"oci\"].",
			},
			"period": map[string]any{
				"type":        "string",
				"enum":        []string{periodCurrentMonth, periodLast30Days, periodLastMonth},
				"description": "Billing window. Defaults to current_month.",
			},
		},
	}
}

func (s *FocusReportSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	providers := extractProviderList(input)
	period := extractPeriod(input)

	report := cost.FocusReport{
		Currency:  "USD",
		Period:    period,
		Providers: make([]cost.ProviderReport, 0, len(providers)),
	}

	for _, name := range providers {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "aws":
			report.Providers = append(report.Providers, s.fetchAWS(ctx, period))
		case "gcp":
			report.Providers = append(report.Providers, s.fetchGCP(ctx, period))
		case "oci":
			report.Providers = append(report.Providers, notSupported("oci",
				"OCI Usage API integration not yet wired",
				"https://github.com/marciozampiron/multix/issues/46#oci"))
		default:
			report.Providers = append(report.Providers, notSupported(name,
				"unknown provider", ""))
		}
	}

	for _, p := range report.Providers {
		report.GrandTotal += p.ProviderTotal
	}

	return report, nil
}

func (s *FocusReportSkill) fetchAWS(ctx context.Context, period string) cost.ProviderReport {
	rows, err := s.awsCostFetchFunc(ctx, period)
	if err != nil {
		return cost.ProviderReport{
			Provider:  "aws",
			Supported: true,
			Reason:    fmt.Sprintf("Cost Explorer call failed: %v", err),
		}
	}
	pr := cost.ProviderReport{
		Provider:  "aws",
		Supported: true,
		Rows:      rows,
	}
	for _, r := range rows {
		pr.ProviderTotal += r.BilledCost
	}
	return pr
}

func (s *FocusReportSkill) fetchGCP(ctx context.Context, period string) cost.ProviderReport {
	if strings.TrimSpace(os.Getenv("MULTIX_GCP_BILLING_DATASET")) == "" {
		return cost.ProviderReport{
			Provider:     "gcp",
			Supported:    true,
			Reason:       "BigQuery export not configured; set MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>",
			TrackedUnder: "https://github.com/marciozampiron/multix/issues/46#gcp",
		}
	}

	rows, err := s.gcpBillingFetchFunc(ctx, period)
	if err != nil {
		return cost.ProviderReport{
			Provider:  "gcp",
			Supported: true,
			Reason:    fmt.Sprintf("BigQuery billing export query failed: %v", err),
		}
	}

	pr := cost.ProviderReport{
		Provider:  "gcp",
		Supported: true,
		Rows:      rows,
	}
	for _, r := range rows {
		pr.ProviderTotal += r.BilledCost
	}
	return pr
}

func notSupported(provider, reason, trackedUnder string) cost.ProviderReport {
	return cost.ProviderReport{
		Provider:     provider,
		Supported:    false,
		Reason:       reason,
		TrackedUnder: trackedUnder,
	}
}

func (s *FocusReportSkill) defaultAWSCostFetch(ctx context.Context, period string) ([]cost.FocusBillingRow, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := costexplorer.NewFromConfig(cfg)

	start, end, err := resolvePeriod(period, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	out, err := client.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: cetypes.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")},
			{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("REGION")},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Cost Explorer GetCostAndUsage failed; ensure ce:GetCostAndUsage is allowed: %w", err)
	}

	return mapCostExplorerToFocus(out), nil
}

func (s *FocusReportSkill) defaultGCPBillingFetch(ctx context.Context, period string) ([]cost.FocusBillingRow, error) {
	tableID := strings.TrimSpace(os.Getenv("MULTIX_GCP_BILLING_DATASET"))
	projectID, tableRef, err := parseBigQueryTable(tableID)
	if err != nil {
		return nil, err
	}

	start, end, err := resolvePeriod(period, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	defer client.Close()

	queryText := fmt.Sprintf(
		"SELECT\n"+
			"  service.description AS service,\n"+
			"  location.region AS region,\n"+
			"  SUM(cost) AS billed,\n"+
			"  currency,\n"+
			"  invoice.month AS billing_period\n"+
			"FROM `%s`\n"+
			"WHERE _PARTITIONTIME BETWEEN @start AND @end\n"+
			"GROUP BY service, region, currency, billing_period\n",
		tableRef,
	)
	query := client.Query(queryText)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
	}

	iter, err := query.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("BigQuery read failed; ensure roles/bigquery.dataViewer is granted on the billing export dataset: %w", err)
	}

	return mapGCPBillingRowsToFocus(iter)
}

func mapCostExplorerToFocus(out *costexplorer.GetCostAndUsageOutput) []cost.FocusBillingRow {
	if out == nil {
		return nil
	}
	var rows []cost.FocusBillingRow
	for _, period := range out.ResultsByTime {
		billingPeriod := ""
		if period.TimePeriod != nil && period.TimePeriod.Start != nil && len(*period.TimePeriod.Start) >= 7 {
			billingPeriod = (*period.TimePeriod.Start)[:7]
		}
		for _, g := range period.Groups {
			if len(g.Keys) < 2 {
				continue
			}
			service, region := g.Keys[0], g.Keys[1]
			amount := 0.0
			if metric, ok := g.Metrics["UnblendedCost"]; ok && metric.Amount != nil {
				if v, err := strconv.ParseFloat(*metric.Amount, 64); err == nil {
					amount = v
				}
			}
			if amount == 0 {
				continue
			}
			rows = append(rows, cost.FocusBillingRow{
				BilledCost:       amount,
				BillingPeriod:    billingPeriod,
				ProviderName:     "AWS",
				ChargeType:       "Usage",
				ResourceCategory: categorizeAWSService(service),
				ServiceName:      service,
				Region:           region,
			})
		}
	}
	return rows
}

type gcpBillingExportRow struct {
	Service       bigquery.NullString  `bigquery:"service"`
	Region        bigquery.NullString  `bigquery:"region"`
	Billed        bigquery.NullFloat64 `bigquery:"billed"`
	Currency      bigquery.NullString  `bigquery:"currency"`
	BillingPeriod bigquery.NullString  `bigquery:"billing_period"`
}

type gcpRowIterator interface {
	Next(dst any) error
}

func mapGCPBillingRowsToFocus(iter gcpRowIterator) ([]cost.FocusBillingRow, error) {
	var rows []cost.FocusBillingRow
	for {
		var row gcpBillingExportRow
		err := iter.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if !row.Billed.Valid || row.Billed.Float64 == 0 {
			continue
		}
		service := nullString(row.Service)
		rows = append(rows, cost.FocusBillingRow{
			BilledCost:       row.Billed.Float64,
			BillingPeriod:    normalizeGCPBillingPeriod(nullString(row.BillingPeriod)),
			ProviderName:     "GCP",
			ChargeType:       "Usage",
			ResourceCategory: categorizeGCPService(service),
			ServiceName:      service,
			Region:           nullString(row.Region),
		})
	}
	return rows, nil
}

// categorizeAWSService maps the most common AWS service identifiers to
// FOCUS resource categories. Unknown services fall through as "Other".
func categorizeAWSService(service string) string {
	s := strings.ToLower(service)
	switch {
	case strings.Contains(s, "ec2") || strings.Contains(s, "elastic compute") ||
		strings.Contains(s, "lambda") || strings.Contains(s, "fargate") ||
		strings.Contains(s, "ecs") || strings.Contains(s, "eks"):
		return "Compute"
	case strings.Contains(s, "s3") || strings.Contains(s, "ebs") ||
		strings.Contains(s, "efs") || strings.Contains(s, "glacier") ||
		strings.Contains(s, "storage"):
		return "Storage"
	case strings.Contains(s, "vpc") || strings.Contains(s, "cloudfront") ||
		strings.Contains(s, "route 53") || strings.Contains(s, "data transfer"):
		return "Network"
	default:
		return "Other"
	}
}

// categorizeGCPService maps common Google Cloud billing services to FOCUS
// resource categories. Unknown services fall through as "Other".
func categorizeGCPService(service string) string {
	s := strings.ToLower(service)
	switch {
	case strings.Contains(s, "compute engine") || strings.Contains(s, "cloud run") ||
		strings.Contains(s, "kubernetes engine") || strings.Contains(s, "app engine") ||
		strings.Contains(s, "cloud functions"):
		return "Compute"
	case strings.Contains(s, "cloud storage") || strings.Contains(s, "persistent disk") ||
		strings.Contains(s, "filestore"):
		return "Storage"
	case strings.Contains(s, "network") || strings.Contains(s, "load balancing") ||
		strings.Contains(s, "cloud cdn") || strings.Contains(s, "cloud dns"):
		return "Network"
	case strings.Contains(s, "cloud sql") || strings.Contains(s, "spanner") ||
		strings.Contains(s, "bigquery") || strings.Contains(s, "firestore"):
		return "Database"
	default:
		return "Other"
	}
}

var bigQueryIdentifierRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func parseBigQueryTable(value string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("MULTIX_GCP_BILLING_DATASET must use <project>.<dataset>.<table>")
	}
	for _, part := range parts {
		if !bigQueryIdentifierRE.MatchString(part) {
			return "", "", fmt.Errorf("invalid BigQuery billing export identifier %q", value)
		}
	}
	return parts[0], strings.Join(parts, "."), nil
}

func normalizeGCPBillingPeriod(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 6 && strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
		return value[:4] + "-" + value[4:]
	}
	return value
}

func nullString(value bigquery.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.StringVal
}

// resolvePeriod converts the input period token into start/end times in the
// half-open interval Cost Explorer expects: [start, end).
func resolvePeriod(period string, now time.Time) (time.Time, time.Time, error) {
	year, month, _ := now.Date()
	switch period {
	case periodCurrentMonth, "":
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := now.AddDate(0, 0, 1) // up to tomorrow so today's data is included
		return start, end, nil
	case periodLast30Days:
		end := now.AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -30)
		return start, end, nil
	case periodLastMonth:
		startOfThisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := startOfThisMonth
		start := end.AddDate(0, -1, 0)
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported period %q", period)
	}
}

func extractProviderList(input map[string]any) []string {
	defaults := []string{"aws", "gcp", "oci"}
	raw, ok := input["providers"]
	if !ok || raw == nil {
		return defaults
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return defaults
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return defaults
	}
	return out
}

func extractPeriod(input map[string]any) string {
	if v, ok := input["period"].(string); ok && v != "" {
		return v
	}
	return periodCurrentMonth
}
