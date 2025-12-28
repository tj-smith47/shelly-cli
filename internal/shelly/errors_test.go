package shelly

import (
	"errors"
	"strings"
	"testing"
)

const (
	testPlatformEsphome = "esphome"
)

func TestErrCommandNotSupported(t *testing.T) {
	t.Parallel()
	if ErrCommandNotSupported == nil {
		t.Error("ErrCommandNotSupported should not be nil")
	}
	if ErrCommandNotSupported.Error() != "command not supported for this platform" {
		t.Errorf("ErrCommandNotSupported.Error() = %q", ErrCommandNotSupported.Error())
	}
}

func TestErrPluginNotFound(t *testing.T) {
	t.Parallel()
	if ErrPluginNotFound == nil {
		t.Error("ErrPluginNotFound should not be nil")
	}
	if ErrPluginNotFound.Error() != "no plugin found for device platform" {
		t.Errorf("ErrPluginNotFound.Error() = %q", ErrPluginNotFound.Error())
	}
}

func TestErrPluginHookMissing(t *testing.T) {
	t.Parallel()
	if ErrPluginHookMissing == nil {
		t.Error("ErrPluginHookMissing should not be nil")
	}
	if ErrPluginHookMissing.Error() != "plugin does not implement required hook" {
		t.Errorf("ErrPluginHookMissing.Error() = %q", ErrPluginHookMissing.Error())
	}
}

func TestPlatformError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *PlatformError
		contains []string
	}{
		{
			name: "without hint",
			err: &PlatformError{
				Platform: "tasmota",
				Command:  "switch on",
			},
			contains: []string{"switch on", "tasmota", "not supported"},
		},
		{
			name: "with hint",
			err: &PlatformError{
				Platform: "tasmota",
				Command:  "switch on",
				Hint:     "Use the web interface instead",
			},
			contains: []string{"switch on", "tasmota", "Hint:", "web interface"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := tt.err.Error()
			for _, s := range tt.contains {
				if !strings.Contains(msg, s) {
					t.Errorf("Error() = %q, should contain %q", msg, s)
				}
			}
		})
	}
}

func TestNewPlatformError(t *testing.T) {
	t.Parallel()

	err := NewPlatformError(testPlatformEsphome, "reboot")

	if err.Platform != testPlatformEsphome {
		t.Errorf("Platform = %q, want %s", err.Platform, testPlatformEsphome)
	}
	if err.Command != "reboot" {
		t.Errorf("Command = %q, want reboot", err.Command)
	}
	if err.Hint != "" {
		t.Errorf("Hint = %q, want empty", err.Hint)
	}
}

func TestNewPlatformErrorWithHint(t *testing.T) {
	t.Parallel()

	err := NewPlatformErrorWithHint(testPlatformEsphome, "reboot", "Use ESPHome dashboard")

	if err.Platform != testPlatformEsphome {
		t.Errorf("Platform = %q, want %s", err.Platform, testPlatformEsphome)
	}
	if err.Command != "reboot" {
		t.Errorf("Command = %q, want reboot", err.Command)
	}
	if err.Hint != "Use ESPHome dashboard" {
		t.Errorf("Hint = %q, want 'Use ESPHome dashboard'", err.Hint)
	}
}

func TestPluginNotFoundError_Error(t *testing.T) {
	t.Parallel()

	err := &PluginNotFoundError{
		Platform:   "tasmota",
		PluginName: "shelly-tasmota",
	}

	msg := err.Error()
	if !strings.Contains(msg, "tasmota") {
		t.Error("Error() should contain platform name")
	}
	if !strings.Contains(msg, "shelly-tasmota") {
		t.Error("Error() should contain plugin name")
	}
	if !strings.Contains(msg, "shelly plugin install") {
		t.Error("Error() should contain install hint")
	}
}

func TestPluginNotFoundError_Is(t *testing.T) {
	t.Parallel()

	err := &PluginNotFoundError{
		Platform:   "tasmota",
		PluginName: "shelly-tasmota",
	}

	if !errors.Is(err, ErrPluginNotFound) {
		t.Error("PluginNotFoundError should match ErrPluginNotFound")
	}

	if errors.Is(err, ErrPluginHookMissing) {
		t.Error("PluginNotFoundError should not match ErrPluginHookMissing")
	}
}

func TestNewPluginNotFoundError(t *testing.T) {
	t.Parallel()

	err := NewPluginNotFoundError("esphome")

	if err.Platform != "esphome" {
		t.Errorf("Platform = %q, want esphome", err.Platform)
	}
	if err.PluginName != "shelly-esphome" {
		t.Errorf("PluginName = %q, want shelly-esphome", err.PluginName)
	}
}

func TestPluginHookMissingError_Error(t *testing.T) {
	t.Parallel()

	err := &PluginHookMissingError{
		PluginName: "shelly-tasmota",
		Hook:       "SwitchOn",
	}

	msg := err.Error()
	if !strings.Contains(msg, "shelly-tasmota") {
		t.Error("Error() should contain plugin name")
	}
	if !strings.Contains(msg, "SwitchOn") {
		t.Error("Error() should contain hook name")
	}
}

func TestPluginHookMissingError_Is(t *testing.T) {
	t.Parallel()

	err := &PluginHookMissingError{
		PluginName: "shelly-tasmota",
		Hook:       "SwitchOn",
	}

	if !errors.Is(err, ErrPluginHookMissing) {
		t.Error("PluginHookMissingError should match ErrPluginHookMissing")
	}

	if errors.Is(err, ErrPluginNotFound) {
		t.Error("PluginHookMissingError should not match ErrPluginNotFound")
	}
}

func TestNewPluginHookMissingError(t *testing.T) {
	t.Parallel()

	err := NewPluginHookMissingError("my-plugin", "GetStatus")

	if err.PluginName != "my-plugin" {
		t.Errorf("PluginName = %q, want my-plugin", err.PluginName)
	}
	if err.Hook != "GetStatus" {
		t.Errorf("Hook = %q, want GetStatus", err.Hook)
	}
}
