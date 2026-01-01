package config

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
	if cmd.Use != "config <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config <device>")
	}

	// Test Aliases
	wantAliases := []string{"set", "configure"}
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

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

	// Test id flag exists
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("--id flag not found")
	}

	// Test freq flag exists
	freqFlag := cmd.Flags().Lookup("freq")
	if freqFlag == nil {
		t.Fatal("--freq flag not found")
	}
	if freqFlag.DefValue != "0" {
		t.Errorf("--freq default = %q, want %q", freqFlag.DefValue, "0")
	}

	// Test bw flag exists
	bwFlag := cmd.Flags().Lookup("bw")
	if bwFlag == nil {
		t.Fatal("--bw flag not found")
	}

	// Test dr flag exists
	drFlag := cmd.Flags().Lookup("dr")
	if drFlag == nil {
		t.Fatal("--dr flag not found")
	}

	// Test power flag exists
	powerFlag := cmd.Flags().Lookup("power")
	if powerFlag == nil {
		t.Fatal("--power flag not found")
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
		"shelly lora config",
		"--freq",
		"--power",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:    "test-device",
		Factory:   f,
		Freq:      868000000,
		Bandwidth: 7,
		DataRate:  7,
		TxPower:   14,
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	if opts.Freq != 868000000 {
		t.Errorf("Freq = %d, want %d", opts.Freq, 868000000)
	}

	if opts.Bandwidth != 7 {
		t.Errorf("Bandwidth = %d, want %d", opts.Bandwidth, 7)
	}

	if opts.DataRate != 7 {
		t.Errorf("DataRate = %d, want %d", opts.DataRate, 7)
	}

	if opts.TxPower != 14 {
		t.Errorf("TxPower = %d, want %d", opts.TxPower, 14)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_NoOptionsSpecified(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Device:  "test-device",
		// ID is set via ComponentFlags.ID, but run() uses opts.ID directly
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "No configuration options specified") {
		t.Errorf("expected warning about no options, got: %s", output)
	}
	if !strings.Contains(output, "--freq") {
		t.Errorf("expected hint about --freq flag, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_SetFrequency(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Device:  "test-device",
		Freq:    868000000,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Frequency") {
		t.Errorf("expected frequency info, got: %s", output)
	}
	if !strings.Contains(output, "868") {
		t.Errorf("expected 868 MHz in output, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_SetAllOptions(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Factory:   tf.Factory,
		Device:    "test-device",
		Freq:      915000000,
		Bandwidth: 8,
		DataRate:  10,
		TxPower:   20,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Frequency") {
		t.Errorf("expected frequency info, got: %s", output)
	}
	if !strings.Contains(output, "Bandwidth") {
		t.Errorf("expected bandwidth info, got: %s", output)
	}
	if !strings.Contains(output, "Data Rate") {
		t.Errorf("expected data rate info, got: %s", output)
	}
	if !strings.Contains(output, "TX Power") {
		t.Errorf("expected tx power info, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_SetBandwidth(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Factory:   tf.Factory,
		Device:    "test-device",
		Bandwidth: 9,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Bandwidth: 9") {
		t.Errorf("expected bandwidth info with value 9, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_SetDataRate(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Factory:  tf.Factory,
		Device:   "test-device",
		DataRate: 12,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Data Rate: 12") {
		t.Errorf("expected data rate info with value 12, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_SetTxPower(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
		Device:  "test-device",
		TxPower: 17,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "TX Power: 17 dBm") {
		t.Errorf("expected tx power info with value 17, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{},
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
		Device:  "nonexistent",
		Freq:    868000000,
	}
	opts.ID = 100

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithFreq(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
	cmd.SetArgs([]string{"test-device", "--freq", "868000000"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithMultipleFlags(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
	cmd.SetArgs([]string{"test-device", "--freq", "915000000", "--power", "20", "--dr", "7"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "LoRa configuration updated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_NoFlags(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "No configuration options specified") {
		t.Errorf("expected warning about no options, got: %s", output)
	}
}
