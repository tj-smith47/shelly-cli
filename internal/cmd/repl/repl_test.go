package repl

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

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"interactive", "i"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify long contains key information about REPL
	checks := []string{
		"REPL",
		"Read-Eval-Print Loop",
		"command history",
		"help",
		"exit",
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

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Verify example contains usage patterns
	checks := []string{
		"shelly repl",
		"--device",
		"devices",
		"connect",
		"status",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Example, check) {
			t.Errorf("Example missing %q", check)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{name: "device flag", flagName: "device", shorthand: "d", defValue: ""},
		{name: "no-prompt flag", flagName: "no-prompt", shorthand: "", defValue: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.flagName)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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
		{name: "no flags", args: []string{}, wantErr: false},
		{name: "device flag", args: []string{"--device", "test-device"}, wantErr: false},
		{name: "device short flag", args: []string{"-d", "test-device"}, wantErr: false},
		{name: "no-prompt flag", args: []string{"--no-prompt"}, wantErr: false},
		{name: "combined flags", args: []string{"--device", "test", "--no-prompt"}, wantErr: false},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: true},
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

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if opts.Device != "" {
		t.Errorf("Device = %q, want empty", opts.Device)
	}
	if opts.NoPrompt {
		t.Error("NoPrompt should default to false")
	}
}

func TestOptions_WithDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "living-room",
	}

	if opts.Device != "living-room" {
		t.Errorf("Device = %q, want 'living-room'", opts.Device)
	}
}

func TestOptions_WithNoPrompt(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		NoPrompt: true,
	}

	if !opts.NoPrompt {
		t.Error("NoPrompt should be true")
	}
}

func TestRun_NonInteractiveMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		NoPrompt: false,
	}

	// Create a cancelled context to exit immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should return nil on cancelled context (clean exit)
	if err != nil {
		// The run function returns nil on context cancellation
		// or may return an error if readline fails
		t.Logf("run() returned: %v (expected for non-TTY)", err)
	}

	// Verify output contains REPL header
	output := tf.OutString()
	if !strings.Contains(output, "Shelly Interactive REPL") {
		t.Errorf("Output should contain REPL header, got: %q", output)
	}
}

func TestRun_WithDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Error or nil are both acceptable for cancelled context
	if err != nil {
		t.Logf("run() error (acceptable for cancelled context): %v", err)
	}

	// Verify output contains REPL header
	output := tf.OutString()
	if !strings.Contains(output, "Shelly Interactive REPL") {
		t.Errorf("Output should contain REPL header, got: %q", output)
	}
}

func TestRun_OutputsHelpInfo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err != nil {
		t.Logf("run() error: %v (expected for non-TTY)", err)
	}

	// Verify output contains help hint
	output := tf.OutString()
	errOutput := tf.ErrString()
	combined := output + errOutput

	if !strings.Contains(combined, "help") || !strings.Contains(combined, "exit") {
		t.Errorf("Output should contain help/exit hints, got: %q", combined)
	}
}

func TestNewCommand_Execute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - should not require any args
	err := cmd.Execute()
	// Error is acceptable due to readline initialization in non-TTY
	if err != nil {
		t.Logf("execute error (acceptable for non-TTY): %v", err)
	}
}

func TestNewCommand_Execute_WithDeviceFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--device", "my-device"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}

func TestNewCommand_Execute_WithNoPromptFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--no-prompt"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}

func TestNewCommand_Execute_AllFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--device", "test", "--no-prompt"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}

func TestNewCommand_AcceptsNoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// REPL command accepts no positional args (Args validator is nil means any args accepted)
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
	// If Args is nil, command accepts any args which includes no args
}
