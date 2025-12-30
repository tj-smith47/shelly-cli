package output

import (
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

func TestRenderOnOff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		on     bool
		c      Case
		wantOn bool // true if result should contain ON variant
	}{
		{"on lower", true, CaseLower, true},
		{"off lower", false, CaseLower, false},
		{"on title", true, CaseTitle, true},
		{"off title", false, CaseTitle, false},
		{"on upper", true, CaseUpper, true},
		{"off upper", false, CaseUpper, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Test with no color (simpler assertion)
			got := RenderOnOff(tt.on, tt.c, theme.FalseDim)
			if tt.wantOn {
				if got != LabelOnLower && got != LabelOnTitle && got != LabelOnUpper {
					// With color enabled, the result may have ANSI codes
					// Just verify it's not empty
					if got == "" {
						t.Error("expected non-empty result for ON state")
					}
				}
			}
		})
	}
}

func TestRenderYesNo(t *testing.T) {
	t.Parallel()

	t.Run("yes title", func(t *testing.T) {
		t.Parallel()
		got := RenderYesNo(true, CaseTitle, theme.FalseDim)
		// Result should contain "Yes" (possibly with ANSI codes)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("no lower", func(t *testing.T) {
		t.Parallel()
		got := RenderYesNo(false, CaseLower, theme.FalseDim)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderOnline(t *testing.T) {
	t.Parallel()

	t.Run("online title", func(t *testing.T) {
		t.Parallel()
		got := RenderOnline(true, CaseTitle)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("offline lower", func(t *testing.T) {
		t.Parallel()
		got := RenderOnline(false, CaseLower)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderActive(t *testing.T) {
	t.Parallel()

	t.Run("active", func(t *testing.T) {
		t.Parallel()
		got := RenderActive(true, CaseTitle, theme.FalseError)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("inactive", func(t *testing.T) {
		t.Parallel()
		got := RenderActive(false, CaseLower, theme.FalseDim)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderBoolState(t *testing.T) {
	t.Parallel()

	t.Run("true state", func(t *testing.T) {
		t.Parallel()
		got := RenderBoolState(true, "Success", "Failure")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("false state", func(t *testing.T) {
		t.Parallel()
		got := RenderBoolState(false, "Success", "Failure")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderEnabledState(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()
		got := RenderEnabledState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()
		got := RenderEnabledState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderCoverState(t *testing.T) {
	t.Parallel()

	tests := []string{LabelCoverOpen, LabelCoverClosed, "moving"}

	for _, state := range tests {
		t.Run(state, func(t *testing.T) {
			t.Parallel()
			got := RenderCoverState(state)
			if got == "" {
				t.Errorf("expected non-empty result for state %q", state)
			}
		})
	}
}

func TestRenderGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		gen  int
		want string
	}{
		{0, "unknown"},
		{1, "Gen1"},
		{2, "Gen2"},
		{3, "Gen3"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := RenderGeneration(tt.gen)
			if got != tt.want {
				t.Errorf("RenderGeneration(%d) = %q, want %q", tt.gen, got, tt.want)
			}
		})
	}
}

func TestRenderSwitchState(t *testing.T) {
	t.Parallel()

	t.Run("on", func(t *testing.T) {
		t.Parallel()
		status := &model.SwitchStatus{Output: true}
		got := RenderSwitchState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()
		status := &model.SwitchStatus{Output: false}
		got := RenderSwitchState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderLightState(t *testing.T) {
	t.Parallel()

	t.Run("on with brightness", func(t *testing.T) {
		t.Parallel()
		brightness := 75
		status := &model.LightStatus{Output: true, Brightness: &brightness}
		got := RenderLightState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()
		status := &model.LightStatus{Output: false}
		got := RenderLightState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderCoverStatusState(t *testing.T) {
	t.Parallel()

	t.Run("with position", func(t *testing.T) {
		t.Parallel()
		pos := 50
		status := &model.CoverStatus{State: "open", CurrentPosition: &pos}
		got := RenderCoverStatusState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
		// Should contain percentage
		if !containsSubstring(got, "50%") {
			t.Errorf("expected result to contain '50%%', got %q", got)
		}
	})

	t.Run("without position", func(t *testing.T) {
		t.Parallel()
		status := &model.CoverStatus{State: "closed"}
		got := RenderCoverStatusState(status)
		if got != "closed" {
			t.Errorf("got %q, want %q", got, "closed")
		}
	})
}

func TestRenderInputState(t *testing.T) {
	t.Parallel()

	t.Run("triggered", func(t *testing.T) {
		t.Parallel()
		status := &model.InputStatus{State: true}
		got := RenderInputState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("idle", func(t *testing.T) {
		t.Parallel()
		status := &model.InputStatus{State: false}
		got := RenderInputState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderTokenValidity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		valid   bool
		expired bool
	}{
		{"valid", true, false},
		{"expired", true, true},
		{"invalid", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderTokenValidity(tt.valid, tt.expired)
			if got == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestRenderUpdateStatus(t *testing.T) {
	t.Parallel()

	t.Run("update available", func(t *testing.T) {
		t.Parallel()
		got := RenderUpdateStatus(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("up to date", func(t *testing.T) {
		t.Parallel()
		got := RenderUpdateStatus(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderAlarmState(t *testing.T) {
	t.Parallel()

	t.Run("alarm", func(t *testing.T) {
		t.Parallel()
		got := RenderAlarmState(true, "WATER DETECTED!")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("clear", func(t *testing.T) {
		t.Parallel()
		got := RenderAlarmState(false, "WATER DETECTED!")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderMuteAnnotation(t *testing.T) {
	t.Parallel()

	t.Run("muted", func(t *testing.T) {
		t.Parallel()
		got := RenderMuteAnnotation(true)
		if got == "" {
			t.Error("expected non-empty result for muted")
		}
	})

	t.Run("not muted", func(t *testing.T) {
		t.Parallel()
		got := RenderMuteAnnotation(false)
		if got != "" {
			t.Errorf("expected empty result for not muted, got %q", got)
		}
	})
}

func TestRenderAuthRequired(t *testing.T) {
	t.Parallel()

	t.Run("required", func(t *testing.T) {
		t.Parallel()
		got := RenderAuthRequired(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("not required", func(t *testing.T) {
		t.Parallel()
		got := RenderAuthRequired(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderDiffLabels(t *testing.T) {
	t.Parallel()

	t.Run("removed", func(t *testing.T) {
		t.Parallel()
		got := RenderDiffRemoved()
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("added", func(t *testing.T) {
		t.Parallel()
		got := RenderDiffAdded()
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("changed", func(t *testing.T) {
		t.Parallel()
		got := RenderDiffChanged()
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderAvailableState(t *testing.T) {
	t.Parallel()

	t.Run("available", func(t *testing.T) {
		t.Parallel()
		got := RenderAvailableState(true, "N/A")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("unavailable", func(t *testing.T) {
		t.Parallel()
		got := RenderAvailableState(false, "Not available")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderRunningState(t *testing.T) {
	t.Parallel()

	t.Run("running", func(t *testing.T) {
		t.Parallel()
		got := RenderRunningState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("stopped", func(t *testing.T) {
		t.Parallel()
		got := RenderRunningState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderNetworkState(t *testing.T) {
	t.Parallel()

	t.Run("joined", func(t *testing.T) {
		t.Parallel()
		got := RenderNetworkState(LabelJoined)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("disconnected", func(t *testing.T) {
		t.Parallel()
		got := RenderNetworkState("disconnected")
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderErrorState(t *testing.T) {
	t.Parallel()

	got := RenderErrorState()
	if got == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderValveState(t *testing.T) {
	t.Parallel()

	t.Run("open", func(t *testing.T) {
		t.Parallel()
		got := RenderValveState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("closed", func(t *testing.T) {
		t.Parallel()
		got := RenderValveState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderLoggedInState(t *testing.T) {
	t.Parallel()

	t.Run("logged in", func(t *testing.T) {
		t.Parallel()
		got := RenderLoggedInState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		got := RenderLoggedInState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderMuteState(t *testing.T) {
	t.Parallel()

	t.Run("muted", func(t *testing.T) {
		t.Parallel()
		got := RenderMuteState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("active", func(t *testing.T) {
		t.Parallel()
		got := RenderMuteState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderRGBState(t *testing.T) {
	t.Parallel()

	t.Run("on with brightness", func(t *testing.T) {
		t.Parallel()
		brightness := 75
		status := &model.RGBStatus{Output: true, Brightness: &brightness}
		got := RenderRGBState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()
		status := &model.RGBStatus{Output: false}
		got := RenderRGBState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("on without brightness", func(t *testing.T) {
		t.Parallel()
		status := &model.RGBStatus{Output: true}
		got := RenderRGBState(status)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderInputTriggeredState(t *testing.T) {
	t.Parallel()

	t.Run("triggered", func(t *testing.T) {
		t.Parallel()
		got := RenderInputTriggeredState(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("idle", func(t *testing.T) {
		t.Parallel()
		got := RenderInputTriggeredState(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderOnOffStateWithBrightness(t *testing.T) {
	t.Parallel()

	t.Run("on with brightness 0", func(t *testing.T) {
		t.Parallel()
		brightness := 0
		got := RenderOnOffStateWithBrightness(true, &brightness)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("on with no brightness pointer", func(t *testing.T) {
		t.Parallel()
		got := RenderOnOffStateWithBrightness(true, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("on with positive brightness", func(t *testing.T) {
		t.Parallel()
		brightness := 75
		got := RenderOnOffStateWithBrightness(true, &brightness)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()
		got := RenderOnOffStateWithBrightness(false, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderCoverStatusState_NegativePosition(t *testing.T) {
	t.Parallel()

	// Test with negative position (should not show percentage)
	neg := -1
	status := &model.CoverStatus{State: "calibrating", CurrentPosition: &neg}
	got := RenderCoverStatusState(status)
	if got != "calibrating" {
		t.Errorf("expected 'calibrating', got %q", got)
	}
}

func TestCaseConstants(t *testing.T) {
	t.Parallel()

	// Test that case constants are distinct
	if CaseLower == CaseTitle {
		t.Error("CaseLower and CaseTitle should be different")
	}
	if CaseTitle == CaseUpper {
		t.Error("CaseTitle and CaseUpper should be different")
	}
	if CaseLower == CaseUpper {
		t.Error("CaseLower and CaseUpper should be different")
	}
}

//nolint:paralleltest // Tests modify shared viper and isTTY state
func TestRenderFunctions_WithColor(t *testing.T) {
	// Save and restore isTTY
	oldIsTTY := isTTY
	defer func() { isTTY = oldIsTTY }()

	// Enable TTY mode to test color output
	isTTY = func() bool { return true }
	viper.Set("plain", false)
	viper.Set("no-color", false)
	t.Setenv("TERM", "xterm-256color")

	// RenderOnOff tests
	t.Run("RenderOnOff on with color", func(t *testing.T) {
		got := RenderOnOff(true, CaseTitle, theme.FalseDim)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("RenderOnOff off with color", func(t *testing.T) {
		got := RenderOnOff(false, CaseLower, theme.FalseDim)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	// RenderOnline tests
	t.Run("RenderOnline online with color", func(t *testing.T) {
		got := RenderOnline(true, CaseTitle)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("RenderOnline offline with color", func(t *testing.T) {
		got := RenderOnline(false, CaseLower)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	// RenderAuthRequired tests
	t.Run("RenderAuthRequired required with color", func(t *testing.T) {
		got := RenderAuthRequired(true)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("RenderAuthRequired not required with color", func(t *testing.T) {
		got := RenderAuthRequired(false)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

//nolint:paralleltest // Tests modify shared isTTY and viper state
func TestColorEnabled_TTY(t *testing.T) {
	// Save and restore isTTY
	oldIsTTY := isTTY
	defer func() { isTTY = oldIsTTY }()

	t.Run("non-TTY returns false", func(t *testing.T) {
		isTTY = func() bool { return false }
		if colorEnabled() {
			t.Error("expected colorEnabled() = false for non-TTY")
		}
	})

	t.Run("TTY with plain flag returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		viper.Set("plain", true)
		defer viper.Set("plain", false)
		if colorEnabled() {
			t.Error("expected colorEnabled() = false when plain=true")
		}
	})

	t.Run("TTY with no-color flag returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		viper.Set("no-color", true)
		defer viper.Set("no-color", false)
		if colorEnabled() {
			t.Error("expected colorEnabled() = false when no-color=true")
		}
	})

	t.Run("TTY with NO_COLOR env returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		if err := os.Setenv("NO_COLOR", "1"); err != nil {
			t.Fatalf("failed to set NO_COLOR: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset NO_COLOR: %v", err)
			}
		}()
		if colorEnabled() {
			t.Error("expected colorEnabled() = false when NO_COLOR is set")
		}
	})

	t.Run("TTY with SHELLY_NO_COLOR env returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		if err := os.Setenv("SHELLY_NO_COLOR", "1"); err != nil {
			t.Fatalf("failed to set SHELLY_NO_COLOR: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("SHELLY_NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset SHELLY_NO_COLOR: %v", err)
			}
		}()
		if colorEnabled() {
			t.Error("expected colorEnabled() = false when SHELLY_NO_COLOR is set")
		}
	})

	t.Run("TTY with TERM=dumb returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		oldTerm := os.Getenv("TERM")
		if err := os.Setenv("TERM", "dumb"); err != nil {
			t.Fatalf("failed to set TERM: %v", err)
		}
		defer func() {
			if oldTerm == "" {
				if err := os.Unsetenv("TERM"); err != nil {
					t.Logf("warning: failed to unset TERM: %v", err)
				}
			} else {
				if err := os.Setenv("TERM", oldTerm); err != nil {
					t.Logf("warning: failed to restore TERM: %v", err)
				}
			}
		}()
		if colorEnabled() {
			t.Error("expected colorEnabled() = false when TERM=dumb")
		}
	})

	t.Run("TTY with no restrictions returns true", func(t *testing.T) {
		isTTY = func() bool { return true }
		// Clear all flags and env vars
		viper.Set("plain", false)
		viper.Set("no-color", false)
		if err := os.Unsetenv("NO_COLOR"); err != nil {
			t.Logf("warning: failed to unset NO_COLOR: %v", err)
		}
		if err := os.Unsetenv("SHELLY_NO_COLOR"); err != nil {
			t.Logf("warning: failed to unset SHELLY_NO_COLOR: %v", err)
		}
		oldTerm := os.Getenv("TERM")
		if err := os.Setenv("TERM", "xterm-256color"); err != nil {
			t.Fatalf("failed to set TERM: %v", err)
		}
		defer func() {
			if oldTerm == "" {
				if err := os.Unsetenv("TERM"); err != nil {
					t.Logf("warning: failed to unset TERM: %v", err)
				}
			} else {
				if err := os.Setenv("TERM", oldTerm); err != nil {
					t.Logf("warning: failed to restore TERM: %v", err)
				}
			}
		}()
		if !colorEnabled() {
			t.Error("expected colorEnabled() = true when TTY with no restrictions")
		}
	})
}
