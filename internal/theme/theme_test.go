package theme

import (
	"strings"
	"testing"
)

func TestListThemes(t *testing.T) {
	t.Parallel()

	themes := ListThemes()

	if len(themes) == 0 {
		t.Error("ListThemes() returned empty list")
	}

	// Should have common themes
	hasTheme := func(name string) bool {
		for _, th := range themes {
			if strings.EqualFold(th, name) {
				return true
			}
		}
		return false
	}

	if !hasTheme("dracula") {
		t.Error("expected 'dracula' theme to be available")
	}
	if !hasTheme("nord") {
		t.Error("expected 'nord' theme to be available")
	}
}

//nolint:paralleltest // Test modifies global theme state
func TestSetTheme(t *testing.T) {
	// Valid theme
	if !SetTheme("nord") {
		t.Error("SetTheme('nord') returned false")
	}

	// Invalid theme
	if SetTheme("nonexistent-theme-12345") {
		t.Error("SetTheme('nonexistent') should return false")
	}

	// Reset to default
	SetTheme("dracula")
}

//nolint:paralleltest // Test modifies global theme state
func TestCurrent(t *testing.T) {
	SetTheme("dracula")
	tint := Current()

	if tint == nil {
		t.Fatal("Current() returned nil")
	}
}

func TestStatusStyles(t *testing.T) {
	t.Parallel()

	// These should not panic
	_ = StatusOK()
	_ = StatusWarn()
	_ = StatusError()
	_ = StatusInfo()
}

func TestDeviceStatus(t *testing.T) {
	t.Parallel()

	online := DeviceOnline()
	if !strings.Contains(online, "online") {
		t.Errorf("DeviceOnline() = %q, expected to contain 'online'", online)
	}

	offline := DeviceOffline()
	if !strings.Contains(offline, "offline") {
		t.Errorf("DeviceOffline() = %q, expected to contain 'offline'", offline)
	}

	updating := DeviceUpdating()
	if !strings.Contains(updating, "updating") {
		t.Errorf("DeviceUpdating() = %q, expected to contain 'updating'", updating)
	}
}

func TestSwitchStates(t *testing.T) {
	t.Parallel()

	on := SwitchOn()
	if on == "" {
		t.Error("SwitchOn() returned empty string")
	}

	off := SwitchOff()
	if off == "" {
		t.Error("SwitchOff() returned empty string")
	}
}

func TestFormatPower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watts    float64
		contains string
	}{
		{0.0, "0"},
		{100.5, "100"},
		{1500.0, "1500"},
	}

	for _, tt := range tests {
		result := FormatPower(tt.watts)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("FormatPower(%v) = %q, expected to contain %q", tt.watts, result, tt.contains)
		}
		if !strings.Contains(result, "W") {
			t.Errorf("FormatPower(%v) = %q, expected to contain 'W'", tt.watts, result)
		}
	}
}

func TestFormatEnergy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wh       float64
		contains string
		unit     string
	}{
		{500.0, "500", "Wh"},   // 500 Wh (< 1000)
		{1000.0, "1", "kWh"},   // 1 kWh
		{2500.0, "2.5", "kWh"}, // 2.5 kWh
	}

	for _, tt := range tests {
		result := FormatEnergy(tt.wh)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("FormatEnergy(%v) = %q, expected to contain %q", tt.wh, result, tt.contains)
		}
		if !strings.Contains(result, tt.unit) {
			t.Errorf("FormatEnergy(%v) = %q, expected to contain %q", tt.wh, result, tt.unit)
		}
	}
}

func TestTextStyles(t *testing.T) {
	t.Parallel()

	// These should not panic and should return valid styles
	_ = Bold()
	_ = Dim()
	_ = Highlight()
	_ = Title()
	_ = Subtitle()
	_ = Link()
	_ = Code()
}

//nolint:paralleltest // Test modifies global theme state
func TestNextPrevTheme(t *testing.T) {
	SetTheme("dracula")
	initial := Current().ID()

	NextTheme()
	after := Current().ID()

	if initial == after {
		t.Error("NextTheme() did not change theme")
	}

	PrevTheme()
	restored := Current().ID()

	if restored != initial {
		t.Error("PrevTheme() did not restore previous theme")
	}
}
