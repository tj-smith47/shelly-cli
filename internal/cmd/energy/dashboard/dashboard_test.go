package dashboard

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
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
	if cmd.Use != "dashboard" {
		t.Errorf("Use = %q, want %q", cmd.Use, "dashboard")
	}

	// Test Aliases
	wantAliases := []string{"dash", "summary"}
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
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test devices flag
	devicesFlag := cmd.Flags().Lookup("devices")
	if devicesFlag == nil {
		t.Fatal("--devices flag not found")
	}

	// Test cost flag
	costFlag := cmd.Flags().Lookup("cost")
	if costFlag == nil {
		t.Fatal("--cost flag not found")
	}
	if costFlag.DefValue != "0" {
		t.Errorf("--cost default = %q, want %q", costFlag.DefValue, "0")
	}

	// Test currency flag
	currencyFlag := cmd.Flags().Lookup("currency")
	if currencyFlag == nil {
		t.Fatal("--currency flag not found")
	}
	if currencyFlag.DefValue != "USD" {
		t.Errorf("--currency default = %q, want %q", currencyFlag.DefValue, "USD")
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
		"shelly energy dashboard",
		"--devices",
		"--cost",
		"--currency",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestRun_NoDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Empty config with no devices
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error for no devices, got: %v", err)
	}

	// Should warn about no devices
	output := tf.TestIO.Out.String() + tf.TestIO.ErrOut.String()
	if !strings.Contains(output, "No devices") {
		t.Errorf("expected warning about no devices, got: %s", output)
	}
}

func TestRun_ConfigError(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	// Create factory with a config function that returns an error
	f := cmdutil.NewFactory().SetIOStreams(ios)
	// Override the Config function to return an error
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("test config error")
	}

	opts := &Options{
		Factory: f,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Fatal("expected error for config error")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("expected config error, got: %v", err)
	}
}

func TestRun_WithSpecificDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set up with specific device flag
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--devices", "device1,device2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Execute - will fail on device connection but exercises the devices flag parsing
	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Logf("expected timeout/connection error: %v", err)
	}
}

func TestRun_WithCostEstimation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add a device to config
	tf.Config.Devices["test-device"] = model.Device{
		Name:       "test-device",
		Address:    "192.168.1.100",
		Generation: 2,
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--cost", "0.15", "--currency", "EUR"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Execute - will timeout but exercises cost flag parsing
	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Logf("expected timeout error: %v", err)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.CostPerKwh != 0 {
		t.Errorf("Default CostPerKwh = %f, want 0", opts.CostPerKwh)
	}

	if opts.CostCurrency != "" {
		t.Errorf("Default CostCurrency = %q, want empty", opts.CostCurrency)
	}

	if opts.Factory != nil {
		t.Error("Default Factory should be nil")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMockServer(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em": {
				"em:0": map[string]any{
					"id":               0,
					"a_current":        1.5,
					"a_voltage":        230.0,
					"a_act_power":      345.0,
					"total_act_power":  1000.0,
					"total_current":    4.5,
					"total_aprt_power": 1050.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithCostEstimationAndMock(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em": {
				"em:0": map[string]any{
					"id":              0,
					"total_act_power": 1000.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--cost", "0.12", "--currency", "USD"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMultipleDevices(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "em-device-1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "em-device-2",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"em-device-1": {
				"em:0": map[string]any{
					"id":              0,
					"total_act_power": 500.0,
				},
			},
			"em-device-2": {
				"em:0": map[string]any{
					"id":              0,
					"total_act_power": 750.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithDevicesFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device-a",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "device-b",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "device-c",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device-a": {"em:0": map[string]any{"id": 0, "total_act_power": 100.0}},
			"device-b": {"em:0": map[string]any{"id": 0, "total_act_power": 200.0}},
			"device-c": {"em:0": map[string]any{"id": 0, "total_act_power": 300.0}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Only request device-a and device-c
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--devices", "device-a,device-c"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

func TestRun_DevicesSortedAlphabetically(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add devices in non-alphabetical order
	tf.Config.Devices["zebra-device"] = model.Device{Name: "zebra-device", Address: "192.168.1.3", Generation: 2}
	tf.Config.Devices["alpha-device"] = model.Device{Name: "alpha-device", Address: "192.168.1.1", Generation: 2}
	tf.Config.Devices["middle-device"] = model.Device{Name: "middle-device", Address: "192.168.1.2", Generation: 2}

	opts := &Options{
		Factory: tf.Factory,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Run - will timeout on device connection but exercises sorting
	if err := run(ctx, opts); err != nil {
		t.Logf("expected timeout error: %v", err)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"aggregated energy dashboard",
		"power consumption",
		"cost estimation",
		"--devices",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
	}
}

func TestRun_WithRegisteredDevicesFromConfig(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add multiple devices to config
	tf.Config.Devices["device1"] = model.Device{
		Name:       "device1",
		Address:    "192.168.1.1",
		Generation: 2,
	}
	tf.Config.Devices["device2"] = model.Device{
		Name:       "device2",
		Address:    "192.168.1.2",
		Generation: 2,
	}

	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Execute - exercises the config device loading path
	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Logf("expected timeout error: %v", err)
	}
}

func TestRun_CostWithZeroEnergy(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Setup config manager for the test with a registered device
	cfg := &config.Config{
		Devices: map[string]model.Device{
			"test-device": {Name: "test-device", Address: "192.168.1.100", Generation: 2},
		},
	}
	mgr := config.NewTestManager(cfg)
	tf.SetConfigManager(mgr)

	opts := &Options{
		Factory:      tf.Factory,
		CostPerKwh:   0.15,
		CostCurrency: "EUR",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Run - exercises cost calculation path
	if err := run(ctx, opts); err != nil {
		t.Logf("expected timeout error: %v", err)
	}
}
