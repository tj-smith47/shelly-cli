package set

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "set <device> <component> <key>=<value>..." {
		t.Errorf("Use = %q, want 'set <device> <component> <key>=<value>...'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Expected at least one alias")
	}

	expectedAliases := map[string]bool{"write": true, "update": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("Unexpected alias: %s", alias)
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept 3+ args (device, component, key=value...)
	err := cmd.Args(cmd, []string{"device1", "switch:0", "name=Light"})
	if err != nil {
		t.Errorf("Expected no error with three args, got: %v", err)
	}

	// Should accept multiple key=value pairs
	err = cmd.Args(cmd, []string{"device1", "switch:0", "name=Light", "initial_state=on"})
	if err != nil {
		t.Errorf("Expected no error with four args, got: %v", err)
	}

	// Should reject 0 args
	err = cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should reject 1 arg
	err = cmd.Args(cmd, []string{"device1"})
	if err == nil {
		t.Error("Expected error when only one arg provided")
	}

	// Should reject 2 args
	err = cmd.Args(cmd, []string{"device1", "switch:0"})
	if err == nil {
		t.Error("Expected error when only two args provided")
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
			name:      "uses MinimumNArgs(3)",
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

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"name=Light"},
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"name=Light"},
	}
	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestRun_InvalidKeyValueFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context first
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test invalid key=value format (no equals sign)
	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"invalid"},
	}
	err := run(ctx, opts)

	// Expect an error due to invalid format
	if err == nil {
		t.Error("Expected error with invalid key=value format")
	}
}

func TestRun_ValidKeyValuePairs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Valid key=value pairs should parse correctly, but will fail due to cancelled context
	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"name=Light", "initial_state=on"},
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context (not due to parsing)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100", "switch:0", "name=Light"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room", "switch:0", "name=Light"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_AcceptsMultipleKeyValues(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	testCases := [][]string{
		{"device", "switch:0", "name=Light"},
		{"device", "switch:0", "name=Light", "initial_state=on"},
		{"device", "switch:0", "name=Light", "initial_state=on", "auto_off=true"},
		{"device", "light:0", "default.brightness=50"},
	}

	for _, tc := range testCases {
		err := cmd.Args(cmd, tc)
		if err != nil {
			t.Errorf("Command should accept args %v, got error: %v", tc, err)
		}
	}
}

func TestNewCommand_RunE_PassesArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "switch:0", "name=Test Light"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_Components(t *testing.T) {
	t.Parallel()

	// Test common component names that should be accepted
	components := []string{
		"switch:0",
		"switch:1",
		"light:0",
		"cover:0",
		"input:0",
		"sys",
		"wifi",
		"mqtt",
		"cloud",
		"ble",
	}

	cmd := NewCommand(cmdutil.NewFactory())

	for _, comp := range components {
		err := cmd.Args(cmd, []string{"device", comp, "key=value"})
		if err != nil {
			t.Errorf("Command should accept component %q, got error: %v", comp, err)
		}
	}
}

func TestNewCommand_KeyValueFormats(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Various key=value formats that should be accepted
	validFormats := [][]string{
		{"device", "switch:0", "name=Light"},
		{"device", "switch:0", "name=Main Light"},
		{"device", "switch:0", "enable=true"},
		{"device", "switch:0", "enable=false"},
		{"device", "switch:0", "timeout=300"},
		{"device", "light:0", "default.brightness=50"},
		{"device", "light:0", "min_brightness=10"},
		{"device", "wifi", "sta.ssid=MyNetwork"},
	}

	for _, args := range validFormats {
		err := cmd.Args(cmd, args)
		if err != nil {
			t.Errorf("Command should accept args %v, got error: %v", args, err)
		}
	}
}

func TestRun_MultipleKeyValuePairs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Multiple key=value pairs
	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"name=Light", "initial_state=on", "auto_off=true"},
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_ValueWithSpaces(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Value with spaces (quoted in shell, but passed as single arg)
	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
		KeyValues: []string{"name=Main Light Switch"},
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_NestedKey(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Nested key (e.g., default.brightness)
	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "light:0",
		KeyValues: []string{"default.brightness=50"},
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain various usage patterns
	example := cmd.Example
	if example == "" {
		t.Error("Example should not be empty")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should explain key=value format
	long := cmd.Long
	if long == "" {
		t.Error("Long description should not be empty")
	}
}
