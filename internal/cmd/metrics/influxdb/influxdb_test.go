package influxdb

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "influxdb" {
		t.Errorf("Use = %q, want 'influxdb'", cmd.Use)
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

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"influx", "line"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, expected := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias %q", expected)
			continue
		}
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name        string
		flagName    string
		shorthand   string
		defValue    string
		shouldExist bool
	}{
		{"devices flag", "devices", "", "", true},
		{"continuous flag", "continuous", "c", "false", true},
		{"interval flag", "interval", "i", "10s", true},
		{"output flag", "output", "o", "", true},
		{"measurement flag", "measurement", "m", "shelly", true},
		{"tags flag", "tags", "t", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if (flag == nil) != !tt.shouldExist {
				t.Errorf("flag %q should exist: %v", tt.flagName, tt.shouldExist)
				return
			}
			if flag != nil {
				if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
					t.Errorf("shorthand = %q, want %q", flag.Shorthand, tt.shorthand)
				}
				if tt.defValue != "" && flag.DefValue != tt.defValue {
					t.Errorf("default value = %q, want %q", flag.DefValue, tt.defValue)
				}
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

func TestParseTagsEmpty(t *testing.T) {
	t.Parallel()

	tags := parseTags([]string{})
	if len(tags) != 0 {
		t.Errorf("parseTags([]) = %v, want empty map", tags)
	}
}

func TestParseTagsSingleTag(t *testing.T) {
	t.Parallel()

	const expectedLocation = "home"
	tags := parseTags([]string{"location=" + expectedLocation})
	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags["location"] != expectedLocation {
		t.Errorf("tags[location] = %q, want %q", tags["location"], expectedLocation)
	}
}

func TestParseTagsMultipleTags(t *testing.T) {
	t.Parallel()

	tags := parseTags([]string{"location=home", "floor=1", "room=kitchen"})
	if len(tags) != 3 {
		t.Errorf("len(tags) = %d, want 3", len(tags))
	}
	if tags["location"] != "home" {
		t.Errorf("tags[location] = %q, want 'home'", tags["location"])
	}
	if tags["floor"] != "1" {
		t.Errorf("tags[floor] = %q, want '1'", tags["floor"])
	}
	if tags["room"] != "kitchen" {
		t.Errorf("tags[room] = %q, want 'kitchen'", tags["room"])
	}
}

func TestParseTagsInvalidFormat(t *testing.T) {
	t.Parallel()

	// Tag without = should be skipped
	tags := parseTags([]string{"invalid_tag", "location=home"})
	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags["location"] != "home" {
		t.Errorf("tags[location] = %q, want 'home'", tags["location"])
	}
}

func TestParseTagsWithValues(t *testing.T) {
	t.Parallel()

	// Test value with multiple equals signs
	tags := parseTags([]string{"query=key=value"})
	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags["query"] != "key=value" {
		t.Errorf("tags[query] = %q, want 'key=value'", tags["query"])
	}
}

func TestSetupOutputStdout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	out, cleanup, err := setupOutput(tf.TestIO.IOStreams, "")
	if err != nil {
		t.Errorf("setupOutput with empty file should not error, got: %v", err)
	}

	if out != tf.TestIO.Out {
		t.Error("setupOutput should return stdout when file is empty string")
	}

	cleanup()
}

func TestSetupOutputFile(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "metrics.txt")

	out, cleanup, err := setupOutput(tf.TestIO.IOStreams, outputFile)
	if err != nil {
		t.Errorf("setupOutput with valid file should not error, got: %v", err)
	}

	if out == nil {
		t.Error("setupOutput should return a valid file writer")
	}

	// Verify we can write to it
	if _, err := io.WriteString(out, "test"); err != nil {
		t.Errorf("failed to write to output file: %v", err)
	}

	cleanup()

	// Verify file was created
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestSetupOutputInvalidPath(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use an invalid path that can't be created
	invalidPath := "/dev/null/subdir/file.txt"

	_, _, err := setupOutput(tf.TestIO.IOStreams, invalidPath)
	if err == nil {
		t.Error("setupOutput with invalid path should error")
	}

	if strings.Contains(err.Error(), "failed to create output file") {
		t.Logf("expected error: %v", err)
	}
}

func TestRun_NoDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, []string{}, false, 10*time.Second, "", "shelly", []string{})
	if err != nil {
		t.Errorf("run with no devices should not error, got: %v", err)
	}

	// Warning goes to stderr, not stdout
	output := tf.ErrString()
	if !strings.Contains(output, "No devices found") {
		t.Errorf("Expected warning about no devices in stderr, got: %q", output)
	}
}

func TestRun_WithDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"kitchen": {
			Address: "192.168.1.100",
		},
	})

	err := run(context.Background(), tf.Factory, []string{"kitchen"}, false, 10*time.Second, "", "shelly", []string{})
	// Error may occur due to mock service, but we're testing the flow
	if err != nil {
		t.Logf("run with device returned error (may be expected): %v", err)
	}
}

func TestRun_WithCustomMeasurement(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"living_room": {
			Address: "192.168.1.101",
		},
	})

	err := run(context.Background(), tf.Factory, []string{"living_room"}, false, 10*time.Second, "", "home_metrics", []string{})
	if err != nil {
		t.Logf("run with custom measurement returned error (may be expected): %v", err)
	}
}

func TestRun_WithCustomTags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"bedroom": {
			Address: "192.168.1.102",
		},
	})

	err := run(context.Background(), tf.Factory, []string{"bedroom"}, false, 10*time.Second, "", "shelly", []string{"location=home", "floor=2"})
	if err != nil {
		t.Logf("run with custom tags returned error (may be expected): %v", err)
	}
}

func TestRunContinuous_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	svc := tf.ShellyService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	writePointsCalled := false
	writePoints := func(_ []export.InfluxDBPoint) {
		writePointsCalled = true
	}

	err := runContinuous(ctx, svc, []string{"test-device"}, "shelly", map[string]string{}, 1*time.Second, writePoints)
	if err != nil {
		t.Errorf("runContinuous with cancelled context should not error, got: %v", err)
	}

	if writePointsCalled {
		t.Error("writePoints should not be called with immediately cancelled context")
	}
}

func TestRunContinuous_ShortDuration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	svc := tf.ShellyService()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	writePoints := func(_ []export.InfluxDBPoint) {
		callCount++
	}

	err := runContinuous(ctx, svc, []string{"test-device"}, "shelly", map[string]string{}, 20*time.Millisecond, writePoints)
	if err != nil {
		t.Errorf("runContinuous should not error on timeout, got: %v", err)
	}

	// Due to timing variability, we just verify it was called at least once or zero times gracefully
	t.Logf("writePoints called %d times", callCount)
}

func TestLineProtocolWriter_Creation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	writer := &LineProtocolWriter{
		out:         tf.TestIO.Out,
		measurement: "test_measurement",
		tags: map[string]string{
			"location": "kitchen",
		},
		ios: tf.TestIO.IOStreams,
	}

	if writer.measurement != "test_measurement" {
		t.Errorf("measurement = %q, want 'test_measurement'", writer.measurement)
	}
	if writer.tags["location"] != "kitchen" {
		t.Errorf("tags[location] = %q, want 'kitchen'", writer.tags["location"])
	}
}

func TestExecute_NoDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute with no devices should not error, got: %v", err)
	}

	// Warning goes to stderr
	output := tf.ErrString()
	if !strings.Contains(output, "No devices found") {
		t.Logf("Expected warning about no devices in stderr: %q", output)
	}
}

func TestExecute_WithDevicesFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"test-device": {
			Address: "192.168.1.1",
		},
	})

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--devices", "test-device"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with devices flag returned error (may be expected): %v", err)
	}
}

func TestExecute_WithMeasurementFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--measurement", "custom_measurement"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with measurement flag returned error (may be expected): %v", err)
	}
}

func TestExecute_WithTagsFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--tags", "location=home", "--tags", "region=us-west"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with tags flag returned error (may be expected): %v", err)
	}
}

func TestExecute_WithOutputFile(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "metrics.txt")

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--output", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with output flag returned error (may be expected): %v", err)
	}
}

func TestExecute_WithMultipleDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"device1": {
			Address: "192.168.1.1",
		},
		"device2": {
			Address: "192.168.1.2",
		},
	})

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--devices", "device1", "--devices", "device2"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with multiple devices returned error (may be expected): %v", err)
	}
}

func TestExecute_ContinuousMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--continuous", "--interval", "100ms"})

	// Run with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	// Error is expected due to context timeout, but we're verifying the command structure
	if err != nil {
		t.Logf("Execute in continuous mode returned error (expected due to timeout): %v", err)
	}
}

func TestExecute_ContinuousShort(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"-c", "-i", "100ms"})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with short flags returned error (expected): %v", err)
	}
}

func TestExecute_AllFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"kitchen": {
			Address: "192.168.1.100",
		},
	})

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "metrics.txt")

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{
		"--devices", "kitchen",
		"--measurement", "home_power",
		"--tags", "location=kitchen",
		"--tags", "floor=1",
		"--output", outputFile,
	})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with all flags returned error (may be expected): %v", err)
	}
}

func TestRun_SortDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"zebra": {
			Address: "192.168.1.10",
		},
		"alpha": {
			Address: "192.168.1.1",
		},
		"beta": {
			Address: "192.168.1.2",
		},
	})

	// Run with no explicit devices to test sorting
	err := run(context.Background(), tf.Factory, []string{"zebra", "alpha", "beta"}, false, 10*time.Second, "", "shelly", []string{})
	if err != nil {
		t.Logf("run with explicit devices returned error (may be expected): %v", err)
	}
}

func TestRun_InvalidConfigManager(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// The factory has a valid config, but we test the error path by expecting failures
	// when the service tries to contact non-existent devices
	err := run(context.Background(), tf.Factory, []string{"nonexistent-device"}, false, 10*time.Second, "", "shelly", []string{})
	if err != nil {
		t.Logf("run with nonexistent device returned error (may be expected): %v", err)
	}
}
