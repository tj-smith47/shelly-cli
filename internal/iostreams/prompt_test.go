package iostreams_test

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const (
	option1 = "option1"
	option2 = "option2"
	option3 = "option3"
)

// Note: Interactive prompt functions (Confirm, Input, Select, etc.) cannot be
// easily unit tested without mocking survey.AskOne. These tests focus on
// non-interactive fallback behavior and the IOStreams methods.

func TestSelectDevice_EmptyDevices(t *testing.T) {
	t.Parallel()

	_, err := iostreams.SelectDevice("Select a device", []string{})
	if err == nil {
		t.Error("SelectDevice() with empty devices should return error")
	}
	if err.Error() != "no devices available" {
		t.Errorf("SelectDevice() error = %q, want %q", err.Error(), "no devices available")
	}
}

func TestSelectDevices_EmptyDevices(t *testing.T) {
	t.Parallel()

	_, err := iostreams.SelectDevices("Select devices", []string{})
	if err == nil {
		t.Error("SelectDevices() with empty devices should return error")
	}
	if err.Error() != "no devices available" {
		t.Errorf("SelectDevices() error = %q, want %q", err.Error(), "no devices available")
	}
}

func TestQuestion_Struct(t *testing.T) {
	t.Parallel()

	// Verify Question struct can be created
	q := iostreams.Question{
		Name: "test",
	}
	if q.Name != "test" {
		t.Errorf("Question.Name = %q, want %q", q.Name, "test")
	}
}

// IOStreams method tests - these test the non-interactive fallback behavior

func TestIOStreams_Confirm_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	// Test() creates non-TTY iostreams, so CanPrompt() returns false

	// Should return default value for non-TTY
	result, err := ios.Confirm("Proceed?", true)
	if err != nil {
		t.Errorf("Confirm() error = %v, want nil", err)
	}
	if !result {
		t.Error("Confirm() should return default value (true) for non-TTY")
	}

	result, err = ios.Confirm("Proceed?", false)
	if err != nil {
		t.Errorf("Confirm() error = %v, want nil", err)
	}
	if result {
		t.Error("Confirm() should return default value (false) for non-TTY")
	}
}

func TestIOStreams_ConfirmDanger_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Should return false for non-TTY (safe default for dangerous operations)
	result, err := ios.ConfirmDanger("Delete everything?")
	if err != nil {
		t.Errorf("ConfirmDanger() error = %v, want nil", err)
	}
	if result {
		t.Error("ConfirmDanger() should return false for non-TTY")
	}
}

func TestIOStreams_Input_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Should return default value for non-TTY
	result, err := ios.Input("Enter name:", "default-name")
	if err != nil {
		t.Errorf("Input() error = %v, want nil", err)
	}
	if result != "default-name" {
		t.Errorf("Input() = %q, want %q", result, "default-name")
	}
}

func TestIOStreams_Select_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	options := []string{option1, option2, option3}

	// Should return default option for non-TTY
	result, err := ios.Select("Choose:", options, 1)
	if err != nil {
		t.Errorf("Select() error = %v, want nil", err)
	}
	if result != option2 {
		t.Errorf("Select() = %q, want %q", result, option2)
	}

	// With invalid default index, should return first option
	result, err = ios.Select("Choose:", options, -1)
	if err != nil {
		t.Errorf("Select() error = %v, want nil", err)
	}
	if result != option1 {
		t.Errorf("Select() with invalid index = %q, want %q", result, option1)
	}

	// With out of range default index, should return first option
	result, err = ios.Select("Choose:", options, 10)
	if err != nil {
		t.Errorf("Select() error = %v, want nil", err)
	}
	if result != option1 {
		t.Errorf("Select() with out of range index = %q, want %q", result, option1)
	}
}

func TestIOStreams_Select_EmptyOptions(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Should return error for empty options
	_, err := ios.Select("Choose:", []string{}, 0)
	if err == nil {
		t.Error("Select() with empty options should return error")
	}
	if err.Error() != "no options available" {
		t.Errorf("Select() error = %q, want %q", err.Error(), "no options available")
	}
}

func TestIOStreams_MultiSelect_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	options := []string{option1, option2, option3}
	defaults := []string{option1, option3}

	// Should return defaults for non-TTY
	result, err := ios.MultiSelect("Choose:", options, defaults)
	if err != nil {
		t.Errorf("MultiSelect() error = %v, want nil", err)
	}
	if len(result) != 2 {
		t.Errorf("MultiSelect() len = %d, want 2", len(result))
	}
	if result[0] != option1 || result[1] != option3 {
		t.Errorf("MultiSelect() = %v, want %v", result, defaults)
	}
}

func TestIOStreams_MultiSelect_NilDefaults(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	options := []string{option1, option2, option3}

	// Should return nil for non-TTY with nil defaults
	result, err := ios.MultiSelect("Choose:", options, nil)
	if err != nil {
		t.Errorf("MultiSelect() error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("MultiSelect() = %v, want nil", result)
	}
}

func TestIOStreams_CanPrompt_TTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Test() creates non-TTY iostreams
	if ios.CanPrompt() {
		t.Error("CanPrompt() should be false for non-TTY")
	}

	// Set TTY state and verify
	ios.SetStdinTTY(true)
	ios.SetStdoutTTY(true)
	if !ios.CanPrompt() {
		t.Error("CanPrompt() should be true when stdin and stdout are TTY")
	}

	// Only stdin TTY is not enough
	ios.SetStdinTTY(true)
	ios.SetStdoutTTY(false)
	if ios.CanPrompt() {
		t.Error("CanPrompt() should be false when stdout is not TTY")
	}

	// Only stdout TTY is not enough
	ios.SetStdinTTY(false)
	ios.SetStdoutTTY(true)
	if ios.CanPrompt() {
		t.Error("CanPrompt() should be false when stdin is not TTY")
	}
}
