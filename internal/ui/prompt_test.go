package ui

import (
	"testing"
)

// Note: Most prompt functions use survey which requires a terminal.
// These tests verify the function signatures and basic error handling.
// Full interactive testing would require terminal simulation.

func TestSelectDevice_EmptyList(t *testing.T) {
	t.Parallel()

	_, err := SelectDevice("Select a device", []string{})

	if err == nil {
		t.Error("SelectDevice() should return error for empty list")
	}
	if err.Error() != "no devices available" {
		t.Errorf("SelectDevice() error = %q, want %q", err.Error(), "no devices available")
	}
}

// Note: The following functions require an interactive terminal to test fully:
// - Confirm
// - ConfirmDanger
// - Input
// - InputRequired
// - Password
// - Select
// - MultiSelect
// - Credential
//
// They use AlecAivazis/survey which checks for terminal capabilities.
// Integration tests would need to use go-expect or similar libraries.
