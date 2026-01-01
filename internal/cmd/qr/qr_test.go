package qr

import (
	"context"
	"strings"
	"testing"

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

	if cmd.Use != "qr <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "qr <device>")
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

	expectedAliases := []string{"qrcode"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias at index %d", i)
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
		{name: "wifi", shorthand: "", defValue: "false"},
		{name: "no-qr", shorthand: "", defValue: "false"},
		{name: "size", shorthand: "", defValue: "256"},
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
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg",
			args:      []string{"device1"},
			wantError: false,
		},
		{
			name:      "two args",
			args:      []string{"device1", "extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := cmd.Args(cmd, tt.args)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description contains key information
	checks := []string{
		"QR code",
		"Device web interface URL",
		"WiFi network configuration",
		"--wifi",
		"ASCII art",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Long, check) {
			t.Errorf("Long description missing %q", check)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify examples contain key usage patterns
	checks := []string{
		"shelly qr kitchen-light",
		"--wifi",
		"--no-qr",
		"-o json",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Example, check) {
			t.Errorf("Example missing %q", check)
		}
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command creates options with correct defaults
	// The Options struct is internal, but we can verify via flags
	wifiFlag := cmd.Flags().Lookup("wifi")
	if wifiFlag == nil {
		t.Fatal("wifi flag not found")
	}
	if wifiFlag.DefValue != "false" {
		t.Errorf("wifi default = %q, want %q", wifiFlag.DefValue, "false")
	}

	noQRFlag := cmd.Flags().Lookup("no-qr")
	if noQRFlag == nil {
		t.Fatal("no-qr flag not found")
	}
	if noQRFlag.DefValue != "false" {
		t.Errorf("no-qr default = %q, want %q", noQRFlag.DefValue, "false")
	}

	sizeFlag := cmd.Flags().Lookup("size")
	if sizeFlag == nil {
		t.Fatal("size flag not found")
	}
	if sizeFlag.DefValue != "256" {
		t.Errorf("size default = %q, want %q", sizeFlag.DefValue, "256")
	}
}

func TestNewCommand_FlagMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagName  string
		flagValue string
		wantError bool
	}{
		{
			name:      "set wifi flag",
			flagName:  "wifi",
			flagValue: "true",
			wantError: false,
		},
		{
			name:      "set no-qr flag",
			flagName:  "no-qr",
			flagValue: "true",
			wantError: false,
		},
		{
			name:      "set size flag",
			flagName:  "size",
			flagValue: "512",
			wantError: false,
		},
		{
			name:      "set size flag invalid",
			flagName:  "size",
			flagValue: "not-a-number",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Flags().Set(tt.flagName, tt.flagValue)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantError {
				flag := cmd.Flags().Lookup(tt.flagName)
				if flag.Value.String() != tt.flagValue {
					t.Errorf("flag value = %q, want %q", flag.Value.String(), tt.flagValue)
				}
			}
		})
	}
}

func TestNewCommand_RunERequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no device arg provided")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Verify Run is not set (we should use RunE, not Run)
	if cmd.Run != nil {
		t.Error("Run should be nil, use RunE instead")
	}
}

func TestNewCommand_ArgsValidator(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// qr command uses ExactArgs(1)
	if cmd.Args == nil {
		t.Error("Args validator should be set")
	}

	// Test that it requires exactly 1 arg
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Should reject zero args")
	}
	if err := cmd.Args(cmd, []string{"one"}); err != nil {
		t.Errorf("Should accept one arg, got error: %v", err)
	}
	if err := cmd.Args(cmd, []string{"one", "two"}); err == nil {
		t.Error("Should reject two args")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command can be accessed by name
	if cmd.Name() != "qr" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "qr")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func(cmd *cobra.Command) bool
		wantOK bool
		errMsg string
	}{
		{
			name:   "has use",
			check:  func(c *cobra.Command) bool { return c.Use != "" },
			wantOK: true,
			errMsg: "Use should not be empty",
		},
		{
			name:   "has short",
			check:  func(c *cobra.Command) bool { return c.Short != "" },
			wantOK: true,
			errMsg: "Short should not be empty",
		},
		{
			name:   "has long",
			check:  func(c *cobra.Command) bool { return c.Long != "" },
			wantOK: true,
			errMsg: "Long should not be empty",
		},
		{
			name:   "has example",
			check:  func(c *cobra.Command) bool { return c.Example != "" },
			wantOK: true,
			errMsg: "Example should not be empty",
		},
		{
			name:   "has aliases",
			check:  func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK: true,
			errMsg: "Aliases should not be empty",
		},
		{
			name:   "has RunE",
			check:  func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK: true,
			errMsg: "RunE should be set",
		},
		{
			name:   "uses ExactArgs(1)",
			check:  func(c *cobra.Command) bool { return c.Args != nil },
			wantOK: true,
			errMsg: "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			if tt.check(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
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
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "wifi flag",
			args:    []string{"--wifi"},
			wantErr: false,
		},
		{
			name:    "no-qr flag",
			args:    []string{"--no-qr"},
			wantErr: false,
		},
		{
			name:    "size flag",
			args:    []string{"--size", "512"},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
		{
			name:    "multiple flags",
			args:    []string{"--wifi", "--no-qr", "--size", "128"},
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

func TestNewCommand_WithDeviceArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Execute - will attempt to generate QR for the device
	// Error is expected since device is not reachable
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error from network call: %v", err)
	}

	// Check that some output was produced (even if error)
	output := tf.OutString()
	errOutput := tf.ErrString()
	if output == "" && errOutput == "" {
		t.Error("expected some output")
	}
}

func TestNewCommand_WithWiFiFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device", "--wifi"})

	// Execute - will attempt to generate WiFi QR for the device
	// Error is expected since device is not reachable
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error from network call: %v", err)
	}

	// Verify command executed (may have errors but should produce output)
	output := tf.OutString()
	errOutput := tf.ErrString()
	if output == "" && errOutput == "" {
		t.Error("expected some output")
	}
}

func TestNewCommand_WithNoQRFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device", "--no-qr"})

	// Execute - will attempt to get QR info without displaying QR
	// Error is expected since device is not reachable
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error from network call: %v", err)
	}

	// Verify command executed
	output := tf.OutString()
	errOutput := tf.ErrString()
	if output == "" && errOutput == "" {
		t.Error("expected some output")
	}
}

func TestNewCommand_WithCancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - may fail due to cancelled context
	err := cmd.Execute()
	// We expect an error or the command to complete quickly
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
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

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts named devices
	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestOptions_Initialization(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if opts.WiFi {
		t.Error("WiFi should default to false")
	}
	if opts.NoQR {
		t.Error("NoQR should default to false")
	}
	if opts.Size != 256 {
		t.Errorf("Size = %d, want 256", opts.Size)
	}
}

func TestOptions_WiFiEnabled(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Factory: cmdutil.NewFactory(),
		WiFi:    true,
		NoQR:    false,
		Size:    512,
	}

	if !opts.WiFi {
		t.Error("WiFi should be true")
	}
	if opts.Size != 512 {
		t.Errorf("Size = %d, want 512", opts.Size)
	}
}

func TestOptions_NoQREnabled(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Factory: cmdutil.NewFactory(),
		WiFi:    false,
		NoQR:    true,
		Size:    256,
	}

	if !opts.NoQR {
		t.Error("NoQR should be true")
	}
}

func TestNewCommand_RejectsZeroArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no args provided")
	}
}

func TestNewCommand_RejectsTwoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device1", "device2"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when two args provided")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output contains expected content
	output := tf.OutString()
	if !strings.Contains(output, "QR Code for") {
		t.Errorf("Output should contain 'QR Code for', got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_NoQR(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		WiFi:    false,
		NoQR:    true,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output contains expected content
	output := tf.OutString()
	if !strings.Contains(output, "Content:") {
		t.Errorf("Output should contain 'Content:', got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_WiFi(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		WiFi:    true,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output contains expected content
	output := tf.OutString()
	if !strings.Contains(output, "QR Code for") {
		t.Errorf("Output should contain 'QR Code for', got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_Gen1Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {},
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "gen1-device", opts)
	// Gen1 devices are not supported for QR generation
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2") {
		t.Errorf("Error should mention Gen2, got: %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "nonexistent-device", opts)
	// Should fail because device is not found
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_CancelledContext(t *testing.T) {
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = run(ctx, "test-device", opts)
	// Cancelled context should return error
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_WiFiAndNoQR(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		WiFi:    true,
		NoQR:    true,
		Size:    512,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output contains expected content
	output := tf.OutString()
	if !strings.Contains(output, "Content:") {
		t.Errorf("Output should contain 'Content:', got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_StructuredOutput(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Set JSON output format via viper
	viper.Set("output", "json")
	defer viper.Set("output", "")

	opts := &Options{
		Factory: tf.Factory,
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify JSON output
	output := tf.OutString()
	if !strings.Contains(output, `"device"`) {
		t.Errorf("Output should contain JSON structure, got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithEmptyMAC(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify MAC line is not printed for empty MAC
	output := tf.OutString()
	if strings.Contains(output, "MAC:") && strings.Contains(output, "MAC: \n") {
		t.Errorf("Output should not contain empty MAC")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithEmptyModel(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "",
					Model:      "",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify Model line is not printed for empty Model
	output := tf.OutString()
	if strings.Contains(output, "Model: \n") {
		t.Errorf("Output should not contain empty Model")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithDifferentSize(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		WiFi:    false,
		NoQR:    false,
		Size:    128, // Different size
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_YAMLOutput(t *testing.T) {
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Set YAML output format
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	opts := &Options{
		Factory: tf.Factory,
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify YAML output
	output := tf.OutString()
	if !strings.Contains(output, "device:") {
		t.Errorf("Output should contain YAML structure, got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithBothEmptyMACAndModel(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "",
					Type:       "",
					Model:      "",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		WiFi:    false,
		NoQR:    false,
		Size:    256,
	}

	err = run(context.Background(), "test-device", opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should NOT contain "MAC:" or "Model:" lines since they're empty
	if strings.Contains(output, "MAC: ") || strings.Contains(output, "Model: ") {
		t.Errorf("Output should not contain empty MAC/Model, got: %q", output)
	}
}
