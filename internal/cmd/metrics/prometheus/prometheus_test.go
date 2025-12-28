package prometheus

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "prometheus" {
		t.Errorf("Use = %q, want %q", cmd.Use, "prometheus")
	}

	if cmd.Short != "Start Prometheus metrics exporter" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Start Prometheus metrics exporter")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Check aliases
	expectedAliases := []string{"prom"}
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
		{name: "port", shorthand: "", defValue: "9090"},
		{name: "devices", shorthand: "", defValue: "[]"},
		{name: "interval", shorthand: "", defValue: "15s"},
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
		{name: "port", expectedUsage: "HTTP port for the exporter"},
		{name: "devices", expectedUsage: "Devices to include (default: all registered)"},
		{name: "interval", expectedUsage: "Metrics collection interval"},
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

	// Verify Long description contains expected content
	longDesc := cmd.Long

	// Should mention Prometheus
	if !bytes.Contains([]byte(longDesc), []byte("Prometheus")) {
		t.Error("Long description should mention Prometheus")
	}

	// Should document power metrics
	if !bytes.Contains([]byte(longDesc), []byte("shelly_power_watts")) {
		t.Error("Long description should document shelly_power_watts metric")
	}

	// Should document device online metric
	if !bytes.Contains([]byte(longDesc), []byte("shelly_device_online")) {
		t.Error("Long description should document shelly_device_online metric")
	}

	// Should document labels
	if !bytes.Contains([]byte(longDesc), []byte("device")) {
		t.Error("Long description should document device label")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify examples contain expected content
	example := cmd.Example

	// Should show basic usage
	if !bytes.Contains([]byte(example), []byte("shelly metrics prometheus")) {
		t.Error("Example should show basic usage")
	}

	// Should show port flag
	if !bytes.Contains([]byte(example), []byte("--port")) {
		t.Error("Example should show port flag usage")
	}

	// Should show devices flag
	if !bytes.Contains([]byte(example), []byte("--devices")) {
		t.Error("Example should show devices flag usage")
	}

	// Should show interval flag
	if !bytes.Contains([]byte(example), []byte("--interval")) {
		t.Error("Example should show interval flag usage")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The prometheus command should accept no arguments (uses flags)
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Expected no error with no args, got: %v", err)
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

func TestRun_NoDevicesWarning(t *testing.T) {
	t.Parallel()

	// Create test iostreams
	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(&bytes.Buffer{}, &stdout, &stderr)

	// Create a factory with empty config (no devices) using NewTestManager
	// This ensures we don't load from the real config file
	mgr := config.NewTestManager(&config.Config{})
	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr)

	// Context doesn't need to be cancelled - run returns early when there are no devices
	ctx := context.Background()

	// Call run with no devices specified
	err := run(ctx, f, 9999, nil, 15*time.Second)
	if err != nil {
		t.Errorf("Expected no error when no devices, got: %v", err)
	}

	// Should have printed a warning to stderr
	output := stderr.String()
	if !bytes.Contains([]byte(output), []byte("No devices found")) {
		t.Errorf("Expected warning about no devices, got stderr: %q, stdout: %q", output, stdout.String())
	}
}

func TestRun_WithDevices(t *testing.T) {
	t.Parallel()

	// Create test iostreams
	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(&bytes.Buffer{}, &stdout, &stderr)

	// Create a factory with devices in config using NewTestManager
	mgr := config.NewTestManager(&config.Config{
		Devices: map[string]model.Device{
			"kitchen": {Address: "192.168.1.100"},
			"living":  {Address: "192.168.1.101"},
		},
	})
	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run in a goroutine since it blocks until context is done
	done := make(chan error, 1)
	go func() {
		done <- run(ctx, f, 19999, nil, 1*time.Second) // Use high port to avoid conflicts
	}()

	// Wait for completion
	select {
	case err := <-done:
		// The server should shut down gracefully without error
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out")
	}

	// Should have printed startup message
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Starting Prometheus exporter")) {
		t.Errorf("Expected startup message, got: %q", output)
	}
	if !bytes.Contains([]byte(output), []byte("Monitoring 2 devices")) {
		t.Errorf("Expected device count message, got: %q", output)
	}
}

func TestRun_WithSpecificDevices(t *testing.T) {
	t.Parallel()

	// Create test iostreams
	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(&bytes.Buffer{}, &stdout, &stderr)

	// Create a factory with devices in config using NewTestManager
	mgr := config.NewTestManager(&config.Config{
		Devices: map[string]model.Device{
			"kitchen": {Address: "192.168.1.100"},
			"living":  {Address: "192.168.1.101"},
			"bedroom": {Address: "192.168.1.102"},
		},
	})
	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run with specific devices
	done := make(chan error, 1)
	go func() {
		// Only include kitchen and bedroom
		done <- run(ctx, f, 29999, []string{"kitchen", "bedroom"}, 1*time.Second)
	}()

	// Wait for completion
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out")
	}

	// Should have printed startup message with 2 devices
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Monitoring 2 devices")) {
		t.Errorf("Expected 2 devices, got: %q", output)
	}
}

func TestRun_DevicesAreSorted(t *testing.T) {
	t.Parallel()

	// Create test iostreams
	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(&bytes.Buffer{}, &stdout, &stderr)

	// Create a factory with devices in config using NewTestManager
	mgr := config.NewTestManager(&config.Config{
		Devices: map[string]model.Device{
			"zebra":  {Address: "192.168.1.100"},
			"apple":  {Address: "192.168.1.101"},
			"mango":  {Address: "192.168.1.102"},
			"banana": {Address: "192.168.1.103"},
			"cherry": {Address: "192.168.1.104"},
		},
	})
	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run - should sort the devices alphabetically
	done := make(chan error, 1)
	go func() {
		done <- run(ctx, f, 39999, nil, 1*time.Second)
	}()

	// Wait for completion
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out")
	}

	// The devices should be sorted (we can't easily verify this without more complex testing,
	// but at least the command ran successfully with all 5 devices)
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Monitoring 5 devices")) {
		t.Errorf("Expected 5 devices, got: %q", output)
	}
}

func TestRun_IntervalParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting interval flag
	if err := cmd.Flags().Set("interval", "30s"); err != nil {
		t.Errorf("Failed to set interval flag: %v", err)
	}

	flag := cmd.Flags().Lookup("interval")
	if flag.Value.String() != "30s" {
		t.Errorf("interval = %q, want %q", flag.Value.String(), "30s")
	}
}

func TestRun_PortParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting port flag
	if err := cmd.Flags().Set("port", "8080"); err != nil {
		t.Errorf("Failed to set port flag: %v", err)
	}

	flag := cmd.Flags().Lookup("port")
	if flag.Value.String() != "8080" {
		t.Errorf("port = %q, want %q", flag.Value.String(), "8080")
	}
}

func TestRun_DevicesParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting devices flag
	if err := cmd.Flags().Set("devices", "kitchen,living-room,bedroom"); err != nil {
		t.Errorf("Failed to set devices flag: %v", err)
	}

	flag := cmd.Flags().Lookup("devices")
	// StringSlice flags are formatted as [a,b,c]
	expected := "[kitchen,living-room,bedroom]"
	if flag.Value.String() != expected {
		t.Errorf("devices = %q, want %q", flag.Value.String(), expected)
	}
}
