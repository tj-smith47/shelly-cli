package create

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
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

	if cmd.Use != "create <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <name>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"add", "new"}
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

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should reject no arguments
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}

	// Should accept exactly one argument
	if err := cmd.Args(cmd, []string{"test-device"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}

	// Should reject multiple arguments
	if err := cmd.Args(cmd, []string{"device1", "device2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check model flag exists
	modelFlag := cmd.Flags().Lookup("model")
	if modelFlag == nil {
		t.Error("model flag is missing")
	} else if modelFlag.DefValue != "Plus 1PM" {
		t.Errorf("model default = %q, want %q", modelFlag.DefValue, "Plus 1PM")
	}

	// Check firmware flag exists
	firmwareFlag := cmd.Flags().Lookup("firmware")
	if firmwareFlag == nil {
		t.Error("firmware flag is missing")
	} else if firmwareFlag.DefValue != "1.0.0" {
		t.Errorf("firmware default = %q, want %q", firmwareFlag.DefValue, "1.0.0")
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_CreateDevice(t *testing.T) {
	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "my-device", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have created device file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "my-device.json")
	if _, err := os.Stat(deviceFile); os.IsNotExist(err) {
		t.Errorf("expected device file to be created: %s", deviceFile)
	}

	output := stdout.String() + stderr.String()
	// Should mention device name
	if !bytes.Contains([]byte(output), []byte("my-device")) {
		t.Errorf("output should mention my-device, got: %s", output)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_CustomModel(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "custom-device", "Plus 2PM", "1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read and verify the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "custom-device.json")
	data, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
	}

	if device.Model != "Plus 2PM" {
		t.Errorf("device model = %q, want %q", device.Model, "Plus 2PM")
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_CustomFirmware(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "fw-device", "Plus 1PM", "2.0.5")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read and verify the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "fw-device.json")
	data, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
	}

	if device.Firmware != "2.0.5" {
		t.Errorf("device firmware = %q, want %q", device.Firmware, "2.0.5")
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_DeviceHasMAC(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "mac-device", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read and verify the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "mac-device.json")
	data, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
	}

	// MAC should be set and match expected format
	if device.MAC == "" {
		t.Error("device MAC should not be empty")
	}

	// MAC should start with AA:BB:CC: (as per GenerateMAC)
	if len(device.MAC) < 9 || device.MAC[:9] != "AA:BB:CC:" {
		t.Errorf("device MAC = %q, should start with AA:BB:CC:", device.MAC)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_DeviceHasInitialState(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "state-device", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read and verify the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "state-device.json")
	data, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
	}

	// State should be set
	if device.State == nil {
		t.Error("device State should not be nil")
	}

	// Check switch:0 state exists
	switchState, ok := device.State["switch:0"]
	if !ok {
		t.Error("device State should contain switch:0")
	}

	// Check switch state values
	switchMap, ok := switchState.(map[string]interface{})
	if !ok {
		t.Error("switch:0 should be a map")
		return
	}

	if output, ok := switchMap["output"]; !ok || output != false {
		t.Errorf("switch:0.output = %v, want false", output)
	}

	if apower, ok := switchMap["apower"]; !ok || apower != 0.0 {
		t.Errorf("switch:0.apower = %v, want 0.0", apower)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_OutputContainsModelAndFirmware(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "output-device", "Plug S", "3.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()

	// Should display model
	if !bytes.Contains([]byte(output), []byte("Plug S")) {
		t.Errorf("output should contain model 'Plug S', got: %s", output)
	}

	// Should display firmware
	if !bytes.Contains([]byte(output), []byte("3.0.0")) {
		t.Errorf("output should contain firmware '3.0.0', got: %s", output)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_FilePermissions(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f, "perms-device", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check file permissions
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "perms-device.json")
	info, err := os.Stat(deviceFile)
	if err != nil {
		t.Fatalf("failed to stat device file: %v", err)
	}

	// Should have 0600 permissions (owner read/write only)
	mode := info.Mode().Perm()
	if mode != 0o600 {
		t.Errorf("file permissions = %o, want %o", mode, 0o600)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_DeterministicMAC(t *testing.T) {
	// This test is NOT parallel due to XDG_CONFIG_HOME environment variable manipulation

	// Create temp directory for mock config
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to control config directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
			t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create same device twice
	err := run(context.Background(), f, "deterministic", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	// Read first MAC
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "deterministic.json")
	data1, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device1 mock.Device
	if err := json.Unmarshal(data1, &device1); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
	}

	// Create again (overwrites)
	stdout.Reset()
	stderr.Reset()
	err = run(context.Background(), f, "deterministic", "Plus 1PM", "1.0.0")
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	// Read second MAC
	data2, err := os.ReadFile(deviceFile) //nolint:gosec // Test file path from temp dir
	if err != nil {
		t.Fatalf("failed to read device file second time: %v", err)
	}

	var device2 mock.Device
	if err := json.Unmarshal(data2, &device2); err != nil {
		t.Fatalf("failed to parse device file second time: %v", err)
	}

	// MACs should be identical (deterministic based on name)
	if device1.MAC != device2.MAC {
		t.Errorf("MACs should be deterministic: first = %q, second = %q", device1.MAC, device2.MAC)
	}
}
