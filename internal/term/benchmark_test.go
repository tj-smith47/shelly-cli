package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayLatencyStats_BasicStats(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	stats := model.LatencyStats{
		Min:    5 * time.Millisecond,
		Max:    100 * time.Millisecond,
		Avg:    25 * time.Millisecond,
		P50:    20 * time.Millisecond,
		P95:    80 * time.Millisecond,
		P99:    95 * time.Millisecond,
		Errors: 0,
	}
	DisplayLatencyStats(ios, stats)

	output := out.String()
	if !strings.Contains(output, "Min:") {
		t.Error("expected Min label")
	}
	if !strings.Contains(output, "Max:") {
		t.Error("expected Max label")
	}
	if !strings.Contains(output, "Avg:") {
		t.Error("expected Avg label")
	}
	if !strings.Contains(output, "P50:") {
		t.Error("expected P50 label")
	}
	if !strings.Contains(output, "P95:") {
		t.Error("expected P95 label")
	}
	if !strings.Contains(output, "P99:") {
		t.Error("expected P99 label")
	}
	// Should not contain errors since Errors is 0
	if strings.Contains(output, "Errors:") {
		t.Error("should not show Errors when count is 0")
	}
}

func TestDisplayLatencyStats_WithErrors(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	stats := model.LatencyStats{
		Min:    10 * time.Millisecond,
		Max:    500 * time.Millisecond,
		Avg:    50 * time.Millisecond,
		P50:    40 * time.Millisecond,
		P95:    200 * time.Millisecond,
		P99:    400 * time.Millisecond,
		Errors: 5,
	}
	DisplayLatencyStats(ios, stats)

	output := out.String()
	if !strings.Contains(output, "Errors:") {
		t.Error("expected Errors label when count > 0")
	}
	if !strings.Contains(output, "5") {
		t.Error("expected error count")
	}
}

func TestDisplayLatencyStats_MicrosecondValues(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	stats := model.LatencyStats{
		Min: 100 * time.Microsecond,
		Max: 1 * time.Millisecond,
		Avg: 500 * time.Microsecond,
		P50: 400 * time.Microsecond,
		P95: 800 * time.Microsecond,
		P99: 950 * time.Microsecond,
	}
	DisplayLatencyStats(ios, stats)

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}
}
