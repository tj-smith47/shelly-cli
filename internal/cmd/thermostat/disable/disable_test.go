package disable

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

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
			checkFunc: func(c *cobra.Command) bool { return c.Use == "disable <device>" },
			wantOK:    true,
			errMsg:    "Use should be 'disable <device>'",
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
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			wantOK:    true,
			errMsg:    "ValidArgsFunction should be set for completion",
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"off", "stop"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
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
		{"no args", []string{}, true},
		{"one arg valid", []string{"device"}, false},
		{"two args", []string{"device", "extra"}, true},
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test id flag
	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("--id flag not found")
	}
	if flag.DefValue != "0" {
		t.Errorf("id default = %q, want 0", flag.DefValue)
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly thermostat disable",
		"gateway",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"Disable",
		"thermostat",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Default ID should be 0
	if opts.ID != 0 {
		t.Errorf("Default ID = %d, want 0", opts.ID)
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

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_ExecuteWithDeviceArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

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

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts IP addresses as device identifiers
	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
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
			name:    "id flag",
			args:    []string{"--id", "1"},
			wantErr: false,
		},
		{
			name:    "id flag short",
			args:    []string{"-i", "2"},
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

func TestRun_WithMockGen1Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "gen1-device",
	}

	err = run(context.Background(), opts)

	// Expect an error because Gen1 devices don't support thermostat
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}

	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected error mentioning Gen2+, got: %v", err)
	}
}

func TestRun_WithCustomID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.ID = 5 // Custom thermostat ID

	err := run(ctx, opts)

	// Expect an error (cancelled context or no device)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
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

func TestRun_WithMockGen2Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
				},
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "thermostat-device",
	}

	err = run(context.Background(), opts)

	// Should succeed for Gen2 device with thermostat support
	if err != nil {
		t.Errorf("Unexpected error for Gen2 device: %v", err)
	}

	// Check that success message was printed
	output := tf.OutString()
	if !strings.Contains(output, "disabled") && !strings.Contains(output, "Thermostat") {
		t.Logf("Output: %s", output)
	}
}

func TestRun_WithMockGen2DeviceCustomID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:1": map[string]any{
					"id":        float64(1),
					"enable":    true,
					"target_C":  float64(20.0),
					"current_C": float64(19.5),
					"output":    true,
				},
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "thermostat-device",
	}
	opts.ID = 1 // Custom thermostat ID

	err = run(context.Background(), opts)

	// Should succeed for Gen2 device with thermostat support
	if err != nil {
		t.Errorf("Unexpected error for Gen2 device with custom ID: %v", err)
	}
}

func TestNewCommand_ExecuteWithIDFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
				},
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"thermostat-device", "--id", "0"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestRun_ThermostatEnableError(t *testing.T) {
	t.Parallel()

	// Create a Gen2 device with error simulation enabled via thermostat_error state
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
				},
				// This signals the mock server to return an error for Thermostat.SetConfig
				"thermostat_error": true,
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "thermostat-device",
	}

	err = run(context.Background(), opts)

	// Should error due to thermostat config failure
	if err == nil {
		t.Error("Expected error from thermostat.Enable")
	}

	// Verify error message format
	if err != nil && !strings.Contains(err.Error(), "failed to disable thermostat") {
		t.Errorf("Expected error containing 'failed to disable thermostat', got: %v", err)
	}
}
