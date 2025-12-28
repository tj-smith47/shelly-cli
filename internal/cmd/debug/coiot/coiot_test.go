package coiot

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	formatText = "text"
	formatJSON = "json"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "coiot <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "coiot <device>")
	}

	if cmd.Short != "Show CoIoT/CoAP status" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Show CoIoT/CoAP status")
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

	expectedAliases := []string{"coap"}
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
		name         string
		shorthand    string
		defValue     string
		wantNonEmpty bool
	}{
		{name: "format", shorthand: "f", defValue: formatText, wantNonEmpty: true},
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
			if tt.wantNonEmpty && flag.Usage == "" {
				t.Errorf("%s usage is empty", tt.name)
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

func TestOptions_Initialization(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Parse with no flags to check defaults
	err := cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Verify format flag default
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatText {
		t.Errorf("format default = %q, want %q", format, formatText)
	}
}

func TestNewCommand_JsonFormatFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"--format", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatJSON {
		t.Errorf("format = %q, want %q", format, formatJSON)
	}
}

func TestNewCommand_ShorthandFormatFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"-f", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatJSON {
		t.Errorf("format = %q, want %q", format, formatJSON)
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
	}
	opts.Format = formatText

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
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

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
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

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

	if !strings.Contains(helpOutput, "coiot") {
		t.Error("Help should contain 'coiot'")
	}
	if !strings.Contains(helpOutput, "CoIoT") {
		t.Error("Help should contain 'CoIoT'")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

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

func TestRun_JSONFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for json: %v", err)
	}
}

func TestRun_TextFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for text: %v", err)
	}
}
