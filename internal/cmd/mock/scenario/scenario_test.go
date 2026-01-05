package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

const testConfigDir = "/test/config"

// setupTestEnv sets up an isolated environment for tests using afero.
func setupTestEnv(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", testConfigDir)
	return fs
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

	if cmd.Use != "scenario <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "scenario <name>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"load", "setup"}
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
	if err := cmd.Args(cmd, []string{"home"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}

	// Should reject multiple arguments
	if err := cmd.Args(cmd, []string{"home", "office"}); err == nil {
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

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_UnknownScenario(t *testing.T) {
	_ = setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "unknown-scenario"}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for unknown scenario")
	}

	// Should mention available scenarios
	if err != nil && !strings.Contains(err.Error(), "unknown scenario") {
		t.Errorf("error should mention unknown scenario, got: %v", err)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_MinimalScenario(t *testing.T) {
	fs := setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "minimal"}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have created device file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	// Check that device file was created
	deviceFile := mockDir + "/test-switch.json"
	if _, err := fs.Stat(deviceFile); err != nil {
		t.Errorf("expected device file to be created: %s", deviceFile)
	}

	output := stdout.String() + stderr.String()
	// Should mention test-switch
	if !strings.Contains(output, "test-switch") {
		t.Errorf("output should mention test-switch, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_HomeScenario(t *testing.T) {
	fs := setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "home"}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have created 3 device files for home scenario
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	expectedDevices := []string{"living-room.json", "bedroom.json", "kitchen.json"}
	for _, device := range expectedDevices {
		deviceFile := mockDir + "/" + device
		if _, err := fs.Stat(deviceFile); err != nil {
			t.Errorf("expected device file to be created: %s", deviceFile)
		}
	}

	output := stdout.String() + stderr.String()
	// Should mention all devices
	for _, name := range []string{"living-room", "bedroom", "kitchen"} {
		if !strings.Contains(output, name) {
			t.Errorf("output should mention %s, got: %s", name, output)
		}
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_OfficeScenario(t *testing.T) {
	fs := setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "office"}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have created 5 device files for office scenario
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	expectedDevices := []string{
		"desk-lamp.json",
		"monitor.json",
		"printer.json",
		"air-purifier.json",
		"heater.json",
	}
	for _, device := range expectedDevices {
		deviceFile := mockDir + "/" + device
		if _, err := fs.Stat(deviceFile); err != nil {
			t.Errorf("expected device file to be created: %s", deviceFile)
		}
	}

	output := stdout.String() + stderr.String()
	// Should mention success and device count
	if !strings.Contains(output, "5") {
		t.Errorf("output should mention 5 devices created, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_ScenarioCreatesValidJSON(t *testing.T) {
	fs := setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "minimal"}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read and validate the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := mockDir + "/test-switch.json"
	data, err := afero.ReadFile(fs, deviceFile)
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Errorf("created file is not valid JSON: %v", err)
	}

	if device.Name != "test-switch" {
		t.Errorf("device name = %q, want %q", device.Name, "test-switch")
	}

	if device.Model != "Plus 1PM" {
		t.Errorf("device model = %q, want %q", device.Model, "Plus 1PM")
	}

	if device.Firmware != "1.0.0" {
		t.Errorf("device firmware = %q, want %q", device.Firmware, "1.0.0")
	}

	// MAC should be set
	if device.MAC == "" {
		t.Error("device MAC should not be empty")
	}

	// State should be set
	if device.State == nil {
		t.Error("device State should not be nil")
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_DeviceHasInitialState(t *testing.T) {
	fs := setupTestEnv(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Scenario: "minimal"}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read the created file
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("failed to get mock dir: %v", err)
	}

	deviceFile := mockDir + "/test-switch.json"
	data, err := afero.ReadFile(fs, deviceFile)
	if err != nil {
		t.Fatalf("failed to read device file: %v", err)
	}

	var device mock.Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("failed to parse device file: %v", err)
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
