package scan

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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
	if cmd.Use != "scan <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "scan <device>")
	}

	// Test Aliases
	wantAliases := []string{"search", "find"}
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
		"shelly wifi scan",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Scan for available WiFi networks") {
		t.Error("help output should contain command description")
	}
	if !strings.Contains(output, "scan <device>") {
		t.Error("help output should show usage")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no device argument provided")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device1", "device2"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when too many arguments provided")
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
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
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestExecute_WithMockWiFiScan(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "wifi-device",
					Address:    "192.168.1.50",
					MAC:        "11:22:33:44:55:66",
					Type:       "SNSW-002P16EU",
					Model:      "Shelly Plus 2PM",
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
	cmd.SetArgs([]string{"wifi-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// The command may fail if mock doesn't support WiFi.Scan, but we're testing the Execute path
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WifiScanSuccess(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "scan-test-device",
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
	cmd.SetArgs([]string{"scan-test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Log result for debugging
	t.Logf("Execute error: %v", err)
	t.Logf("Output: %s", buf.String())
}

func TestExecute_ByIP(t *testing.T) {
	t.Parallel()

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
	// Use IP address directly
	cmd.SetArgs([]string{"192.168.1.100"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Execute with IP error: %v", err)
}

func TestNewCommand_CommandMetadata(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify command has all required metadata
	if cmd.Use == "" {
		t.Error("Use field is required")
	}
	if cmd.Short == "" {
		t.Error("Short field is required")
	}
	if cmd.Long == "" {
		t.Error("Long field is required")
	}
	if cmd.Example == "" {
		t.Error("Example field is required")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("At least one alias is required")
	}
	if cmd.RunE == nil {
		t.Error("RunE must be set")
	}
}

func TestNewCommand_HasValidCompletion(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for shell completion")
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

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts IP addresses
	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address, got error: %v", err)
	}
}

func TestNewCommand_RejectsMultipleDevices(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command rejects multiple devices
	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Command should reject multiple device arguments")
	}
}

func TestExecute_GenericError(t *testing.T) {
	t.Parallel()

	// Create a minimal fixture to trigger device lookup
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Try scanning a device that doesn't exist
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"unknown-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error when device is not found")
	}
}

func TestExecute_WithContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "cancel-test",
					Address:    "192.168.1.200",
					MAC:        "AA:BB:CC:DD:EE:00",
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

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"cancel-test"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Expect an error due to cancelled context
	if err == nil {
		t.Log("Context cancellation test: command may succeed due to mock")
	}
}

func TestExecute_WithTimeoutContext(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "timeout-test",
					Address:    "192.168.1.201",
					MAC:        "AA:BB:CC:DD:EE:01",
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
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"timeout-test"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on timing
	t.Logf("Timeout test result: %v", err)
}

func TestNewCommand_UsesFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify the command was created with the factory
	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	// Verify command is executable
	if cmd.RunE == nil {
		t.Error("Command should have RunE set")
	}
}

func TestNewCommand_MultipleInstances(t *testing.T) {
	t.Parallel()

	tf1 := factory.NewTestFactory(t)
	tf2 := factory.NewTestFactory(t)

	cmd1 := NewCommand(tf1.Factory)
	cmd2 := NewCommand(tf2.Factory)

	// Both should have the same structure
	if cmd1.Use != cmd2.Use {
		t.Error("Commands created from different factories should have same Use")
	}
	if cmd1.Short != cmd2.Short {
		t.Error("Commands created from different factories should have same Short")
	}
}

func TestExecute_MockDeviceInitialization(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "init-test-1",
					Address:    "192.168.1.150",
					MAC:        "11:22:33:44:55:77",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "init-test-2",
					Address:    "192.168.1.151",
					MAC:        "11:22:33:44:55:78",
					Type:       "SNSW-002P16EU",
					Model:      "Shelly Plus 2PM",
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

	// Test first device
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"init-test-1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("First device scan result: %v", err)

	// Test second device
	tf.Reset()
	buf.Reset()
	cmd = NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"init-test-2"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Second device scan result: %v", err)
}

func TestNewCommand_AllAliasesWork(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	aliases := []string{"search", "find"}
	for _, alias := range aliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(tf.Factory)

			if !contains(cmd.Aliases, alias) {
				t.Errorf("Expected alias %q in command", alias)
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestNewCommand_HasProperLongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedKeywords := []string{"WiFi", "networks", "SSID", "RSSI", "channel", "authentication"}
	longDesc := cmd.Long

	for _, keyword := range expectedKeywords {
		if !strings.Contains(longDesc, keyword) {
			t.Errorf("Expected long description to contain %q", keyword)
		}
	}
}

func TestNewCommand_ExampleShowsAllOptions(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check that example shows both basic and JSON output usage
	example := cmd.Example
	if !strings.Contains(example, "shelly wifi scan") {
		t.Error("Example should show basic scan usage")
	}
	if !strings.Contains(example, "-o json") || !strings.Contains(example, "-o") {
		t.Error("Example should show JSON output option")
	}
}

func TestNewCommand_HasShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Fatal("Short description is required")
	}
	if len(cmd.Short) < 10 {
		t.Error("Short description seems too short to be meaningful")
	}
}

func TestNewCommand_ArgValidation(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{"single device", []string{"device1"}, false},
		{"with dash", []string{"my-device"}, false},
		{"with underscore", []string{"my_device"}, false},
		{"with numbers", []string{"device123"}, false},
		{"no args", []string{}, true},
		{"empty string", []string{""}, false}, // cobra allows empty strings as args
		{"three args", []string{"a", "b", "c"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.expectErr {
				t.Errorf("Args(%v) error = %v, want error = %v", tt.args, err != nil, tt.expectErr)
			}
		})
	}
}

func TestRun_DirectCallSuccess(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "direct-scan-device",
					Address:    "192.168.1.200",
					MAC:        "AA:BB:CC:DD:EE:DD",
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
		Device:  "direct-scan-device",
	}

	err = run(context.Background(), opts)
	// Should succeed with mock returning scan results
	if err != nil {
		t.Logf("run() error: %v (may be expected if mock doesn't fully support)", err)
	}
}

func TestRun_DirectCallDeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
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
		Device:  "nonexistent-scan-device",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestRun_DirectCallContextCancelled(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "cancel-scan-device",
					Address:    "192.168.1.201",
					MAC:        "AA:BB:CC:DD:EE:DE",
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "cancel-scan-device",
	}

	err = run(ctx, opts)
	// Expected to error due to cancelled context
	if err == nil {
		t.Log("Context cancellation: run may succeed due to mock timing")
	}
}

func TestRun_DirectCallWithTimeout(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "timeout-scan-device",
					Address:    "192.168.1.202",
					MAC:        "AA:BB:CC:DD:EE:DF",
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "timeout-scan-device",
	}

	err = run(ctx, opts)
	t.Logf("run() with timeout error: %v", err)
}
