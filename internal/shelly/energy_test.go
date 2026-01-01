package shelly

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

func TestCalculateTimeRange_Periods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		period string
	}{
		{"empty period defaults to day", ""},
		{"period hour", "hour"},
		{"period day", "day"},
		{"period week", "week"},
		{"period month", "month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			startTS, endTS, err := CalculateTimeRange(tt.period, "", "")
			if err != nil {
				t.Errorf("CalculateTimeRange() unexpected error: %v", err)
				return
			}
			if startTS == nil || endTS == nil {
				t.Errorf("CalculateTimeRange() expected non-nil timestamps")
				return
			}
			if *startTS >= *endTS {
				t.Errorf("CalculateTimeRange() startTS >= endTS: %v >= %v", *startTS, *endTS)
			}
		})
	}
}

func TestCalculateTimeRange_InvalidPeriod(t *testing.T) {
	t.Parallel()

	_, _, err := CalculateTimeRange("invalid", "", "")
	if err == nil {
		t.Errorf("CalculateTimeRange() expected error for invalid period, got nil")
	}
	if !containsString(err.Error(), "invalid period") {
		t.Errorf("CalculateTimeRange() error = %v, want error containing 'invalid period'", err)
	}
}

func TestCalculateTimeRange_ExplicitTimes(t *testing.T) {
	t.Parallel()

	t.Run("from only", func(t *testing.T) {
		t.Parallel()
		startTS, endTS, err := CalculateTimeRange("", "2025-01-01", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if startTS == nil {
			t.Errorf("expected non-nil startTS")
		}
		if endTS != nil {
			t.Errorf("expected nil endTS, got %v", *endTS)
		}
	})

	t.Run("to only", func(t *testing.T) {
		t.Parallel()
		startTS, endTS, err := CalculateTimeRange("", "", "2025-01-07")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if startTS != nil {
			t.Errorf("expected nil startTS, got %v", *startTS)
		}
		if endTS == nil {
			t.Errorf("expected non-nil endTS")
		}
	})

	t.Run("from and to", func(t *testing.T) {
		t.Parallel()
		startTS, endTS, err := CalculateTimeRange("", "2025-01-01", "2025-01-07")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if startTS == nil || endTS == nil {
			t.Errorf("expected non-nil timestamps")
		}
	})
}

func TestCalculateTimeRange_InvalidTimes(t *testing.T) {
	t.Parallel()

	t.Run("invalid from", func(t *testing.T) {
		t.Parallel()
		_, _, err := CalculateTimeRange("", "invalid-date", "")
		if err == nil {
			t.Errorf("expected error for invalid from time")
		}
	})

	t.Run("invalid to", func(t *testing.T) {
		t.Parallel()
		_, _, err := CalculateTimeRange("", "", "invalid-date")
		if err == nil {
			t.Errorf("expected error for invalid to time")
		}
	})
}

func TestCalculateTimeRange_PeriodDurations(t *testing.T) {
	t.Parallel()

	// Test that periods produce approximately correct durations
	tests := []struct {
		period          string
		expectedMinutes int
		tolerance       int // tolerance in minutes
	}{
		{"hour", 60, 1},
		{"day", 24 * 60, 1},
		{"week", 7 * 24 * 60, 1},
		{"month", 30 * 24 * 60, 1},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			t.Parallel()

			startTS, endTS, err := CalculateTimeRange(tt.period, "", "")
			if err != nil {
				t.Fatalf("CalculateTimeRange() error: %v", err)
			}

			duration := time.Duration(*endTS-*startTS) * time.Second
			expectedDuration := time.Duration(tt.expectedMinutes) * time.Minute
			tolerance := time.Duration(tt.tolerance) * time.Minute

			diff := duration - expectedDuration
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("CalculateTimeRange() duration = %v, expected ~%v (tolerance %v)",
					duration, expectedDuration, tolerance)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "RFC3339",
			input: "2025-01-15T10:30:00Z",
			want:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:  "RFC3339 with timezone",
			input: "2025-01-15T10:30:00+05:00",
			want:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.FixedZone("", 5*60*60)),
		},
		{
			name:  "date only",
			input: "2025-01-15",
			want:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "datetime with space",
			input: "2025-01-15 10:30:00",
			want:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "invalid day",
			input:   "2025-01-99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseTime(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTime(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTime(%q) unexpected error: %v", tt.input, err)
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("ParseTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestComponentTypeConstants(t *testing.T) {
	t.Parallel()

	// Verify constants have expected values
	if ComponentTypeAuto != "auto" {
		t.Errorf("ComponentTypeAuto = %q, want %q", ComponentTypeAuto, "auto")
	}
	if ComponentTypeEM != "em" {
		t.Errorf("ComponentTypeEM = %q, want %q", ComponentTypeEM, "em")
	}
	if ComponentTypeEM1 != "em1" {
		t.Errorf("ComponentTypeEM1 = %q, want %q", ComponentTypeEM1, "em1")
	}
}

func TestCalculateEMMetrics(t *testing.T) {
	t.Parallel()

	t.Run("calculates from EM data", func(t *testing.T) {
		t.Parallel()

		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{
				{
					Period: 60,
					Values: []components.EMDataValues{
						{TotalActivePower: 100.0},
						{TotalActivePower: 200.0},
						{TotalActivePower: 150.0},
					},
				},
			},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEMMetrics(data)

		if dataPoints != 3 {
			t.Errorf("dataPoints = %d, want 3", dataPoints)
		}
		if peakPower != 200.0 {
			t.Errorf("peakPower = %f, want 200.0", peakPower)
		}
		if avgPower != 150.0 {
			t.Errorf("avgPower = %f, want 150.0", avgPower)
		}
		if energy < 0.007 || energy > 0.008 {
			t.Errorf("energy = %f, expected ~0.0075 kWh", energy)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEMMetrics(data)

		if dataPoints != 0 || energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty data")
		}
	})
}

func TestCalculateEM1Metrics(t *testing.T) {
	t.Parallel()

	t.Run("calculates from EM1 data", func(t *testing.T) {
		t.Parallel()

		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{
				{
					Period: 60,
					Values: []components.EM1DataValues{
						{ActivePower: 50.0},
						{ActivePower: 100.0},
						{ActivePower: 75.0},
					},
				},
			},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEM1Metrics(data)

		if dataPoints != 3 {
			t.Errorf("dataPoints = %d, want 3", dataPoints)
		}
		if peakPower != 100.0 {
			t.Errorf("peakPower = %f, want 100.0", peakPower)
		}
		if avgPower != 75.0 {
			t.Errorf("avgPower = %f, want 75.0", avgPower)
		}
		// energy should be small positive value
		if energy <= 0 {
			t.Errorf("energy = %f, expected positive value", energy)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEM1Metrics(data)

		if dataPoints != 0 || energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty data")
		}
	})
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	if substr == "" {
		return true
	}
	if s == "" {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
