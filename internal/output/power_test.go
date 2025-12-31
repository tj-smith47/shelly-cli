package output

import (
	"strings"
	"testing"
)

func TestFormatPowerValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		watts float64
		opts  PowerValueOptions
		want  string
	}{
		{
			name:  "basic formatting small value",
			watts: 50,
			opts:  PowerValueOptions{},
			want:  "50.0 W",
		},
		{
			name:  "basic formatting kW",
			watts: 1500,
			opts:  PowerValueOptions{},
			want:  "1.50 kW",
		},
		{
			name:  "zero value",
			watts: 0,
			opts:  PowerValueOptions{},
			want:  "0.0 W",
		},
		{
			name:  "zero as placeholder",
			watts: 0,
			opts:  PowerValueOptions{ZeroAsPlaceholder: true},
			want:  "-",
		},
		{
			name:  "negative as placeholder",
			watts: -10,
			opts:  PowerValueOptions{ZeroAsPlaceholder: true},
			want:  "-",
		},
		{
			name:  "custom placeholder",
			watts: 0,
			opts:  PowerValueOptions{ZeroAsPlaceholder: true, Placeholder: "N/A"},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatPowerValue(tt.watts, tt.opts)
			if got != tt.want {
				t.Errorf("FormatPowerValue(%v, %+v) = %q, want %q", tt.watts, tt.opts, got, tt.want)
			}
		})
	}
}

func TestFormatPowerValueColored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		watts   float64
		wantSub string // Substring that must be present
	}{
		{
			name:    "colored low power contains value",
			watts:   50,
			wantSub: "50.0 W",
		},
		{
			name:    "colored medium power contains value",
			watts:   500,
			wantSub: "500.0 W",
		},
		{
			name:    "colored high power contains value",
			watts:   1500,
			wantSub: "1.50 kW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatPowerValue(tt.watts, PowerValueOptions{Colored: true})
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("FormatPowerValue(%v, Colored) = %q, want to contain %q", tt.watts, got, tt.wantSub)
			}
			// Colored output should contain ANSI escape sequences
			if !strings.Contains(got, "\x1b[") {
				t.Errorf("FormatPowerValue(%v, Colored) = %q, expected ANSI escape sequences", tt.watts, got)
			}
		})
	}
}

func TestFormatPowerValueWithChange(t *testing.T) {
	t.Parallel()

	prev := 100.0

	t.Run("increase shows up arrow", func(t *testing.T) {
		t.Parallel()
		got := FormatPowerValue(150, PowerValueOptions{
			Colored:    true,
			ShowChange: true,
			PrevPower:  &prev,
		})
		if !strings.Contains(got, "150.0 W") {
			t.Errorf("FormatPowerValue(150) should contain '150.0 W', got %q", got)
		}
		if !strings.Contains(got, "↑") {
			t.Errorf("FormatPowerValue(150) with prev=100 should show up arrow, got %q", got)
		}
	})

	t.Run("decrease shows down arrow", func(t *testing.T) {
		t.Parallel()
		got := FormatPowerValue(50, PowerValueOptions{
			Colored:    true,
			ShowChange: true,
			PrevPower:  &prev,
		})
		if !strings.Contains(got, "50.0 W") {
			t.Errorf("FormatPowerValue(50) should contain '50.0 W', got %q", got)
		}
		if !strings.Contains(got, "↓") {
			t.Errorf("FormatPowerValue(50) with prev=100 should show down arrow, got %q", got)
		}
	})

	t.Run("no change has no arrow", func(t *testing.T) {
		t.Parallel()
		got := FormatPowerValue(100, PowerValueOptions{
			Colored:    true,
			ShowChange: true,
			PrevPower:  &prev,
		})
		if !strings.Contains(got, "100.0 W") {
			t.Errorf("FormatPowerValue(100) should contain '100.0 W', got %q", got)
		}
		if strings.Contains(got, "↑") || strings.Contains(got, "↓") {
			t.Errorf("FormatPowerValue(100) with prev=100 should not have arrow, got %q", got)
		}
	})

	t.Run("nil previous with colored applies color", func(t *testing.T) {
		t.Parallel()
		got := FormatPowerValue(100, PowerValueOptions{
			Colored:    true,
			ShowChange: true,
			PrevPower:  nil,
		})
		if !strings.Contains(got, "100.0 W") {
			t.Errorf("FormatPowerValue(100) should contain '100.0 W', got %q", got)
		}
		// Should have ANSI escape sequences for coloring
		if !strings.Contains(got, "\x1b[") {
			t.Errorf("FormatPowerValue(100) with Colored=true should have ANSI escapes, got %q", got)
		}
	})
}

func TestFormatMeterReading(t *testing.T) {
	t.Parallel()

	id := 1
	pf := 0.95

	tests := []struct {
		name string
		opts MeterLineOptions
		want string
	}{
		{
			name: "basic meter line",
			opts: MeterLineOptions{
				Label:   "PM",
				ID:      &id,
				Power:   100,
				Voltage: 230,
				Current: 0.43,
			},
			want: "  PM 1: 100.0 W  230.0V  0.43A",
		},
		{
			name: "meter line without ID",
			opts: MeterLineOptions{
				Label:   "Test",
				Power:   100,
				Voltage: 120,
				Current: 0.83,
			},
			want: "  Test: 100.0 W  120.0V  0.83A",
		},
		{
			name: "meter line with PF",
			opts: MeterLineOptions{
				Label:   "EM1",
				ID:      &id,
				Power:   500,
				Voltage: 230,
				Current: 2.17,
				PF:      &pf,
			},
			want: "  EM1 1: 500.0 W  230.0V  2.17A  PF:0.95",
		},
		{
			name: "meter line with energy",
			opts: MeterLineOptions{
				Label:   "PM",
				ID:      &id,
				Power:   100,
				Voltage: 230,
				Current: 0.43,
				Energy:  floatPtr(1234.56),
			},
			want: "  PM 1: 100.0 W  230.0V  0.43A  1234.56 Wh",
		},
		{
			name: "custom indent",
			opts: MeterLineOptions{
				Label:   "PM",
				ID:      &id,
				Power:   100,
				Voltage: 230,
				Current: 0.43,
				Indent:  "    ",
			},
			want: "    PM 1: 100.0 W  230.0V  0.43A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatMeterReading(tt.opts)
			if got != tt.want {
				t.Errorf("FormatMeterReading() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPhaseLine(t *testing.T) {
	t.Parallel()

	pf := 0.98

	tests := []struct {
		name string
		opts PhaseLineOptions
		want string
	}{
		{
			name: "phase A",
			opts: PhaseLineOptions{
				Phase:   "A",
				Power:   500,
				Voltage: 230,
				Current: 2.17,
			},
			want: "    Phase A: 500.0 W  230.0V  2.17A",
		},
		{
			name: "phase with PF",
			opts: PhaseLineOptions{
				Phase:   "B",
				Power:   600,
				Voltage: 232,
				Current: 2.59,
				PF:      &pf,
			},
			want: "    Phase B: 600.0 W  232.0V  2.59A  PF:0.98",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatPhaseLine(tt.opts)
			if got != tt.want {
				t.Errorf("FormatPhaseLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

// floatPtr is a helper to create a pointer to a float64.
func floatPtr(f float64) *float64 {
	return &f
}
