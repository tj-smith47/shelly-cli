package list

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

// setupTestEnv sets up an isolated environment for tests.
func setupTestEnv(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
}

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

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"ls"}
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

	// Should accept no arguments
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("unexpected error with no args: %v", err)
	}

	// Should reject any arguments
	if err := cmd.Args(cmd, []string{"arg1"}); err == nil {
		t.Error("expected error with 1 arg")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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
func TestRun_NoDevices(t *testing.T) {
	setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String() + stderr.String()
	if output == "" {
		t.Error("expected output for no devices case")
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_WithDevices(t *testing.T) {
	setupTestEnv(t)

	// Create mock dir and device
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	device := mock.Device{
		Name:     "test-device",
		Model:    "Plus 1PM",
		Firmware: "1.0.0",
		MAC:      "AA:BB:CC:DD:EE:FF",
		State:    map[string]interface{}{"switch:0": map[string]interface{}{"output": false}},
	}

	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal device: %v", err)
	}

	filename := filepath.Join(mockDir, "test-device.json")
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatalf("failed to write device file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("expected output listing devices")
	}

	// Should contain device name
	if !bytes.Contains([]byte(output), []byte("test-device")) {
		t.Errorf("output should contain device name, got: %s", output)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_SkipsNonJSONFiles(t *testing.T) {
	setupTestEnv(t)

	// Create mock dir
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	// Create a non-JSON file
	txtFile := filepath.Join(mockDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("This is not a device"), 0o600); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}

	// Create a valid device file
	device := mock.Device{
		Name:     "valid-device",
		Model:    "Plus 2PM",
		Firmware: "1.0.0",
		MAC:      "AA:BB:CC:11:22:33",
	}

	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal device: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "valid-device.json")
	if err := os.WriteFile(deviceFile, data, 0o600); err != nil {
		t.Fatalf("failed to write device file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should contain valid device
	if !bytes.Contains([]byte(output), []byte("valid-device")) {
		t.Errorf("output should contain valid-device, got: %s", output)
	}
	// Should NOT contain readme.txt
	if bytes.Contains([]byte(output), []byte("readme")) {
		t.Errorf("output should NOT contain readme, got: %s", output)
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_SkipsDirectories(t *testing.T) {
	setupTestEnv(t)

	// Create mock dir
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(mockDir, "subdir")
	if err := os.MkdirAll(subDir, 0o700); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should not error even with subdirectory present
	output := stdout.String() + stderr.String()
	if output == "" {
		t.Error("expected some output")
	}
}

//nolint:paralleltest // Manipulates XDG_CONFIG_HOME environment variable
func TestRun_SkipsInvalidJSON(t *testing.T) {
	setupTestEnv(t)

	// Create mock dir
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	// Create an invalid JSON file
	invalidFile := filepath.Join(mockDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("{ not valid json "), 0o600); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	// Create a valid device file
	device := mock.Device{
		Name:     "valid-device",
		Model:    "Plus 2PM",
		Firmware: "1.0.0",
		MAC:      "AA:BB:CC:11:22:33",
	}

	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal device: %v", err)
	}

	deviceFile := filepath.Join(mockDir, "valid-device.json")
	if err := os.WriteFile(deviceFile, data, 0o600); err != nil {
		t.Fatalf("failed to write device file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should contain valid device but not crash on invalid
	if !bytes.Contains([]byte(output), []byte("valid-device")) {
		t.Errorf("output should contain valid-device, got: %s", output)
	}
}
