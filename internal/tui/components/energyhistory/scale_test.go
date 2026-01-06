package energyhistory

import (
	"testing"
	"time"
)

func TestScaleDataToWidth(t *testing.T) {
	t.Parallel()
	now := time.Now()
	makeHistory := func(n int) []DataPoint {
		history := make([]DataPoint, n)
		for i := range n {
			history[i] = DataPoint{
				Value:     float64(i * 10),
				Timestamp: now.Add(time.Duration(i) * time.Second),
			}
		}
		return history
	}

	tests := []struct {
		name    string
		histLen int
		width   int
		wantLen int
	}{
		{"equal", 60, 60, 60},
		{"scale down", 120, 60, 60},
		{"scale up (typical)", 36, 80, 80},  // 3 minutes of data, 80 char sparkline
		{"scale up (max data)", 60, 80, 80}, // 5 minutes of data, 80 char sparkline
		{"small scale up", 10, 20, 20},
		{"single point", 1, 10, 10},
		{"empty", 0, 10, 0},
		{"width 1", 5, 1, 1},
		{"width 2", 5, 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			history := makeHistory(tt.histLen)
			result := scaleDataToWidth(history, tt.width)

			if len(result) != tt.wantLen {
				t.Errorf("scaleDataToWidth(%d history, %d width) = %d points, want %d",
					tt.histLen, tt.width, len(result), tt.wantLen)
			}

			// For non-empty results, verify values are in expected range
			if len(result) > 0 && tt.histLen > 0 {
				minVal := float64(0)
				maxVal := float64((tt.histLen - 1) * 10)
				for i, dp := range result {
					if dp.Value < minVal || dp.Value > maxVal {
						t.Errorf("point %d value %f out of range [%f, %f]",
							i, dp.Value, minVal, maxVal)
					}
				}
			}
		})
	}
}

func TestScaleDataToWidthFillsFullWidth(t *testing.T) {
	t.Parallel()
	// Simulate 36 data points (3 minutes) being displayed in 80-char sparkline
	now := time.Now()
	history := make([]DataPoint, 36)
	for i := range 36 {
		history[i] = DataPoint{
			Value:     float64(i%10) * 10, // 0-90 cycling
			Timestamp: now.Add(time.Duration(i) * 5 * time.Second),
		}
	}

	width := 80
	result := scaleDataToWidth(history, width)

	if len(result) != width {
		t.Fatalf("Expected %d points, got %d", width, len(result))
	}

	// The result should span the full width with interpolated values
	// Not just 36 points padded with zeros
	nonZeroCount := 0
	for _, dp := range result {
		if dp.Value > 0 {
			nonZeroCount++
		}
	}

	// Should have reasonable distribution of non-zero values throughout
	if nonZeroCount < width/2 {
		t.Errorf("Only %d non-zero values in %d points - data not properly distributed",
			nonZeroCount, width)
	}
}
