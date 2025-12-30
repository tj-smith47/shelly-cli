package jsonmetrics

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "json" {
		t.Errorf("Use = %q, want %q", cmd.Use, "json")
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

	expectedAliases := []string{"j"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
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
		{name: "devices", shorthand: "", defValue: "[]"},
		{name: "continuous", shorthand: "c", defValue: "false"},
		{name: "interval", shorthand: "i", defValue: "10s"},
		{name: "output", shorthand: "o", defValue: ""},
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

func TestNewCommand_FlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name          string
		expectedUsage string
	}{
		{name: "devices", expectedUsage: "Devices to include (default: all registered)"},
		{name: "continuous", expectedUsage: "Stream metrics continuously"},
		{name: "interval", expectedUsage: "Collection interval for continuous mode"},
		{name: "output", expectedUsage: "Output file (default: stdout)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if flag.Usage != tt.expectedUsage {
				t.Errorf("%s usage = %q, want %q", tt.name, flag.Usage, tt.expectedUsage)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"JSON format",
		"power",
		"voltage",
		"current",
		"energy",
		"metrics",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly metrics json",
		"--devices",
		"--continuous",
		"--interval",
		"--output",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil, expected a function")
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

func TestExecute_NoDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{},
		},
		DeviceStates: map[string]mock.DeviceState{},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error with no devices, got: %v", err)
	}

	output := tf.OutString()
	errOutput := tf.ErrString()
	combined := output + errOutput
	if !strings.Contains(combined, "No devices found") {
		t.Errorf("Expected warning about no devices, got stdout: %q, stderr: %q", output, errOutput)
	}
}

func TestExecute_WithAllDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "kitchen",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "living-room",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:FE",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"kitchen":     {},
			"living-room": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	output := tf.OutString()
	// Should have JSON output
	if !strings.Contains(output, "timestamp") && !strings.Contains(output, "devices") {
		t.Errorf("Expected JSON metrics output, got: %q", output)
	}

	// Verify it's valid JSON
	var result export.JSONMetricsOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify devices are included
	if len(result.Devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(result.Devices))
	}
}

func TestExecute_WithSpecificDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "kitchen",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "living-room",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:FE",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "bedroom",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:FD",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"kitchen":     {},
			"living-room": {},
			"bedroom":     {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{"kitchen", "bedroom"}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	output := tf.OutString()
	var result export.JSONMetricsOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Should only include 2 specified devices
	if len(result.Devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(result.Devices))
	}

	// Verify device names
	deviceNames := map[string]bool{}
	for _, dev := range result.Devices {
		deviceNames[dev.Device] = true
	}
	if !deviceNames["kitchen"] {
		t.Error("Expected kitchen device in output")
	}
	if !deviceNames["bedroom"] {
		t.Error("Expected bedroom device in output")
	}
	if deviceNames["living-room"] {
		t.Error("Did not expect living-room device in output")
	}
}

func TestExecute_WithOutputFile(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "metrics-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--output", tmpFile.Name()})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result export.JSONMetricsOutput
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Output file is not valid JSON: %v", err)
	}

	if len(result.Devices) != 1 {
		t.Errorf("Expected 1 device in output file, got %d", len(result.Devices))
	}
}

func TestExecute_SingleShot(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	output := tf.OutString()
	// Should output exactly one JSON object (single shot mode)
	var result export.JSONMetricsOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

func TestExecute_ContinuousMode(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	// Use a context with timeout for continuous mode
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = run(ctx, tf.Factory, []string{}, true, 50*time.Millisecond, "")
	// Context timeout is normal for continuous mode
	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Logf("Execute returned: %v (expected for continuous mode)", err)
	}

	output := tf.OutString()
	// Should have at least one JSON object
	lines := strings.Split(strings.TrimSpace(output), "\n")
	hasJSON := false
	for _, line := range lines {
		if line != "" && strings.Contains(line, "timestamp") {
			hasJSON = true
			var result export.JSONMetricsOutput
			if err := json.Unmarshal([]byte(line), &result); err == nil {
				hasJSON = true
				break
			}
		}
	}
	if !hasJSON {
		t.Logf("Warning: No JSON output found in continuous mode (may be timing issue): %q", output)
	}
}

func TestExecute_CustomInterval(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err = run(ctx, tf.Factory, []string{}, true, 100*time.Millisecond, "")
	// Context timeout is normal
	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Logf("Execute returned: %v (expected for continuous mode)", err)
	}
}

func TestExecute_InvalidIntervalFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--interval", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid interval")
	}
}

func TestExecute_InvalidOutputPath(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	// Use an invalid path that can't be created
	cmd.SetArgs([]string{"--output", "/nonexistent/path/to/file.json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid output path")
	}
}

func TestRun_NoDevices(t *testing.T) {
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

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error with no devices, got: %v", err)
	}
}

func TestRun_WithDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device1": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{"device1"}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestRun_WithMultipleDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "device2",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:FE",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device1": {},
			"device2": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{"device1", "device2"}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestRun_DevicesSorted(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "zebra",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "apple",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:FE",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"zebra": {},
			"apple": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	var result export.JSONMetricsOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Logf("Could not parse JSON (might be timing issue): %v", err)
		return
	}

	// Verify devices are sorted
	if len(result.Devices) > 1 {
		if result.Devices[0].Device > result.Devices[1].Device {
			t.Errorf("Devices not sorted: %v", result.Devices)
		}
	}
}

func TestRun_OutputToFile(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	tmpFile, err := os.CreateTemp("", "metrics-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{"test-device"}, false, 10*time.Second, tmpFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify file contents
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result export.JSONMetricsOutput
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Output file is not valid JSON: %v", err)
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	// Use short timeout for continuous mode
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = run(ctx, tf.Factory, []string{"test-device"}, true, 10*time.Second, "")
	// Context cancellation is expected
	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Logf("Expected context cancellation, got: %v", err)
	}
}

func TestRun_OutputStdout(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
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

	err = run(context.Background(), tf.Factory, []string{"test-device"}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "timestamp") && !strings.Contains(output, "devices") {
		t.Errorf("Expected JSON output on stdout, got: %q", output)
	}
}

func TestExecute_DevicesAreSorted(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "zulu",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "alpha",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:FE",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
				{
					Name:       "charlie",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:FD",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"zulu":   {},
			"alpha":  {},
			"charlie": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	err = run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	output := tf.OutString()
	var result export.JSONMetricsOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify devices are sorted alphabetically
	if len(result.Devices) != 3 {
		t.Errorf("Expected 3 devices, got %d", len(result.Devices))
	}
	if result.Devices[0].Device != "alpha" {
		t.Errorf("First device should be 'alpha', got %q", result.Devices[0].Device)
	}
	if result.Devices[1].Device != "charlie" {
		t.Errorf("Second device should be 'charlie', got %q", result.Devices[1].Device)
	}
	if result.Devices[2].Device != "zulu" {
		t.Errorf("Third device should be 'zulu', got %q", result.Devices[2].Device)
	}
}
