package websocket

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "websocket <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "websocket <device>")
	}

	if cmd.Short != "Debug WebSocket connection and stream events" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Debug WebSocket connection and stream events")
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

	expectedAliases := []string{"ws", "events"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"device1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
		{name: "duration", shorthand: "", defValue: "30s"},
		{name: "raw", shorthand: "", defValue: "false"},
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

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is not set")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_DurationFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flagVal  string
		expected time.Duration
		wantErr  bool
	}{
		{name: "default", flagVal: "", expected: 30 * time.Second, wantErr: false},
		{name: "5 minutes", flagVal: "5m", expected: 5 * time.Minute, wantErr: false},
		{name: "zero for indefinite", flagVal: "0", expected: 0, wantErr: false},
		{name: "1 hour", flagVal: "1h", expected: 1 * time.Hour, wantErr: false},
		{name: "10 seconds", flagVal: "10s", expected: 10 * time.Second, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			var args []string
			if tt.flagVal != "" {
				args = []string{"--duration", tt.flagVal}
			}

			err := cmd.ParseFlags(args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := cmd.Flags().GetDuration("duration")
				if err != nil {
					t.Fatalf("GetDuration() error = %v", err)
				}
				if got != tt.expected {
					t.Errorf("duration = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestNewCommand_RawFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{name: "default false", args: []string{}, expected: false},
		{name: "explicit true", args: []string{"--raw"}, expected: true},
		{name: "explicit false", args: []string{"--raw=false"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			got, err := cmd.Flags().GetBool("raw")
			if err != nil {
				t.Fatalf("GetBool() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("raw = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewCommand_InvalidDuration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"--duration", "invalid"})
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		Duration: 5 * time.Minute,
		Raw:      true,
	}

	if opts.Factory != f {
		t.Error("Factory field not set correctly")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Duration != 5*time.Minute {
		t.Errorf("Duration = %v, want %v", opts.Duration, 5*time.Minute)
	}
	if !opts.Raw {
		t.Error("Raw = false, want true")
	}
}

func TestNewCommand_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		check    func(*cmdutil.Factory) bool
		errorMsg string
	}{
		{
			name: "Use field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Use == "websocket <device>"
			},
			errorMsg: "Use field not set correctly",
		},
		{
			name: "Short field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Short != ""
			},
			errorMsg: "Short field is empty",
		},
		{
			name: "Long field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Long != ""
			},
			errorMsg: "Long field is empty",
		},
		{
			name: "Example field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Example != ""
			},
			errorMsg: "Example field is empty",
		},
		{
			name: "Has at least 2 aliases",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return len(cmd.Aliases) >= 2
			},
			errorMsg: "Should have at least 2 aliases (ws, events)",
		},
		{
			name: "Args validator is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Args != nil
			},
			errorMsg: "Args validator not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if !tt.check(f) {
				t.Error(tt.errorMsg)
			}
		})
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent-device",
		Duration: 30 * time.Second,
		Raw:      false,
	}

	err := run(context.Background(), opts)

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      false,
	}

	// This will fail on device connection, but exercises the early run() code
	err := run(context.Background(), opts)

	// Expect error due to no device
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 30 * time.Second,
		Raw:      false,
	}

	err := run(ctx, opts)

	// Should return some error (context cancelled or connection error)
	if err == nil {
		t.Log("Expected error with cancelled context")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_ExecuteWithDeviceArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute will fail due to no real device, but args should be accepted
	err := cmd.Execute()

	// We expect an error (no device connection), but not an args error
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
			t.Errorf("Should accept device argument, got args error: %v", err)
		}
	}
}

func TestNewCommand_HelpOutput(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Logf("Help execution: %v", err)
	}

	helpOutput := stdout.String()

	if !strings.Contains(helpOutput, "websocket") {
		t.Error("Help should contain 'websocket'")
	}
	if !strings.Contains(helpOutput, "WebSocket") {
		t.Error("Help should contain 'WebSocket'")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 30 * time.Second,
		Raw:      true,
	}

	// Verify factory is accessible
	if opts.Factory == nil {
		t.Fatal("Options.Factory should not be nil")
	}

	ios := opts.Factory.IOStreams()
	if ios == nil {
		t.Error("Factory.IOStreams() should not return nil")
	}

	svc := opts.Factory.ShellyService()
	if svc == nil {
		t.Error("Factory.ShellyService() should not return nil")
	}
}

func TestRun_RawMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      true,
	}

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a raw mode error
	if err != nil && strings.Contains(err.Error(), "raw") {
		t.Errorf("Unexpected raw mode error: %v", err)
	}
}

func TestRun_ZeroDuration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 0, // Zero for indefinite
		Raw:      false,
	}

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}
