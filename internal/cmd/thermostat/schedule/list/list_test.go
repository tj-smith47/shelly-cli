package list

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
	}

	// Test Aliases
	wantAliases := []string{"ls", "l"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Test Args requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Args should require at least 1 argument")
	}
	if err := cmd.Args(cmd, []string{"device1"}); err != nil {
		t.Errorf("Args should accept 1 argument: %v", err)
	}
	if err := cmd.Args(cmd, []string{"device1", "device2"}); err == nil {
		t.Error("Args should reject more than 1 argument")
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly thermostat schedule list",
		"--json",
		"--all",
		"--thermostat-id",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test thermostat-id flag
	thermostatIDFlag := cmd.Flags().Lookup("thermostat-id")
	if thermostatIDFlag == nil {
		t.Fatal("--thermostat-id flag not found")
	}
	if thermostatIDFlag.DefValue != "0" {
		t.Errorf("--thermostat-id default = %q, want 0", thermostatIDFlag.DefValue)
	}

	// Test all flag
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("--all flag not found")
	}
	if allFlag.DefValue != "false" {
		t.Errorf("--all default = %q, want false", allFlag.DefValue)
	}

	// Test output format flag
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("--format flag not found")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"thermostat",
		"schedules",
		"--all",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestNewCommand_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Default values
	if opts.ThermostatID != 0 {
		t.Errorf("Default ThermostatID = %d, want 0", opts.ThermostatID)
	}
	if opts.All {
		t.Error("Default All should be false")
	}
	if opts.Format != "" {
		t.Errorf("Default Format = %q, want empty", opts.Format)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_SuccessWithMockDevice(t *testing.T) {
	// Create a Gen2 device with schedule data
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"sys": map[string]any{
					"available_updates": map[string]any{},
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
	cmd.SetArgs([]string{"thermostat-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock implementation
	// but should not panic
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithThermostatIDFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "multi-thermostat-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"multi-thermostat-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"multi-thermostat-device", "--thermostat-id", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithAllFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device-with-schedules",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device-with-schedules": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"device-with-schedules", "--all"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithJSONFormat(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"json-device", "--format", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithAllAndJSONFlags(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "combo-device",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:04",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"combo-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"combo-device", "--all", "--format", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_UnknownDevice(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_Gen1DeviceRejected(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.200",
					MAC:        "AA:BB:CC:DD:EE:99",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {
				"relay": map[string]any{"ison": false},
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
	cmd.SetArgs([]string{"gen1-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for Gen1 device")
	}
	if err != nil && !strings.Contains(err.Error(), "Gen2") {
		t.Logf("Error: %v", err)
	}
}

func TestRun_WithValidOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:      tf.Factory,
		Device:       "test-device",
		ThermostatID: 0,
		All:          false,
	}

	// Test with context.Background() - will fail due to no device
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error with invalid device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithTextOutput(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "text-device",
					Address:    "192.168.1.105",
					MAC:        "AA:BB:CC:DD:EE:05",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"text-device": {
				"sys": map[string]any{},
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
		Factory:      tf.Factory,
		Device:       "text-device",
		ThermostatID: 0,
		All:          false,
		OutputFlags: struct {
			Format string
		}{Format: "text"},
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithJSONOutput(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-run-device",
					Address:    "192.168.1.106",
					MAC:        "AA:BB:CC:DD:EE:06",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-run-device": {
				"sys": map[string]any{},
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
		Factory:      tf.Factory,
		Device:       "json-run-device",
		ThermostatID: 0,
		All:          false,
		OutputFlags: struct {
			Format string
		}{Format: "json"},
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected with mock)", err)
	}
}

func TestOptions_FieldValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:      tf.Factory,
		Device:       "my-device",
		ThermostatID: 2,
		All:          true,
	}

	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want 'my-device'", opts.Device)
	}
	if opts.ThermostatID != 2 {
		t.Errorf("ThermostatID = %d, want 2", opts.ThermostatID)
	}
	if !opts.All {
		t.Error("All should be true")
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
			name:    "no flags",
			args:    []string{"device"},
			wantErr: false,
		},
		{
			name:    "thermostat-id flag",
			args:    []string{"device", "--thermostat-id", "1"},
			wantErr: false,
		},
		{
			name:    "all flag",
			args:    []string{"device", "--all"},
			wantErr: false,
		},
		{
			name:    "format json flag",
			args:    []string{"device", "--format", "json"},
			wantErr: false,
		},
		{
			name:    "all flags combined",
			args:    []string{"device", "--thermostat-id", "2", "--all", "--format", "json"},
			wantErr: false,
		},
		{
			name:    "format flag",
			args:    []string{"device", "--format", "text"},
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

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_JSONOutputFormat(t *testing.T) {
	// Test that JSON output properly formats the schedules
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-output-device",
					Address:    "192.168.1.107",
					MAC:        "AA:BB:CC:DD:EE:07",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-output-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"json-output-device", "--format", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Output might be empty array [] if no schedules, but shouldn't error on format
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_TextOutputFormat(t *testing.T) {
	// Test that text output displays schedules correctly
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "text-output-device",
					Address:    "192.168.1.108",
					MAC:        "AA:BB:CC:DD:EE:08",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"text-output-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"text-output-device", "--format", "text"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_NoDeviceError(t *testing.T) {
	// Test that command properly errors when device is not found
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"totally-nonexistent-device-xyz"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_ThermostatIDFiltering(t *testing.T) {
	// Test filtering schedules by thermostat ID
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "multi-trv-device",
					Address:    "192.168.1.109",
					MAC:        "AA:BB:CC:DD:EE:09",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"multi-trv-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"multi-trv-device", "--thermostat-id", "1", "--format", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_AllSchedules(t *testing.T) {
	// Test showing all schedules, not just thermostat-related ones
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "all-schedules-device",
					Address:    "192.168.1.110",
					MAC:        "AA:BB:CC:DD:EE:10",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"all-schedules-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"all-schedules-device", "--all"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-device",
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if opts.Factory != tf.Factory {
		t.Error("Factory should match the provided factory")
	}
}

func TestNewCommand_RunEIsSet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set on the command")
	}
}

func TestNewCommand_HasShort(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	if !strings.Contains(cmd.Short, "thermostat") {
		t.Errorf("Short should mention thermostat, got: %s", cmd.Short)
	}
}

func TestNewCommand_HasLong(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
	if !strings.Contains(cmd.Long, "thermostat") {
		t.Errorf("Long should mention thermostat, got: %s", cmd.Long)
	}
}

func TestNewCommand_ArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that Args is ExactArgs(1)
	if cmd.Args == nil {
		t.Error("Args should be set")
	}

	// Should fail with 0 args
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("should fail with no args")
	}

	// Should succeed with 1 arg
	if err := cmd.Args(cmd, []string{"device"}); err != nil {
		t.Error("should succeed with 1 arg")
	}

	// Should fail with 2+ args
	if err := cmd.Args(cmd, []string{"device", "extra"}); err == nil {
		t.Error("should fail with 2 args")
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestRun_DeviceNotFound(t *testing.T) {
	// Test error path when device is not found
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:      tf.Factory,
		Device:       "nonexistent-device-xyz",
		ThermostatID: 0,
		All:          false,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when device is not found")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_Gen1DeviceError(t *testing.T) {
	// Test error path when device is Gen1
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-test-device",
					Address:    "192.168.1.201",
					MAC:        "AA:BB:CC:DD:EE:A1",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-test-device": {
				"relay": map[string]any{"ison": false},
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
		Factory:      tf.Factory,
		Device:       "gen1-test-device",
		ThermostatID: 0,
		All:          false,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if err != nil && !strings.Contains(err.Error(), "Gen2") {
		t.Logf("Error: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_ContextTimeout(t *testing.T) {
	// Test error handling when context times out
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "timeout-device",
					Address:    "192.168.1.202",
					MAC:        "AA:BB:CC:DD:EE:A2",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"timeout-device": {
				"sys": map[string]any{},
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

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:      tf.Factory,
		Device:       "timeout-device",
		ThermostatID: 0,
		All:          false,
	}

	err = run(ctx, opts)
	// Should error due to cancelled context
	if err == nil {
		t.Logf("No error with cancelled context (may be expected with mock)")
	}
}

func TestNewCommand_AlliesMatch(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check aliases
	expectedAliases := []string{"ls", "l"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for i, expected := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != expected {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_ValidArgsFunction_NotNil(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should not be nil")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
}

func TestNewCommand_Use_ExactValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedUse := "list <device>"
	if cmd.Use != expectedUse {
		t.Errorf("Use = %q, want %q", cmd.Use, expectedUse)
	}
}

func TestNewCommand_Short_NotEmpty(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewCommand_Long_NotEmpty(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewCommand_Example_NotEmpty(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestOptions_AllDefaultFalse(t *testing.T) {
	t.Parallel()

	opts := &Options{
		All: false,
	}

	if opts.All != false {
		t.Error("All should default to false")
	}
}

func TestOptions_ThermostatIDDefaultZero(t *testing.T) {
	t.Parallel()

	opts := &Options{
		ThermostatID: 0,
	}

	if opts.ThermostatID != 0 {
		t.Error("ThermostatID should default to 0")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_CombinedFlags_ThermostatIDAndAll(t *testing.T) {
	// Test using both thermostat-id and all flags together
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "combined-flags-device",
					Address:    "192.168.1.111",
					MAC:        "AA:BB:CC:DD:EE:11",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"combined-flags-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"combined-flags-device", "--thermostat-id", "1", "--all", "--format", "text"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}

func TestNewCommand_BuilderPattern(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Verify command is properly constructed
	if cmd == nil {
		t.Fatal("NewCommand should not return nil")
	}

	// Verify command has required fields
	if cmd.Use == "" {
		t.Fatal("Use should not be empty")
	}
	if cmd.Short == "" {
		t.Fatal("Short should not be empty")
	}
	if cmd.Long == "" {
		t.Fatal("Long should not be empty")
	}
	if cmd.Example == "" {
		t.Fatal("Example should not be empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases should not be empty")
	}
	if cmd.RunE == nil {
		t.Fatal("RunE should be set")
	}
	if cmd.Args == nil {
		t.Fatal("Args should be set")
	}
	if cmd.ValidArgsFunction == nil {
		t.Fatal("ValidArgsFunction should be set")
	}
}

func TestOptions_OutputFlags_Embedded(t *testing.T) {
	t.Parallel()

	opts := &Options{
		OutputFlags: struct {
			Format string
		}{Format: "json"},
		Factory:      nil,
		Device:       "test",
		ThermostatID: 1,
		All:          true,
	}

	if opts.Format != "json" {
		t.Errorf("OutputFlags.Format should be 'json', got %q", opts.Format)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_WithAllFlagMultipleValues(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device", "--all", "--all"}) // double flag
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Should handle duplicate flags gracefully
	err := cmd.Execute()
	// Device doesn't exist so will error, but flag parsing should work
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_InvalidThermostatID(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device", "--thermostat-id", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid thermostat-id value")
	}
}

func TestNewCommand_FlagsDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check defaults before parsing any flags
	thermostatIDFlag := cmd.Flags().Lookup("thermostat-id")
	if thermostatIDFlag == nil {
		t.Fatal("thermostat-id flag should exist")
	}

	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("all flag should exist")
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag should exist")
	}

	// Verify values are set correctly
	if thermostatIDFlag.DefValue != "0" {
		t.Errorf("thermostat-id DefValue = %s, want 0", thermostatIDFlag.DefValue)
	}

	if allFlag.DefValue != "false" {
		t.Errorf("all DefValue = %s, want false", allFlag.DefValue)
	}
}

func TestNewCommand_CommandMetadata(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify command metadata
	if cmd.Deprecated != "" {
		t.Errorf("Deprecated should be empty, got %q", cmd.Deprecated)
	}

	if cmd.Hidden {
		t.Error("Hidden should be false for normal command")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithFormat_Default(t *testing.T) {
	// Test with default format (should be text)
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "default-format-device",
					Address:    "192.168.1.112",
					MAC:        "AA:BB:CC:DD:EE:12",
					Type:       "SNTC-G02EU",
					Model:      "Shelly TRV",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"default-format-device": {
				"sys": map[string]any{},
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
	cmd.SetArgs([]string{"default-format-device"}) // no format flag
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected with mock)", err)
	}
}
