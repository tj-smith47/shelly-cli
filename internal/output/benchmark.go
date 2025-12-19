// Package output provides output formatting utilities for the CLI.
package output

import "github.com/tj-smith47/shelly-cli/internal/model"

// FormatBenchmarkSummary returns a human-readable summary based on latency stats.
func FormatBenchmarkSummary(stats model.LatencyStats) string {
	avgMs := float64(stats.Avg.Microseconds()) / 1000

	switch {
	case avgMs < 50:
		return "Excellent - Very fast response times"
	case avgMs < 100:
		return "Good - Normal response times"
	case avgMs < 200:
		return "Fair - Slightly elevated latency"
	case avgMs < 500:
		return "Poor - High latency, check network"
	default:
		return "Critical - Very high latency, network issues likely"
	}
}
