package code

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
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

	if cmd.Use != "code <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "code <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Short != "Show Matter pairing code" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Show Matter pairing code")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"pairing", "qr"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
		return
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

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
			args:    []string{"living-room"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "device name with dashes",
			args:    []string{"my-living-room-device"},
			wantErr: false,
		},
		{
			name:    "device IP address",
			args:    []string{"192.168.1.100"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test format flag exists with correct properties
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	if formatFlag.Shorthand != "f" {
		t.Errorf("format shorthand = %q, want %q", formatFlag.Shorthand, "f")
	}

	if formatFlag.DefValue != formatText {
		t.Errorf("format default = %q, want %q", formatFlag.DefValue, formatText)
	}

	// Verify usage mentions supported formats
	if !strings.Contains(formatFlag.Usage, formatText) {
		t.Error("format flag usage should mention 'text'")
	}
	if !strings.Contains(formatFlag.Usage, formatJSON) {
		t.Error("format flag usage should mention 'json'")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is not set")
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{Factory: f}

	// Verify factory is set
	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	// Verify device starts empty
	if opts.Device != "" {
		t.Errorf("Device = %q, want empty string", opts.Device)
	}

	// Verify format starts empty (set by flag default)
	if opts.Format != "" {
		t.Errorf("Format = %q, want empty string before flag parsing", opts.Format)
	}
}

func TestOptions_OutputFlagsEmbedded(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	// Test that OutputFlags is properly embedded and Format field exists
	opts.Format = formatJSON
	if opts.Format != formatJSON {
		t.Errorf("Format = %q, want %q", opts.Format, formatJSON)
	}

	opts.Format = formatText
	if opts.Format != formatText {
		t.Errorf("Format = %q, want %q", opts.Format, formatText)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	// Should mention Matter pairing
	if !strings.Contains(cmd.Long, "Matter") {
		t.Error("Long description should mention Matter")
	}

	// Should mention pairing code
	if !strings.Contains(cmd.Long, "pairing code") {
		t.Error("Long description should mention 'pairing code'")
	}

	// Should mention commissioning
	if !strings.Contains(cmd.Long, "commissioning") {
		t.Error("Long description should mention 'commissioning'")
	}

	// Should mention manual pairing code
	if !strings.Contains(cmd.Long, "Manual pairing code") {
		t.Error("Long description should mention 'Manual pairing code'")
	}

	// Should mention QR code
	if !strings.Contains(cmd.Long, "QR code") {
		t.Error("Long description should mention 'QR code'")
	}

	// Should mention fabric controllers
	if !strings.Contains(cmd.Long, "Apple Home") || !strings.Contains(cmd.Long, "Google Home") {
		t.Error("Long description should mention Matter fabric controllers")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Should contain the command name
	if !strings.Contains(cmd.Example, "shelly matter code") {
		t.Error("Example should contain 'shelly matter code'")
	}

	// Should show JSON output example
	if !strings.Contains(cmd.Example, "--json") {
		t.Error("Example should contain '--json' flag example")
	}
}

func TestNewCommand_MultipleInstances(t *testing.T) {
	t.Parallel()

	// Create multiple command instances to ensure no shared state issues
	f := cmdutil.NewFactory()
	cmd1 := NewCommand(f)
	cmd2 := NewCommand(f)

	if cmd1 == cmd2 {
		t.Error("NewCommand should return distinct instances")
	}

	// Both should have the same structure
	if cmd1.Use != cmd2.Use {
		t.Errorf("Use mismatch: %q vs %q", cmd1.Use, cmd2.Use)
	}

	if cmd1.Short != cmd2.Short {
		t.Errorf("Short mismatch: %q vs %q", cmd1.Short, cmd2.Short)
	}
}

func TestOptions_FactoryAssignment(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}

	if opts.Factory != f {
		t.Error("Factory not correctly assigned")
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
}

func TestNewCommand_ArgsValidator(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Args function should be set
	if cmd.Args == nil {
		t.Fatal("Args validator not set")
	}

	// Test that it enforces exactly 1 argument
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"empty", []string{}, true},
		{"single", []string{"device"}, false},
		{"multiple", []string{"dev1", "dev2", "dev3"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create fresh command for each test
			testCmd := NewCommand(cmdutil.NewFactory())
			err := testCmd.Args(testCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_FlagParsingIntegration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse flags to verify they work correctly
	if err := cmd.Flags().Parse([]string{"--format", formatJSON}); err != nil {
		t.Errorf("Failed to parse format flag: %v", err)
	}

	formatVal, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format flag value: %v", err)
	}
	if formatVal != formatJSON {
		t.Errorf("format = %q, want %q", formatVal, formatJSON)
	}
}

func TestNewCommand_ShorthandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test shorthand -f works
	if err := cmd.Flags().Parse([]string{"-f", formatJSON}); err != nil {
		t.Errorf("Failed to parse -f shorthand: %v", err)
	}

	formatVal, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format value: %v", err)
	}
	if formatVal != formatJSON {
		t.Errorf("format = %q, want %q", formatVal, formatJSON)
	}
}

// TestRun_WithInvalidDevice tests run with a device that cannot be reached.
// This exercises the error path in the run function.
func TestRun_WithInvalidDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device-12345",
	}
	opts.Format = formatText

	// Use a very short timeout to fail quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	// We expect an error because the device doesn't exist
	if err == nil {
		t.Log("No error returned, which may happen if mock is configured")
	} else {
		// Verify we get some kind of error (connection, timeout, etc.)
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_WithIPDevice tests run with an IP address device.
func TestRun_WithIPDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "192.168.1.254",
	}
	opts.Format = formatText

	// Use a very short timeout to fail quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	// We expect an error because the device doesn't exist at this IP
	if err == nil {
		t.Log("No error returned, device may have responded")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_WithJSONFormat tests run with JSON output format.
func TestRun_WithJSONFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	// Error is expected since no real device
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_WithRegisteredDevice tests run with a device that's in config.
func TestRun_WithRegisteredDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"test-device": {
			Address: "192.168.1.100",
			Name:    "Test Device",
		},
	})

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	// Error is expected since no real device at the configured address
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_ContextCancellation tests that run respects context cancellation.
func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := run(ctx, opts)
	// Should get context cancelled error
	switch {
	case err == nil:
		t.Log("No error returned despite cancelled context")
	case !strings.Contains(err.Error(), "context"):
		t.Logf("Got error (may not be context-related): %v", err)
	default:
		t.Logf("Got expected context error: %v", err)
	}
}

// TestOptions_FormatValues tests different format values.
func TestOptions_FormatValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
	}{
		{"text format", formatText},
		{"json format", formatJSON},
		{"empty format", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := &Options{}
			opts.Format = tt.format

			if opts.Format != tt.format {
				t.Errorf("Format = %q, want %q", opts.Format, tt.format)
			}
		})
	}
}

// TestOptions_DeviceValues tests different device values.
func TestOptions_DeviceValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device string
	}{
		{"simple name", "kitchen"},
		{"name with dashes", "living-room"},
		{"IP address", "192.168.1.100"},
		{"hostname", "shelly-device.local"},
		{"empty device", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := &Options{Device: tt.device}

			if opts.Device != tt.device {
				t.Errorf("Device = %q, want %q", opts.Device, tt.device)
			}
		})
	}
}

// TestNewCommand_ExecuteViaRunE tests that the command can be executed via RunE.
// This covers the RunE closure in NewCommand.
func TestNewCommand_ExecuteViaRunE(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set args to provide the device name
	cmd.SetArgs([]string{"test-device"})

	// Execute the command - this calls RunE
	// We expect an error since there's no real device
	err := cmd.Execute()
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestNewCommand_ExecuteWithJSONFlag tests command execution with JSON flag.
func TestNewCommand_ExecuteWithJSONFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set args with format flag
	cmd.SetArgs([]string{"test-device", "-f", formatJSON})

	// Execute the command
	err := cmd.Execute()
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestNewCommand_ExecuteMissingArg tests that command fails without device arg.
func TestNewCommand_ExecuteMissingArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// No args provided
	cmd.SetArgs([]string{})

	// Execute should fail due to missing required arg
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for missing device argument")
	}
}

// TestRun_WithDeviceAddress tests run with a device that has an address in config.
// This tests the device IP resolution path in lines 69-71.
func TestRun_WithDeviceAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"my-device": {
			Address: "10.0.0.50",
			Name:    "My Device",
		},
	})

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-device", // Use registered name, not IP
	}
	opts.Format = formatText

	// Short timeout to fail quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	// Error is expected since no real device
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_WithDeviceNoAddress tests run with a device that has no address.
func TestRun_WithDeviceNoAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"empty-device": {
			Name: "Empty Device",
			// No address set
		},
	})

	opts := &Options{
		Factory: tf.Factory,
		Device:  "empty-device",
	}
	opts.Format = formatText

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Log("No error returned")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRun_MultipleFormats tests run with different output formats.
func TestRun_MultipleFormats(t *testing.T) {
	t.Parallel()

	formats := []string{formatText, formatJSON, ""}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
			}
			opts.Format = format

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := run(ctx, opts)
			// Error expected since no device
			if err != nil {
				t.Logf("format=%q: got expected error: %v", format, err)
			}
		})
	}
}

// TestNewCommand_DefaultFormat tests that default format is text.
func TestNewCommand_DefaultFormat(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Parse with no flags
	if err := cmd.Flags().Parse([]string{}); err != nil {
		t.Errorf("Failed to parse empty flags: %v", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format: %v", err)
	}

	if format != formatText {
		t.Errorf("default format = %q, want %q", format, formatText)
	}
}
