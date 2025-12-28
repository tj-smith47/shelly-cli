// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"testing"
	"time"
)

func TestCalculateLatencyStats_LargeDataset(t *testing.T) {
	t.Parallel()

	// Create a larger dataset to test percentile calculations
	latencies := make([]time.Duration, 100)
	for i := range 100 {
		latencies[i] = time.Duration(i+1) * time.Millisecond
	}

	stats := CalculateLatencyStats(latencies, 5)

	if stats.Min != 1*time.Millisecond {
		t.Errorf("Min = %v, want 1ms", stats.Min)
	}
	if stats.Max != 100*time.Millisecond {
		t.Errorf("Max = %v, want 100ms", stats.Max)
	}
	if stats.Errors != 5 {
		t.Errorf("Errors = %d, want 5", stats.Errors)
	}
	// Average should be around 50.5ms
	if stats.Avg < 50*time.Millisecond || stats.Avg > 51*time.Millisecond {
		t.Errorf("Avg = %v, want ~50.5ms", stats.Avg)
	}
	// P50 should be around 50ms
	if stats.P50 < 49*time.Millisecond || stats.P50 > 52*time.Millisecond {
		t.Errorf("P50 = %v, want ~50ms", stats.P50)
	}
	// P95 should be around 95ms
	if stats.P95 < 94*time.Millisecond || stats.P95 > 96*time.Millisecond {
		t.Errorf("P95 = %v, want ~95ms", stats.P95)
	}
	// P99 should be around 99ms
	if stats.P99 < 98*time.Millisecond || stats.P99 > 100*time.Millisecond {
		t.Errorf("P99 = %v, want ~99ms", stats.P99)
	}
}

func TestCalculateLatencyStats_UnsortedInput(t *testing.T) {
	t.Parallel()

	// Test with unsorted input to verify sorting works
	latencies := []time.Duration{
		300 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}

	stats := CalculateLatencyStats(latencies, 0)

	if stats.Min != 100*time.Millisecond {
		t.Errorf("Min = %v, want 100ms", stats.Min)
	}
	if stats.Max != 500*time.Millisecond {
		t.Errorf("Max = %v, want 500ms", stats.Max)
	}
}

func TestPercentile_P0(t *testing.T) {
	t.Parallel()

	sorted := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	got := Percentile(sorted, 0)
	if got != 10*time.Millisecond {
		t.Errorf("Percentile(sorted, 0) = %v, want 10ms", got)
	}
}

func TestPercentile_P100(t *testing.T) {
	t.Parallel()

	sorted := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	got := Percentile(sorted, 100)
	// 100% of 3 = 3, which is >= len, so it should return the last element
	if got != 30*time.Millisecond {
		t.Errorf("Percentile(sorted, 100) = %v, want 30ms", got)
	}
}

func TestPercentile_LargePercentile(t *testing.T) {
	t.Parallel()

	sorted := []time.Duration{10 * time.Millisecond}
	// 99% of 1 = 0, so should return first (and only) element
	got := Percentile(sorted, 99)
	if got != 10*time.Millisecond {
		t.Errorf("Percentile(sorted, 99) = %v, want 10ms", got)
	}
}
