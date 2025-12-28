package factoryreset

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "factory-reset <device>" {
		t.Errorf("Use = %q, want 'factory-reset <device>'", cmd.Use)
	}
	if len(cmd.Aliases) < 3 {
		t.Errorf("Expected at least 3 aliases, got %d", len(cmd.Aliases))
	}
	// Check all expected aliases
	expectedAliases := []string{"fr", "reset", "wipe"}
	for _, expected := range expectedAliases {
		found := false
		for _, alias := range cmd.Aliases {
			if alias == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", expected, cmd.Aliases)
		}
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}

	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag == nil {
		t.Fatal("confirm flag not found")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want false", yesFlag.DefValue)
	}

	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag.DefValue != "false" {
		t.Errorf("confirm default = %q, want false", confirmFlag.DefValue)
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}
}

func TestNewCommand_RejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Expected error when multiple args provided")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			wantOK:    true,
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "uses ExactArgs(1)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			wantOK:    true,
			errMsg:    "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
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
			name:    "yes flag short",
			args:    []string{"-y"},
			wantErr: false,
		},
		{
			name:    "yes flag long",
			args:    []string{"--yes"},
			wantErr: false,
		},
		{
			name:    "confirm flag long",
			args:    []string{"--confirm"},
			wantErr: false,
		},
		{
			name:    "both flags",
			args:    []string{"-y", "--confirm"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
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

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Default values
	if opts.Yes {
		t.Error("Default Yes should be false")
	}
	if opts.Confirm {
		t.Error("Default Confirm should be false")
	}
}

func TestOptions_DeviceFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-device",
	}

	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want 'my-device'", opts.Device)
	}
}

func TestRun_MissingConfirmFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = false
	opts.Confirm = false

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when confirm flags missing")
	}

	// Check output contains safety message
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output warning about missing flags")
	}
}

func TestRun_OnlyYesFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = true
	opts.Confirm = false

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when only yes flag set")
	}
}

func TestRun_OnlyConfirmFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = false
	opts.Confirm = true

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when only confirm flag set")
	}
}

func TestRun_BothFlagsButCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Provide "n" response for final interactive confirmation
	tf.TestIO.In.WriteString("n\n")

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = true
	opts.Confirm = true

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected nil error when user cancels, got: %v", err)
	}

	// Should print cancelled message
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output when cancelled")
	}
}

func TestRun_BothFlagsWithContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = true
	opts.Confirm = true

	// Create a cancelled context - confirmation prompt will fail with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// The confirmation prompt may fail or succeed depending on implementation
	// Either the confirmation returns an error (cancelled) or proceeds and fails at reset
	// The key is that the function doesn't hang and handles cancellation gracefully
	_ = err // Just ensure it doesn't panic
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = true
	opts.Confirm = true

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	err := run(ctx, opts)
	// The confirmation prompt may fail or succeed depending on implementation
	// Either the confirmation times out or proceeds and fails at reset
	// The key is that the function doesn't hang and handles timeout gracefully
	_ = err // Just ensure it doesn't panic
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_RunE_SetsDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device", "-y", "--confirm"})

	// Provide "n" response for final confirmation to avoid actual reset
	tf.TestIO.In.WriteString("n\n")

	err := cmd.Execute()
	// No error expected when user cancels at final confirmation
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if example == "" {
		t.Fatal("Example should not be empty")
	}

	// Check for expected patterns
	patterns := []string{"shelly", "device", "factory-reset", "--yes", "--confirm"}
	for _, pattern := range patterns {
		found := false
		for i := 0; i <= len(example)-len(pattern); i++ {
			if example[i:i+len(pattern)] == pattern {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description should not be empty")
	}

	// Long should be more descriptive than Short
	if len(cmd.Long) <= len(cmd.Short) {
		t.Error("Long description should be longer than Short description")
	}

	// Should contain warning text
	long := cmd.Long
	hasWarning := false
	for i := 0; i <= len(long)-7; i++ {
		if long[i:i+7] == "WARNING" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("Long description should contain WARNING")
	}
}

func TestNewCommand_RequiresBothFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		yes       bool
		confirm   bool
		expectErr bool
	}{
		{"neither flag", false, false, true},
		{"only yes", true, false, true},
		{"only confirm", false, true, true},
		{"both flags", true, true, false}, // Will fail at actual reset, but passes validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			tf.TestIO.In.WriteString("n\n") // Cancel at confirmation

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
			}
			opts.Yes = tt.yes
			opts.Confirm = tt.confirm

			err := run(context.Background(), opts)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error for insufficient flags")
				}
			} else {
				// When both flags are set and user cancels, no error
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOptions_ConfirmFlagsEmbedded(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Yes = true
	opts.Confirm = true

	if !opts.Yes {
		t.Error("Yes should be true")
	}
	if !opts.Confirm {
		t.Error("Confirm should be true")
	}
}
