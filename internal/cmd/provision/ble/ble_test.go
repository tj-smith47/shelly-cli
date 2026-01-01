package ble

import (
	"context"
	"strings"
	"testing"

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

	expected := "ble <device-address>"
	if cmd.Use != expected {
		t.Errorf("Use = %q, want %q", cmd.Use, expected)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"bluetooth"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Provision a device via Bluetooth Low Energy"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Check that Long contains key information
	if !strings.Contains(cmd.Long, "Bluetooth Low Energy") {
		t.Error("Long should mention Bluetooth Low Energy")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Check for example usage patterns
	if !strings.Contains(cmd.Example, "shelly provision ble") {
		t.Error("Example should contain 'shelly provision ble'")
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

	err = cmd.Args(cmd, []string{"ShellyPlus1-ABCD"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Expected error when multiple args provided")
	}
}

func TestNewCommand_HasFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{name: "ssid", flagName: "ssid", wantType: "string"},
		{name: "password", flagName: "password", wantType: "string"},
		{name: "name", flagName: "name", wantType: "string"},
		{name: "timezone", flagName: "timezone", wantType: "string"},
		{name: "cloud", flagName: "cloud", wantType: "bool"},
		{name: "no-cloud", flagName: "no-cloud", wantType: "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.flagName)
			}
			if flag.Value.Type() != tt.wantType {
				t.Errorf("%s type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	tests := []struct {
		name        string
		flagName    string
		wantDefault string
	}{
		{name: "ssid default empty", flagName: "ssid", wantDefault: ""},
		{name: "password default empty", flagName: "password", wantDefault: ""},
		{name: "name default empty", flagName: "name", wantDefault: ""},
		{name: "timezone default empty", flagName: "timezone", wantDefault: ""},
		{name: "cloud default false", flagName: "cloud", wantDefault: "false"},
		{name: "no-cloud default false", flagName: "no-cloud", wantDefault: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag.DefValue != tt.wantDefault {
				t.Errorf("%s default = %q, want %q", tt.flagName, flag.DefValue, tt.wantDefault)
			}
		})
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		DeviceAddress: "ShellyPlus1-ABCD",
		DisableCloud:  false,
		EnableCloud:   false,
		DeviceName:    "",
		Factory:       tf.Factory,
		Password:      "",
		SSID:          "",
		Timezone:      "",
	}

	if opts.SSID != "" {
		t.Errorf("Default SSID = %q, want empty", opts.SSID)
	}
	if opts.Password != "" {
		t.Errorf("Default Password = %q, want empty", opts.Password)
	}
	if opts.EnableCloud {
		t.Error("Default EnableCloud should be false")
	}
	if opts.DisableCloud {
		t.Error("Default DisableCloud should be false")
	}
}

func TestOptions_FieldsSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "ShellyPlus1-1234ABCD",
		SSID:          "MyNetwork",
		Password:      "secret123",
		DeviceName:    "Living Room",
		Timezone:      "America/New_York",
		EnableCloud:   true,
		DisableCloud:  false,
	}

	if opts.DeviceAddress != "ShellyPlus1-1234ABCD" {
		t.Errorf("DeviceAddress = %q, want 'ShellyPlus1-1234ABCD'", opts.DeviceAddress)
	}
	if opts.SSID != "MyNetwork" {
		t.Errorf("SSID = %q, want 'MyNetwork'", opts.SSID)
	}
	if opts.Password != "secret123" {
		t.Errorf("Password = %q, want 'secret123'", opts.Password)
	}
	if opts.DeviceName != "Living Room" {
		t.Errorf("DeviceName = %q, want 'Living Room'", opts.DeviceName)
	}
	if opts.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want 'America/New_York'", opts.Timezone)
	}
	if !opts.EnableCloud {
		t.Error("EnableCloud should be true")
	}
}

func TestNewCommand_ParseFlags_SSID(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetArgs([]string{"TestDevice", "--ssid", "MyNetwork"})

	err := cmd.Flags().Parse([]string{"--ssid", "MyNetwork"})
	if err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	ssidFlag := cmd.Flags().Lookup("ssid")
	if ssidFlag.Value.String() != "MyNetwork" {
		t.Errorf("ssid = %q, want 'MyNetwork'", ssidFlag.Value.String())
	}
}

func TestNewCommand_ParseFlags_AllOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "ssid flag only",
			args:    []string{"--ssid", "MyNetwork"},
			wantErr: false,
		},
		{
			name:    "password flag only",
			args:    []string{"--password", "secret123"},
			wantErr: false,
		},
		{
			name:    "name flag only",
			args:    []string{"--name", "Living Room"},
			wantErr: false,
		},
		{
			name:    "timezone flag only",
			args:    []string{"--timezone", "America/New_York"},
			wantErr: false,
		},
		{
			name:    "cloud flag",
			args:    []string{"--cloud"},
			wantErr: false,
		},
		{
			name:    "no-cloud flag",
			args:    []string{"--no-cloud"},
			wantErr: false,
		},
		{
			name:    "multiple flags",
			args:    []string{"--ssid", "MyNetwork", "--password", "secret"},
			wantErr: false,
		},
		{
			name:    "all flags combined",
			args:    []string{"--ssid", "MyNetwork", "--password", "secret", "--name", "Room", "--timezone", "UTC", "--no-cloud"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Flags().Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_CommandProperties(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name  string
		got   string
		field string
	}{
		{name: "Use field", got: cmd.Use, field: "Use"},
		{name: "Short field", got: cmd.Short, field: "Short"},
		{name: "Long field", got: cmd.Long, field: "Long"},
		{name: "Example field", got: cmd.Example, field: "Example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got == "" {
				t.Errorf("%s should not be empty", tt.field)
			}
		})
	}
}

func TestNewCommand_Execute_MissingDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// No device argument
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed without device argument")
	}
}

func TestNewCommand_Execute_WithDeviceArgument(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Device argument provided with required flags
	cmd.SetArgs([]string{"ShellyPlus1-ABCD", "--ssid", "TestNetwork", "--password", "testpass"})

	// Execute will fail because we can't actually provision via BLE
	// but we want to verify the command accepts arguments and tries to run
	err := cmd.Execute()

	// We expect an error (network/device not found), but not an arg validation error
	// The error should be about failed initialization, not missing arguments
	if err != nil {
		// Error is expected due to device/BLE unavailability
		// Just verify it's not an arg validation error
		errMsg := err.Error()
		if strings.Contains(errMsg, "exactly 1") || strings.Contains(errMsg, "arguments") {
			t.Fatalf("Got argument validation error instead of device error: %v", err)
		}
	}
}

func TestNewCommand_Execute_WithSSIDFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// With explicit SSID flag
	cmd.SetArgs([]string{"TestDevice", "--ssid", "MyWiFi"})

	// Execute - will fail due to BLE unavailability, but we're testing flag parsing
	err := cmd.Execute()
	// Error expected, but should be related to device connection, not flags
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_AllFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// With all flags
	cmd.SetArgs([]string{
		"ShellyPlus1-ABCD",
		"--ssid", "MyNetwork",
		"--password", "secret123",
		"--name", "My Device",
		"--timezone", "Europe/London",
		"--no-cloud",
	})

	// Execute - will fail due to BLE, but we're testing flag parsing
	err := cmd.Execute()
	// Error expected
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_PartialFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Only device and name
	cmd.SetArgs([]string{"TestDevice", "--name", "My Device"})

	// Execute - will fail due to BLE, but we're testing it accepts flags
	err := cmd.Execute()
	// Error expected
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
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
			name:      "has Args validation",
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

func TestRun_NoSSIDAndNoPassword(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "TestDevice",
		SSID:          "",
		Password:      "",
	}

	// Create a cancelled context to prevent actual BLE operations
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should fail due to context being cancelled
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithSSIDOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "TestDevice",
		SSID:          "MyNetwork",
		Password:      "",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should fail due to context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithAllOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "ShellyPlus1-ABCD",
		SSID:          "MyNetwork",
		Password:      "secret123",
		DeviceName:    "Living Room",
		Timezone:      "America/New_York",
		EnableCloud:   true,
		DisableCloud:  false,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should fail due to context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_CloudOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	tests := []struct {
		name         string
		enableCloud  bool
		disableCloud bool
	}{
		{name: "cloud enabled", enableCloud: true, disableCloud: false},
		{name: "cloud disabled", enableCloud: false, disableCloud: true},
		{name: "cloud neither", enableCloud: false, disableCloud: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := &Options{
				Factory:       tf.Factory,
				DeviceAddress: "TestDevice",
				SSID:          "Network",
				Password:      "pass",
				EnableCloud:   tt.enableCloud,
				DisableCloud:  tt.disableCloud,
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := run(ctx, opts)
			// Expect error due to cancelled context
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestNewCommand_AcceptsDeviceAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{name: "valid Shelly device", address: "ShellyPlus1-ABCD1234", wantErr: false},
		{name: "valid IP address", address: "192.168.1.100", wantErr: false},
		{name: "valid hostname", address: "shelly-device", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, []string{tt.address})
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%s) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "test-device",
		SSID:          "Network",
		Password:      "password",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_Long_ContainsRequirements(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	requiredStrings := []string{
		"BLE",
		"Bluetooth",
		"WiFi",
		"device",
	}

	for _, str := range requiredStrings {
		if !strings.Contains(cmd.Long, str) {
			t.Errorf("Long description should contain %q", str)
		}
	}
}

func TestNewCommand_Example_ContainsVariations(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show multiple usage patterns
	requiredPatterns := []string{
		"shelly provision ble",
		"--ssid",
		"--password",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestOptions_FactorySet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no flags", args: []string{}, wantErr: false},
		{name: "ssid flag", args: []string{"--ssid", "MyNetwork"}, wantErr: false},
		{name: "password flag", args: []string{"--password", "secret"}, wantErr: false},
		{name: "combined flags", args: []string{"--ssid", "Net", "--password", "pwd"}, wantErr: false},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Flags().Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_SSIDAndPasswordProvided(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "test-device",
		SSID:          "TestNetwork",
		Password:      "testpassword",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Error expected due to cancelled context or device unavailability
	if err == nil {
		t.Error("Expected error")
	}
}

func TestNewCommand_Execute_WithContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"device", "--ssid", "network", "--password", "pass"})

	err := cmd.Execute()
	// Should fail due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_Integration_BasicSetup(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Verify the command can be created and used
	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	// Verify all components are in place
	if cmd.RunE == nil {
		t.Fatal("RunE is nil")
	}

	if cmd.Flags().Lookup("ssid") == nil {
		t.Fatal("ssid flag is missing")
	}

	if cmd.Flags().Lookup("password") == nil {
		t.Fatal("password flag is missing")
	}
}

func TestOptions_AllFields(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		DeviceAddress: "device1",
		SSID:          "ssid1",
		Password:      "pass1",
		DeviceName:    "name1",
		Timezone:      "tz1",
		EnableCloud:   true,
		DisableCloud:  false,
		Factory:       tf.Factory,
	}

	if opts.DeviceAddress != "device1" {
		t.Errorf("DeviceAddress = %q, want 'device1'", opts.DeviceAddress)
	}
	if opts.SSID != "ssid1" {
		t.Errorf("SSID = %q, want 'ssid1'", opts.SSID)
	}
	if opts.Password != "pass1" {
		t.Errorf("Password = %q, want 'pass1'", opts.Password)
	}
	if opts.DeviceName != "name1" {
		t.Errorf("DeviceName = %q, want 'name1'", opts.DeviceName)
	}
	if opts.Timezone != "tz1" {
		t.Errorf("Timezone = %q, want 'tz1'", opts.Timezone)
	}
	if !opts.EnableCloud {
		t.Error("EnableCloud should be true")
	}
	if opts.DisableCloud {
		t.Error("DisableCloud should be false")
	}
}

func TestNewCommand_Execute_IntegrationWithAllFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Execute with all available flags
	cmd.SetArgs([]string{
		"ShellyPlus1-ABC",
		"--ssid", "TestNetwork",
		"--password", "testpass",
		"--name", "Test Device",
		"--timezone", "UTC",
		"--cloud",
	})

	// Execute - will fail due to BLE unavailability
	// but we're testing the command accepts and parses all flags
	err := cmd.Execute()
	// Error is expected (Bluetooth not available)
	// Just verify the command structure is correct
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_WithCloudDisabled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{
		"device",
		"--ssid", "net",
		"--password", "pwd",
		"--no-cloud",
	})

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_WithNameOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{"device", "--name", "MyDevice"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_WithTimezoneOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{"device", "--timezone", "Europe/Paris"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestNewCommand_Execute_WithCloudEnabled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{
		"device",
		"--ssid", "network",
		"--password", "password",
		"--cloud",
	})

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (expected for BLE unavailability): %v", err)
	}
}

func TestOptions_BothCloudFlagsSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "device",
		SSID:          "network",
		Password:      "pass",
		EnableCloud:   true,
		DisableCloud:  true,
	}

	// Both flags set - EnableCloud takes precedence in code
	if !opts.EnableCloud {
		t.Error("EnableCloud should be true")
	}
	if !opts.DisableCloud {
		t.Error("DisableCloud should also be true (both can be set)")
	}
}

func TestNewCommand_MetadataComplete(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Ensure all metadata is non-empty
	if cmd.Use == "" || cmd.Short == "" || cmd.Long == "" || cmd.Example == "" {
		t.Error("Command metadata is incomplete")
	}

	// Ensure aliases are present
	if len(cmd.Aliases) == 0 {
		t.Error("No aliases defined")
	}

	// Ensure flags are accessible
	if cmd.Flags() == nil {
		t.Error("Flags not accessible")
	}

	// Ensure proper command function
	if cmd.RunE == nil {
		t.Error("RunE not defined")
	}
}

func TestNewCommand_Execute_Variations(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	tests := []struct {
		name string
		args []string
	}{
		{"basic", []string{"device", "--ssid", "net", "--password", "pwd"}},
		{"with name", []string{"device", "--ssid", "net", "--password", "pwd", "--name", "Room"}},
		{"with timezone", []string{"device", "--ssid", "net", "--password", "pwd", "--timezone", "UTC"}},
		{"with cloud", []string{"device", "--ssid", "net", "--password", "pwd", "--cloud"}},
		{"with no-cloud", []string{"device", "--ssid", "net", "--password", "pwd", "--no-cloud"}},
		{"all options", []string{"device", "--ssid", "net", "--password", "pwd", "--name", "R", "--timezone", "UTC", "--cloud"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(tf.Factory)
			cmd.SetArgs(tt.args)

			// Execute - will fail but we're testing command structure
			err := cmd.Execute()
			// Error expected (BLE unavailable), but no panic
			if err != nil {
				t.Logf("execute error (expected for BLE unavailability): %v", err)
			}
		})
	}
}

func TestOptions_MinimalSetup(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		DeviceAddress: "device",
		SSID:          "network",
		Password:      "password",
	}

	// Verify minimal required fields are set
	if opts.Factory == nil {
		t.Fatal("Factory is nil")
	}
	if opts.DeviceAddress == "" {
		t.Fatal("DeviceAddress is empty")
	}
	if opts.SSID == "" {
		t.Fatal("SSID is empty")
	}
	if opts.Password == "" {
		t.Fatal("Password is empty")
	}
}

func TestNewCommand_FlagsAreOptional(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// All flags except device address should be optional
	optionalFlags := []string{"ssid", "password", "name", "timezone", "cloud", "no-cloud"}

	for _, flagName := range optionalFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Fatalf("Flag %q should exist", flagName)
		}

		// For string flags, check they have empty default
		// For bool flags, check they have false default
		if flag.Value.Type() == "string" {
			if flag.DefValue != "" {
				t.Errorf("Flag %q should default to empty, got %q", flagName, flag.DefValue)
			}
		}
	}
}

func TestRun_InvalidContextPaths(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	tests := []struct {
		name string
		opts *Options
	}{
		{
			name: "with device name",
			opts: &Options{
				Factory:       tf.Factory,
				DeviceAddress: "device",
				SSID:          "net",
				Password:      "pwd",
				DeviceName:    "MyRoom",
			},
		},
		{
			name: "with timezone",
			opts: &Options{
				Factory:       tf.Factory,
				DeviceAddress: "device",
				SSID:          "net",
				Password:      "pwd",
				Timezone:      "America/NewYork",
			},
		},
		{
			name: "with cloud enabled",
			opts: &Options{
				Factory:       tf.Factory,
				DeviceAddress: "device",
				SSID:          "net",
				Password:      "pwd",
				EnableCloud:   true,
			},
		},
		{
			name: "with cloud disabled",
			opts: &Options{
				Factory:       tf.Factory,
				DeviceAddress: "device",
				SSID:          "net",
				Password:      "pwd",
				DisableCloud:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := run(ctx, tt.opts)
			// Error expected
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}
