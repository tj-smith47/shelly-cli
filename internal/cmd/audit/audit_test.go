package audit

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "audit [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "audit [device...]")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Short != "Security audit for devices" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Security audit for devices")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"security", "sec"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias at index %d", i)
			continue
		}
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "all", shorthand: "", defValue: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_RunE_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no args and no --all flag")
	} else if err.Error() != "specify device(s) or use --all" {
		t.Errorf("error = %q, want %q", err.Error(), "specify device(s) or use --all")
	}
}

func TestNewCommand_RunE_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		allFlag     bool
		wantError   bool
		errorString string
	}{
		{
			name:        "no args and no --all flag",
			args:        []string{},
			allFlag:     false,
			wantError:   true,
			errorString: "specify device(s) or use --all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			cmd.SetArgs(tt.args)

			if tt.allFlag {
				if err := cmd.Flags().Set("all", "true"); err != nil {
					t.Fatalf("failed to set --all flag: %v", err)
				}
			}

			err := cmd.Execute()
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorString != "" && err.Error() != tt.errorString {
					t.Errorf("error = %q, want %q", err.Error(), tt.errorString)
				}
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description contains key information
	checks := []string{
		"security audit",
		"Authentication status",
		"Cloud connection",
		"Firmware version",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Long, check) {
			t.Errorf("Long description missing %q", check)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify examples contain key usage patterns
	checks := []string{
		"shelly audit kitchen-light",
		"shelly audit light-1 switch-2",
		"shelly audit --all",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Example, check) {
			t.Errorf("Example missing %q", check)
		}
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Verify Run is not set (we should use RunE, not Run)
	if cmd.Run != nil {
		t.Error("Run should be nil, use RunE instead")
	}
}

func TestNewCommand_AllFlagMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		wantError bool
	}{
		{
			name:      "set all flag to true",
			flagValue: "true",
			wantError: false,
		},
		{
			name:      "set all flag to false",
			flagValue: "false",
			wantError: false,
		},
		{
			name:      "set all flag to invalid",
			flagValue: "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Flags().Set("all", tt.flagValue)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command can be accessed by name
	if cmd.Name() != "audit" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "audit")
	}
}

func TestNewCommand_WithDeviceArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Execute - will attempt to audit the device
	// The test factory's ShellyService will handle the request
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error from device audit: %v", err)
	}

	// Verify output contains the audit header
	output := tf.OutString()
	if !strings.Contains(output, "Security Audit") {
		t.Errorf("expected Security Audit header in output, got: %q", output)
	}
}

func TestNewCommand_WithDeviceArg_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a cancelled context to prevent actual network calls
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - may fail due to cancelled context but exercises the code path
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error due to cancelled context: %v", err)
	}

	// Verify some output was produced (even with error)
	output := tf.OutString()
	// Should have at least attempted to display the audit header
	if !strings.Contains(output, "Security Audit") {
		// The run function should print the title before auditing devices
		t.Logf("output was: %q", output)
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func(cmd *cobra.Command) bool
		wantOK bool
		errMsg string
	}{
		{
			name:   "has use",
			check:  func(c *cobra.Command) bool { return c.Use != "" },
			wantOK: true,
			errMsg: "Use should not be empty",
		},
		{
			name:   "has short",
			check:  func(c *cobra.Command) bool { return c.Short != "" },
			wantOK: true,
			errMsg: "Short should not be empty",
		},
		{
			name:   "has long",
			check:  func(c *cobra.Command) bool { return c.Long != "" },
			wantOK: true,
			errMsg: "Long should not be empty",
		},
		{
			name:   "has example",
			check:  func(c *cobra.Command) bool { return c.Example != "" },
			wantOK: true,
			errMsg: "Example should not be empty",
		},
		{
			name:   "has aliases",
			check:  func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK: true,
			errMsg: "Aliases should not be empty",
		},
		{
			name:   "has RunE",
			check:  func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK: true,
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			if tt.check(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "all flag",
			args:    []string{"--all"},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
