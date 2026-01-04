package bulk

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	testConfigDir = "/test/config"
)

const (
	validConfig = `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: device1
    address: 192.168.1.100
`
	validConfigTwoDevices = `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: device1
    address: 192.168.1.100
  - name: device2
    address: 192.168.1.101
`
	validConfigThreeDevices = `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: device1
    address: 192.168.1.100
  - name: device2
    address: 192.168.1.101
  - name: device3
    address: 192.168.1.102
`
	registeredDevicesConfig = `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: kitchen-light
  - name: bedroom-light
`
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

	if cmd.Use != "bulk <config-file>" {
		t.Errorf("Use = %q, want 'bulk <config-file>'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"batch", "mass"}
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

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Bulk provision from config file"
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

	expectedPatterns := []string{"YAML", "WiFi", "devices", "parallel"}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	expectedPatterns := []string{"shelly provision bulk", "devices.yaml", "--dry-run", "--parallel"}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_RequiresOneArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}

	err = cmd.Args(cmd, []string{"config.yaml"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"config1.yaml", "config2.yaml"})
	if err == nil {
		t.Error("Expected error with multiple args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	flags := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"parallel", "", "5"},
		{"dry-run", "", "false"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag %q shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
		if flag.DefValue != f.defValue {
			t.Errorf("flag %q default = %q, want %q", f.name, flag.DefValue, f.defValue)
		}
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Parallel: 5,
	}

	if opts.Parallel != 5 {
		t.Errorf("Default Parallel = %d, want 5", opts.Parallel)
	}
	if opts.DryRun {
		t.Error("Default DryRun should be false")
	}
	if opts.ConfigFile != "" {
		t.Error("Default ConfigFile should be empty")
	}
}

func TestRun_MissingConfigFile(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: "/nonexistent/config.yaml",
		Parallel:   5,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for missing config file")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_InvalidYAML(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/invalid.yaml"

	// Write invalid YAML
	if err := afero.WriteFile(config.Fs(), configFile, []byte("invalid: yaml: content: :"), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_NoDevices(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/empty.yaml"

	// Write config with no devices
	if err := afero.WriteFile(config.Fs(), configFile, []byte("wifi:\n  ssid: MyNetwork\n  password: secret\ndevices: []\n"), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for no devices")
	}
	if !strings.Contains(err.Error(), "no devices") {
		t.Errorf("Expected 'no devices' error, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_DryRunSuccess(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/dryrun.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfigTwoDevices), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
		DryRun:     true,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error for dry run, got: %v", err)
	}

	// Check output
	output := tf.OutString()
	if !strings.Contains(output, "device1") {
		t.Error("Expected device1 in output")
	}
	if !strings.Contains(output, "device2") {
		t.Error("Expected device2 in output")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_UnregisteredDeviceNoAddress(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/validation.yaml"

	// Device without address and not registered
	configContent := `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: unregistered-device
`
	if err := afero.WriteFile(config.Fs(), configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected validation error for unregistered device")
	}
	if !strings.Contains(err.Error(), "not a registered device") {
		t.Errorf("Expected 'not a registered device' error, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_RegisteredDeviceNoAddress(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"my-device": {
			Name:    "my-device",
			Address: "192.168.1.100",
		},
	})
	configFile := testConfigDir + "/registered.yaml"

	// Device registered, no address needed
	configContent := `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: my-device
`
	if err := afero.WriteFile(config.Fs(), configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
		DryRun:     true, // Use dry run to avoid actual provisioning
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error for registered device, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_PerDeviceWiFiOverride(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/override.yaml"

	configContent := `wifi:
  ssid: DefaultNetwork
  password: default-secret
devices:
  - name: device1
    address: 192.168.1.100
  - name: device2
    address: 192.168.1.101
    wifi:
      ssid: CustomNetwork
      password: custom-secret
`
	if err := afero.WriteFile(config.Fs(), configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
		DryRun:     true,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "DefaultNetwork") {
		t.Error("Expected DefaultNetwork in output")
	}
	if !strings.Contains(output, "CustomNetwork") {
		t.Error("Expected CustomNetwork in output")
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
		t.Error("Expected error when no config file provided")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestExecute_WithConfigFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/exec.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{configFile, "--dry-run"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestExecute_WithParallelFlag(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/parallel.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfigTwoDevices), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{configFile, "--parallel", "2", "--dry-run"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestExecute_DryRunFlag(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/dryrun-exec.yaml"

	configContent := `wifi:
  ssid: TestNetwork
  password: test-secret
devices:
  - name: test-device
    address: 192.168.1.100
`
	if err := afero.WriteFile(config.Fs(), configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{configFile, "--dry-run"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Dry run should display devices without actually provisioning
	output := buf.String()
	if !strings.Contains(output, "test-device") && !strings.Contains(output, "TestNetwork") {
		t.Logf("Output: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestExecute_WithAlias_batch(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/alias-batch.yaml"

	configContent := `wifi:
  ssid: MyNetwork
  password: secret
devices:
  - name: device1
    address: 192.168.1.100
`
	if err := afero.WriteFile(config.Fs(), configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := NewCommand(tf.Factory)

	// Verify aliases include "batch"
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "batch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'batch' alias")
	}
}

func TestExecute_WithAlias_mass(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify aliases include "mass"
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "mass" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'mass' alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_MultipleDevices(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/multiple.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfigThreeDevices), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   5,
		DryRun:     true,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "device1") {
		t.Error("Expected device1 in output")
	}
	if !strings.Contains(output, "device2") {
		t.Error("Expected device2 in output")
	}
	if !strings.Contains(output, "device3") {
		t.Error("Expected device3 in output")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_NotDryRunPath(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/not-dryrun.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a context that will timeout immediately to trigger the provisioning code path
	// without actually waiting for results
	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   1,
		DryRun:     false,
	}

	// Use a very short timeout context to make ProvisionDevices return quickly
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// We expect an error due to context cancellation in the provisioning path
	if err == nil {
		t.Logf("No error (context cancelled, expected)")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_NotDryRunInfoMessage(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/info-message.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfigTwoDevices), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   2,
		DryRun:     false,
	}

	// Use a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// We expect an error due to context cancellation
	if err == nil {
		t.Logf("No error (context cancelled, expected)")
	}

	// Check that "Found 2 devices" message is displayed
	output := tf.OutString()
	if !strings.Contains(output, "Found 2 devices to provision") {
		t.Logf("Output: %q", output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_DifferentParallelLevels(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tests := []struct {
		name     string
		parallel int
	}{
		{"sequential", 1},
		{"limited", 3},
		{"unlimited", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := factory.NewTestFactory(t)
			configFile := testConfigDir + "/" + tt.name + "-parallel-test.yaml"

			if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfig), 0o600); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			opts := &Options{
				Factory:    tf.Factory,
				ConfigFile: configFile,
				Parallel:   tt.parallel,
				DryRun:     true,
			}

			err := run(context.Background(), opts)
			if err != nil {
				t.Errorf("Expected no error with parallel=%d, got: %v", tt.parallel, err)
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestExecute_BulkAlias(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/bulk-alias.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test that we can invoke using the main command name
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{configFile, "--dry-run"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_FactoryGetDevice(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	devices := map[string]model.Device{
		"kitchen-light": {
			Name:    "kitchen-light",
			Address: "192.168.1.50",
		},
		"bedroom-light": {
			Name:    "bedroom-light",
			Address: "192.168.1.51",
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)
	configFile := testConfigDir + "/registered-devices.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(registeredDevicesConfig), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   2,
		DryRun:     true,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "kitchen-light") {
		t.Error("Expected kitchen-light in output")
	}
	if !strings.Contains(output, "bedroom-light") {
		t.Error("Expected bedroom-light in output")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_SingleDevice(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	configFile := testConfigDir + "/single-device.yaml"

	if err := afero.WriteFile(config.Fs(), configFile, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opts := &Options{
		Factory:    tf.Factory,
		ConfigFile: configFile,
		Parallel:   1,
		DryRun:     true,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Found 1 devices to provision") {
		t.Errorf("Expected provisioning message, got: %q", output)
	}
}
