package create

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const testFalseValue = "false"

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "create <device> [file]" {
		t.Errorf("Use = %q, want 'create <device> [file]'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}

	// Verify example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Aliases(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	aliases := cmd.Aliases
	if len(aliases) < 2 {
		t.Errorf("expected at least 2 aliases, got %d", len(aliases))
	}

	// Check for expected aliases
	hasNew := false
	hasMake := false
	for _, a := range aliases {
		if a == "new" {
			hasNew = true
		}
		if a == "make" {
			hasMake = true
		}
	}
	if !hasNew {
		t.Error("expected 'new' alias")
	}
	if !hasMake {
		t.Error("expected 'make' alias")
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_RequiresDevice(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require at least 1 argument (device)
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should accept 1 arg (device only)
	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	// Should accept 2 args (device and file)
	err = cmd.Args(cmd, []string{"device1", "backup.json"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	// Should reject 3 args
	err = cmd.Args(cmd, []string{"device1", "backup.json", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Flags(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	// Check format flag
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("format flag shorthand = %q, want 'f'", formatFlag.Shorthand)
	}
	if formatFlag.DefValue != "json" {
		t.Errorf("format flag default = %q, want 'json'", formatFlag.DefValue)
	}

	// Check encrypt flag
	encryptFlag := cmd.Flags().Lookup("encrypt")
	if encryptFlag == nil {
		t.Fatal("encrypt flag not found")
	}
	if encryptFlag.Shorthand != "e" {
		t.Errorf("encrypt flag shorthand = %q, want 'e'", encryptFlag.Shorthand)
	}

	// Check skip-scripts flag
	skipScriptsFlag := cmd.Flags().Lookup("skip-scripts")
	if skipScriptsFlag == nil {
		t.Fatal("skip-scripts flag not found")
	}
	if skipScriptsFlag.DefValue != testFalseValue {
		t.Errorf("skip-scripts flag default = %q, want %q", skipScriptsFlag.DefValue, testFalseValue)
	}

	// Check skip-schedules flag
	skipSchedulesFlag := cmd.Flags().Lookup("skip-schedules")
	if skipSchedulesFlag == nil {
		t.Fatal("skip-schedules flag not found")
	}
	if skipSchedulesFlag.DefValue != testFalseValue {
		t.Errorf("skip-schedules flag default = %q, want %q", skipSchedulesFlag.DefValue, testFalseValue)
	}

	// Check skip-webhooks flag
	skipWebhooksFlag := cmd.Flags().Lookup("skip-webhooks")
	if skipWebhooksFlag == nil {
		t.Fatal("skip-webhooks flag not found")
	}
	if skipWebhooksFlag.DefValue != testFalseValue {
		t.Errorf("skip-webhooks flag default = %q, want %q", skipWebhooksFlag.DefValue, testFalseValue)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_FlagParsing(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Test parsing with flags - these won't actually run because no device
	// but they validate flag parsing works
	testCases := []struct {
		name string
		args []string
	}{
		{"format json", []string{"--format", "json", "device"}},
		{"format yaml", []string{"--format", "yaml", "device"}},
		{"format short", []string{"-f", "yaml", "device"}},
		{"encrypt", []string{"--encrypt", "mypassword", "device"}},
		{"encrypt short", []string{"-e", "mypassword", "device"}},
		{"skip-scripts", []string{"--skip-scripts", "device"}},
		{"skip-schedules", []string{"--skip-schedules", "device"}},
		{"skip-webhooks", []string{"--skip-webhooks", "device"}},
		{"all skip flags", []string{"--skip-scripts", "--skip-schedules", "--skip-webhooks", "device"}},
		{"combined flags", []string{"-f", "yaml", "-e", "pass", "--skip-scripts", "device"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags for each test
			formatFlag = ""
			encryptFlag = ""
			skipScriptsFlag = false
			skipSchedulesFlag = false
			skipWebhooksFlag = false

			cmd := NewCommand(f)
			cmd.SetOut(out)
			cmd.SetErr(errOut)
			cmd.SetArgs(tc.args)

			// Parse flags only (don't run command as it requires network)
			err := cmd.ParseFlags(tc.args)
			if err != nil {
				t.Errorf("ParseFlags failed: %v", err)
			}
		})
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Example_Content(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Check that example contains expected usage patterns
	examples := []string{
		"backup create",
		"backup.json",
		"--format yaml",
		"--encrypt",
		"--skip-scripts",
	}

	for _, e := range examples {
		if !containsString(example, e) {
			t.Errorf("expected example to contain %q", e)
		}
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Long_Description(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long

	// Check that long description contains key information
	keywords := []string{
		"backup",
		"device",
		"scripts",
		"schedules",
		"webhooks",
		"encrypt",
	}

	for _, kw := range keywords {
		if !containsString(long, kw) {
			t.Errorf("expected long description to contain %q", kw)
		}
	}
}
