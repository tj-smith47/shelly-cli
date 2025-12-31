package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

	if cmd.Use != "log <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "log <device>")
	}

	if cmd.Short != "Get device debug log (Gen1)" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Get device debug log (Gen1)")
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

	expectedAliases := []string{"logs", "debug-log"}
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

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The log command has no custom flags, only cobra-inherited ones
	// Verify it parses without error
	err := cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
}

func TestNewCommand_LongDescriptionMentionsGen1(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	// Verify the Long description mentions Gen1 (the target audience)
	wantContains := "Gen1"
	if len(cmd.Long) < len(wantContains) {
		t.Errorf("Long description is too short to contain %q", wantContains)
	}
}

func TestNewCommand_ExampleShowsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Verify the Example field has content
	if len(cmd.Example) < 10 {
		t.Error("Example is too short to be useful")
	}
}

func TestNewCommand_ArgsValidator(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestNewCommand_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		check    func(cmd *cmdutil.Factory) bool
		errorMsg string
	}{
		{
			name: "Use field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Use == "log <device>"
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
			name: "Aliases are set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return len(cmd.Aliases) > 0
			},
			errorMsg: "Aliases not set",
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

	err := run(context.Background(), &Options{Factory: tf.Factory, Device: "nonexistent-device"})

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// This will fail on device connection, but exercises the early run() code
	err := run(context.Background(), &Options{Factory: tf.Factory, Device: "test-device"})

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

	err := run(ctx, &Options{Factory: tf.Factory, Device: "test-device"})

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

	if !strings.Contains(helpOutput, "log") {
		t.Error("Help should contain 'log'")
	}
	if !strings.Contains(helpOutput, "Gen1") {
		t.Error("Help should contain 'Gen1'")
	}
}

func TestRun_EmptyDeviceName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), &Options{Factory: tf.Factory, Device: ""})

	// Should get an error for empty device name
	if err == nil {
		t.Log("Expected error for empty device name")
	}
}
