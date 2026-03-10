package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	analyticsSourceID string
	analyticsStart    string
	analyticsEnd      string
)

var analyticsMetrics = map[string][]string{
	"bandwidth": {"bandwidth_consumption"},
	"requests":  {"requests_count"},
	"summary":   {"bandwidth_consumption", "requests_count", "avg_response_time", "transformations_count"},
}

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.AddCommand(analyticsBandwidthCmd)
	analyticsCmd.AddCommand(analyticsRequestsCmd)
	analyticsCmd.AddCommand(analyticsSummaryCmd)

	for _, cmd := range []*cobra.Command{analyticsBandwidthCmd, analyticsRequestsCmd, analyticsSummaryCmd} {
		cmd.Flags().StringVar(&analyticsSourceID, "source-id", "", "Source ID to query (required)")
		cmd.Flags().StringVar(&analyticsStart, "start", "", "Start date (YYYY-MM-DD, default: yesterday)")
		cmd.Flags().StringVar(&analyticsEnd, "end", "", "End date (YYYY-MM-DD, default: today)")
		cmd.MarkFlagRequired("source-id")
	}
}

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Query image analytics and usage",
}

var analyticsBandwidthCmd = &cobra.Command{
	Use:   "bandwidth",
	Short: "Bandwidth usage for a source",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalytics(cmd, "bandwidth")
	},
}

var analyticsRequestsCmd = &cobra.Command{
	Use:   "requests",
	Short: "Request count metrics for a source",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalytics(cmd, "requests")
	},
}

var analyticsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Combined analytics overview for a source",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalytics(cmd, "summary")
	},
}

func runAnalytics(cmd *cobra.Command, analyticsType string) error {
	c, err := getClient(cmd)
	if err != nil {
		return err
	}

	start, end := resolveAnalyticsDates(analyticsStart, analyticsEnd)
	metrics := analyticsMetrics[analyticsType]

	body := map[string]any{
		"source_id": analyticsSourceID,
		"metrics":   metrics,
		"date_range": map[string]string{
			"start_at": start,
			"end_at":   end,
		},
	}

	data, err := c.Post(context.Background(), "v1/image/analytics", body)
	if err != nil {
		return err
	}

	// Transform API response into table rows.
	// API returns: {"metric_name": [{"units": N, "timestamp": T}, ...], ...}
	// We need:     [{"date": "...", "metric_name": N, ...}, ...]
	rows, err := flattenAnalyticsResponse(data, metrics)
	if err != nil {
		return err
	}

	tableCmd := fmt.Sprintf("analytics.%s", analyticsType)
	return printData(tableCmd, rows)
}

type analyticsPoint struct {
	Units     float64 `json:"units"`
	Timestamp int64   `json:"timestamp"`
}

func flattenAnalyticsResponse(data json.RawMessage, metrics []string) (json.RawMessage, error) {
	var raw map[string][]analyticsPoint
	if err := json.Unmarshal(data, &raw); err != nil {
		return data, nil // fallback: return as-is
	}

	// Collect all timestamps from the first metric
	var firstMetric []analyticsPoint
	for _, m := range metrics {
		if pts, ok := raw[m]; ok {
			firstMetric = pts
			break
		}
	}
	if len(firstMetric) == 0 {
		return data, nil
	}

	rows := make([]map[string]any, len(firstMetric))
	for i, pt := range firstMetric {
		t := time.Unix(pt.Timestamp, 0)
		row := map[string]any{
			"date": t.Format("2006-01-02"),
		}
		for _, m := range metrics {
			if pts, ok := raw[m]; ok && i < len(pts) {
				row[m] = formatMetricValue(m, pts[i].Units)
			}
		}
		rows[i] = row
	}

	return json.Marshal(rows)
}

func formatMetricValue(metric string, value float64) any {
	switch metric {
	case "bandwidth_consumption":
		return humanBytes(value)
	case "avg_response_time":
		return fmt.Sprintf("%.0fms", value*1000)
	case "status_2xx", "status_4xx", "status_5xx":
		return fmt.Sprintf("%.1f%%", value*100)
	default:
		if value == float64(int64(value)) {
			return int64(value)
		}
		return value
	}
}

func humanBytes(b float64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.2f GB", b/gb)
	case b >= mb:
		return fmt.Sprintf("%.2f MB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%.2f KB", b/kb)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

func resolveAnalyticsDates(start, end string) (string, string) {
	now := time.Now()
	if end == "" {
		end = now.Format("2006-01-02")
	}
	if start == "" {
		start = now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return start, end
}
