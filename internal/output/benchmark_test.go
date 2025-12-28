package output

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestFormatBenchmarkSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		avgUs    int64 // average in microseconds
		wantText string
	}{
		{
			name:     "excellent < 50ms",
			avgUs:    30000, // 30ms
			wantText: "Excellent",
		},
		{
			name:     "good 50-100ms",
			avgUs:    75000, // 75ms
			wantText: "Good",
		},
		{
			name:     "fair 100-200ms",
			avgUs:    150000, // 150ms
			wantText: "Fair",
		},
		{
			name:     "poor 200-500ms",
			avgUs:    350000, // 350ms
			wantText: "Poor",
		},
		{
			name:     "critical > 500ms",
			avgUs:    700000, // 700ms
			wantText: "Critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := model.LatencyStats{
				Avg: time.Duration(tt.avgUs) * time.Microsecond,
			}
			got := FormatBenchmarkSummary(stats)
			if !containsSubstring(got, tt.wantText) {
				t.Errorf("FormatBenchmarkSummary() = %q, want to contain %q", got, tt.wantText)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
