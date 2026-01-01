package list

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// Format constants for testing.
const (
	formatJSON = "json"
	formatText = "text"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
	}

	if cmd.Short != "List thermostats" {
		t.Errorf("Short = %q, want %q", cmd.Short, "List thermostats")
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

	expectedAliases := map[string]bool{"ls": true, "l": true}

	if len(cmd.Aliases) < 1 {
		t.Fatal("expected at least 1 alias")
	}

	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
	}

	// Verify both expected aliases exist
	found := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		found[alias] = true
	}

	if !found["ls"] {
		t.Error("expected alias \"ls\" not found")
	}
	if !found["l"] {
		t.Error("expected alias \"l\" not found")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use == "list <device>" },
			errMsg:    "Use should be 'list <device>'",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			errMsg:    "RunE should be set",
		},
		{
			name:      "uses ExactArgs(1)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			errMsg:    "Args should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			errMsg:    "ValidArgsFunction should be set for device completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.checkFunc(cmd) {
				t.Error(tt.errMsg)
			}
		})
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
			name:    "no args error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg valid",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args error",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "IP address valid",
			args:    []string{"192.168.1.100"},
			wantErr: false,
		},
		{
			name:    "device name valid",
			args:    []string{"living-room"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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

	// Verify it mentions relevant terms
	if !strings.Contains(cmd.Long, "thermostat") {
		t.Error("Long description should mention 'thermostat'")
	}

	if !strings.Contains(cmd.Long, "Shelly") {
		t.Error("Long description should mention 'Shelly'")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example should not be empty")
	}

	// Verify example contains expected patterns
	patterns := []string{"shelly", "thermostat", "list"}
	for _, pattern := range patterns {
		if !strings.Contains(strings.ToLower(cmd.Example), pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}

	// Example should show JSON usage
	if !strings.Contains(cmd.Example, "--json") && !strings.Contains(cmd.Example, "-f json") {
		t.Error("Example should demonstrate JSON output flag")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check format flag exists
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag should exist")
	}

	if formatFlag.Shorthand != "f" {
		t.Errorf("format flag shorthand = %q, want %q", formatFlag.Shorthand, "f")
	}

	if formatFlag.DefValue != "text" {
		t.Errorf("format flag default = %q, want %q", formatFlag.DefValue, "text")
	}
}

func TestNewCommand_FormatFlagValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "text format",
			format:  formatText,
			wantErr: false,
		},
		{
			name:    "json format",
			format:  formatJSON,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Flags().Set("format", tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("Setting format=%q: error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		})
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

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond) // Ensure timeout expires

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_RunE_PassesDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel to make it fail fast
	cmd.SetContext(ctx)

	// Execute should fail due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestOptions_DeviceAssignment(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create command and set args
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-gateway"})

	// Cancel context to avoid actual execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - it will fail but that's fine, we're testing arg parsing
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestNewCommand_RunE_SetContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"gateway-device"})

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	cmd.SetContext(ctx)

	// Execute should fail due to context
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with expired context")
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
	}

	// Should fail because device doesn't exist in config and isn't a valid IP
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when device not found")
	}
}

func TestRun_InvalidIP(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "999.999.999.999", // Invalid IP
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error with invalid IP address")
	}
}

func TestRun_EmptyDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error with empty device")
	}
}

func TestNewCommand_AcceptsHostname(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test various valid device name formats
	deviceNames := []string{
		"gateway",
		"living-room",
		"device_01",
		"mydevice123",
	}

	for _, name := range deviceNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, []string{name})
			if err != nil {
				t.Errorf("Command should accept device name %q, got error: %v", name, err)
			}
		})
	}
}

func TestNewCommand_ExampleShowsJQ(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The example should show jq usage for filtering
	if !strings.Contains(cmd.Example, "jq") {
		t.Error("Example should demonstrate jq filtering usage")
	}
}

func TestRun_JSONFormatWithCancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set format to json
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context even with JSON format")
	}
}

func TestRun_TextFormatWithCancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context even with text format")
	}
}

func TestOptions_FormatField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "text format",
			format: formatText,
		},
		{
			name:   "json format",
			format: formatJSON,
		},
		{
			name:   "empty format",
			format: "",
		},
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

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-gateway",
	}
	opts.Format = formatJSON

	if opts.Factory != tf.Factory {
		t.Error("Factory not set correctly")
	}

	if opts.Device != "my-gateway" {
		t.Errorf("Device = %q, want %q", opts.Device, "my-gateway")
	}

	if opts.Format != formatJSON {
		t.Errorf("Format = %q, want %q", opts.Format, formatJSON)
	}
}

func TestNewCommand_FlagShorthand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that -f shorthand works
	err := cmd.ParseFlags([]string{"-f", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag should exist")
	}

	if formatFlag.Value.String() != formatJSON {
		t.Errorf("format flag value = %q, want %q", formatFlag.Value.String(), formatJSON)
	}
}

func TestNewCommand_ParseDeviceArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set args
	cmd.SetArgs([]string{"blu-gateway"})

	// Parse (don't execute - just validate args parsing)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// This will execute and fail due to cancelled context,
	// but the args parsing happens first
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestRun_VariousDeviceFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device string
	}{
		{
			name:   "simple name",
			device: "gateway",
		},
		{
			name:   "hyphenated name",
			device: "blu-gateway-1",
		},
		{
			name:   "ip address",
			device: "192.168.1.100",
		},
		{
			name:   "hostname with dots",
			device: "gateway.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)

			opts := &Options{
				Factory: tf.Factory,
				Device:  tt.device,
			}

			// All should fail because device won't be found,
			// but we're testing that run() handles various device formats
			err := run(context.Background(), opts)
			if err == nil {
				t.Error("Expected error when device not found")
			}
		})
	}
}

func TestNewCommand_RunE_IntegrationFlow(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test that RunE is wired correctly
	if cmd.RunE == nil {
		t.Fatal("RunE should not be nil")
	}

	// Execute with a device argument
	cmd.SetArgs([]string{"test-thermostat-gateway"})

	// Create cancelled context to short-circuit actual network call
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute and verify error is propagated
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestNewCommand_FlagsIntegration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set format flag and device arg
	cmd.SetArgs([]string{"gateway", "-f", formatJSON})

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from cancelled context")
	}

	// Verify flag was parsed correctly
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag.Value.String() != formatJSON {
		t.Errorf("format flag = %q, want %q", formatFlag.Value.String(), formatJSON)
	}
}

func TestOptions_OutputFlagsEmbedding(t *testing.T) {
	t.Parallel()

	// Verify OutputFlags is properly embedded
	opts := &Options{}
	opts.Format = "yaml" // Access through embedded struct

	if opts.Format != "yaml" {
		t.Errorf("Format = %q, want %q", opts.Format, "yaml")
	}
}

func TestNewCommand_AllProperties(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Comprehensive property checks
	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if cmd.Long == "" {
		t.Error("Long is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
	if cmd.Args == nil {
		t.Error("Args is nil")
	}
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}
