package get

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
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

	if cmd.Use != "get <device> [component]" {
		t.Errorf("Use = %q, want 'get <device> [component]'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Expected at least one alias")
	}

	expectedAliases := map[string]bool{"show": true, "read": true}
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

	// Should accept 1 arg (device only)
	err := cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	// Should accept 2 args (device + component)
	err = cmd.Args(cmd, []string{"device1", "switch:0"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	// Should reject 0 args
	err = cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should reject 3+ args
	err = cmd.Args(cmd, []string{"device1", "switch:0", "extra"})
	if err == nil {
		t.Error("Expected error when too many args provided")
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
			name:      "uses RangeArgs(1, 2)",
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
		Component: "",
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
		Component: "",
	}
	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestRun_WithComponent(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "switch:0",
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithoutComponent(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:   tf.Factory,
		Device:    "test-device",
		Component: "",
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
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

func TestNewCommand_AcceptsDeviceWithComponent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	testCases := []struct {
		device    string
		component string
	}{
		{"living-room", "switch:0"},
		{"192.168.1.100", "sys"},
		{"kitchen", "wifi"},
		{"bedroom", "light:0"},
	}

	for _, tc := range testCases {
		err := cmd.Args(cmd, []string{tc.device, tc.component})
		if err != nil {
			t.Errorf("Command should accept device %q with component %q, got error: %v", tc.device, tc.component, err)
		}
	}
}

func TestNewCommand_RunE_PassesArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "switch:0"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_RunE_DeviceOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device"})

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
		"eth",
	}

	cmd := NewCommand(cmdutil.NewFactory())

	for _, comp := range components {
		err := cmd.Args(cmd, []string{"device", comp})
		if err != nil {
			t.Errorf("Command should accept component %q, got error: %v", comp, err)
		}
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_GetFullConfig(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "get-device",
					Address:    "192.168.1.220",
					MAC:        "AA:BB:CC:DD:EE:50",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"get-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"get-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	out := tf.OutString()
	if out == "" {
		t.Error("Expected non-empty output")
	}
}

//nolint:paralleltest // Uses shared mock server and viper
func TestRun_GetWithJSONOutput(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-get-device",
					Address:    "192.168.1.221",
					MAC:        "AA:BB:CC:DD:EE:51",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-get-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"json-get-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server and viper
func TestRun_GetWithYAMLOutput(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "yaml-get-device",
					Address:    "192.168.1.222",
					MAC:        "AA:BB:CC:DD:EE:52",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"yaml-get-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	viper.Set("output", "yaml")
	defer viper.Reset()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"yaml-get-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_GetWithValidComponent(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "comp-get-device",
					Address:    "192.168.1.223",
					MAC:        "AA:BB:CC:DD:EE:53",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"comp-get-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Request the "sys" component which should exist in config
	cmd.SetArgs([]string{"comp-get-device", "sys"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_GetWithInvalidComponent(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "invalid-comp-device",
					Address:    "192.168.1.224",
					MAC:        "AA:BB:CC:DD:EE:54",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"invalid-comp-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Request a component that doesn't exist
	cmd.SetArgs([]string{"invalid-comp-device", "nonexistent:99"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent component")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_GetConfigError(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "existing-get-device",
					Address:    "192.168.1.225",
					MAC:        "AA:BB:CC:DD:EE:55",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"existing-get-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}
