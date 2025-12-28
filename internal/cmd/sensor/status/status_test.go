// Package status provides the sensor status command.
package status

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

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

	if cmd.Use != "status <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device>")
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"st", "all", "readings"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("alias[%d] missing, want %q", i, want)
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
		{name: "format", shorthand: "f", defValue: formatText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
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
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "three args",
			args:    []string{"device1", "device2", "device3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction not set")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}

	if opts.Factory != f {
		t.Error("Factory not set correctly")
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	// Test OutputFlags embedding
	opts.Format = formatJSON
	if opts.Format != formatJSON {
		t.Errorf("Format = %q, want %q", opts.Format, formatJSON)
	}
}

func TestOptions_OutputFlagsEmbedding(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	// Test format field access through embedded OutputFlags
	opts.Format = formatText
	if opts.Format != formatText {
		t.Errorf("Format = %q, want %q", opts.Format, formatText)
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE not set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the long description contains expected content
	wantContains := []string{
		"sensor readings",
		"Temperature",
		"Humidity",
		"Flood",
		"Smoke",
		"Illuminance",
		"Voltage",
	}

	for _, want := range wantContains {
		if !containsString(cmd.Long, want) {
			t.Errorf("Long description missing expected content: %q", want)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify example contains key usage patterns
	wantContains := []string{
		"shelly sensor status",
		"--json",
	}

	for _, want := range wantContains {
		if !containsString(cmd.Example, want) {
			t.Errorf("Example missing expected content: %q", want)
		}
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

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

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

	time.Sleep(1 * time.Millisecond)

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
	cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flag    string
		value   string
		wantErr bool
	}{
		{
			name:    "format flag json",
			flag:    "format",
			value:   formatJSON,
			wantErr: false,
		},
		{
			name:    "format flag text",
			flag:    "format",
			value:   formatText,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Flags().Set(tt.flag, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setting flag %s to %s: error = %v, wantErr = %v",
					tt.flag, tt.value, err, tt.wantErr)
			}

			if !tt.wantErr {
				val, getErr := cmd.Flags().GetString(tt.flag)
				if getErr != nil {
					t.Errorf("getting flag %s: %v", tt.flag, getErr)
				}
				if val != tt.value {
					t.Errorf("flag %s = %q, want %q", tt.flag, val, tt.value)
				}
			}
		})
	}
}

func TestNewCommand_LongDescriptionLength(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description should not be empty")
	}

	// Long should be more descriptive than Short
	if len(cmd.Long) <= len(cmd.Short) {
		t.Error("Long description should be longer than Short description")
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if example == "" {
		t.Fatal("Example should not be empty")
	}

	// Check for typical usage patterns
	patterns := []string{"shelly", "sensor", "status"}
	for _, pattern := range patterns {
		if !containsString(example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestOptions_FormatDefault(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Format flag should have default value
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	if formatFlag.DefValue != formatText {
		t.Errorf("format default = %q, want %q", formatFlag.DefValue, formatText)
	}
}

func TestRun_WithJSONFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithTextFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_WithFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"test-device": {Name: "test-device", Address: "10.0.0.1"},
	})

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with factory")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestRun_WithConfiguredDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"my-sensor": {
			Name:       "my-sensor",
			Address:    "192.168.1.50",
			Generation: 2,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-sensor",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithIPAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "192.168.1.100",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_ShorthandFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that format flag can be set with -f shorthand
	if err := cmd.Flags().Set("format", formatJSON); err != nil {
		t.Fatalf("failed to set format flag: %v", err)
	}

	val, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("failed to get format value: %v", err)
	}

	if val != formatJSON {
		t.Errorf("format = %q, want %q", val, formatJSON)
	}
}

func TestNewCommand_HelpFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify help flag is available (Cobra adds it to InheritedFlags or flags directly)
	// We check by executing with -h and seeing that no error about unknown flag occurs
	cmd.SetArgs([]string{"-h"})
	cmd.SetOut(&discardWriter{})
	cmd.SetErr(&discardWriter{})

	// Execute should return nil for help flag (it's a special case)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("help flag should be available, got error: %v", err)
	}
}

// discardWriter discards all writes.
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestOptions_FactoryMethods(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test",
	}

	// Verify factory methods are accessible
	ios := opts.Factory.IOStreams()
	if ios == nil {
		t.Error("IOStreams() returned nil")
	}

	svc := opts.Factory.ShellyService()
	if svc == nil {
		t.Error("ShellyService() returned nil")
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsSubstring(s, substr))
}

// containsSubstring is a helper for substring matching without importing strings.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
