// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"sort"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// CalculateLatencyStats computes statistics from a slice of latency measurements.
func CalculateLatencyStats(latencies []time.Duration, errors int) model.LatencyStats {
	if len(latencies) == 0 {
		return model.LatencyStats{Errors: errors}
	}

	// Sort for percentiles
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate statistics
	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}

	return model.LatencyStats{
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Avg:    sum / time.Duration(len(sorted)),
		P50:    Percentile(sorted, 50),
		P95:    Percentile(sorted, 95),
		P99:    Percentile(sorted, 99),
		Errors: errors,
	}
}

// Percentile returns the p-th percentile value from a sorted slice of durations.
func Percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
