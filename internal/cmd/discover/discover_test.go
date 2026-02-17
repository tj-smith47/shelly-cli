// Package discover provides device discovery commands.
package discover

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

const (
	testMethodHTTP      = "http"
	testPlatformTasmota = "tasmota"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "discover" {
		t.Errorf("Use = %q, want %q", cmd.Use, "discover")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify subcommands are registered
	subcommands := cmd.Commands()
	expectedSubcmds := map[string]bool{
		"mdns":             false,
		"ble":              false,
		"coiot":            false,
		"http [subnet...]": false,
	}

	for _, sub := range subcommands {
		if _, ok := expectedSubcmds[sub.Use]; ok {
			expectedSubcmds[sub.Use] = true
		}
	}

	for name, found := range expectedSubcmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test timeout flag exists
	timeout := cmd.Flags().Lookup("timeout")
	switch {
	case timeout == nil:
		t.Error("timeout flag not found")
	case timeout.Shorthand != "t":
		t.Errorf("timeout shorthand = %q, want %q", timeout.Shorthand, "t")
	case timeout.DefValue != "2m0s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "2m0s")
	}

	// Test register flag exists
	register := cmd.Flags().Lookup("register")
	if register == nil {
		t.Error("register flag not found")
	} else if register.DefValue != "false" {
		t.Errorf("register default = %q, want %q", register.DefValue, "false")
	}

	// Test skip-existing flag exists
	skipExisting := cmd.Flags().Lookup("skip-existing")
	if skipExisting == nil {
		t.Error("skip-existing flag not found")
	} else if skipExisting.DefValue != "true" {
		t.Errorf("skip-existing default = %q, want %q", skipExisting.DefValue, "true")
	}
}

func TestNewCommand_SubcommandCount(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should have exactly 4 subcommands
	if len(cmd.Commands()) != 4 {
		t.Errorf("subcommand count = %d, want 4", len(cmd.Commands()))
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"disc", "find"}
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

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Check for expected examples
	if !strings.Contains(cmd.Example, "shelly discover") {
		t.Error("Example should contain 'shelly discover'")
	}

	if !strings.Contains(cmd.Example, "--register") {
		t.Error("Example should contain '--register'")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_MethodFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	methodFlag := cmd.Flags().Lookup("method")
	if methodFlag == nil {
		t.Error("method flag not found")
		return
	}

	if methodFlag.Shorthand != "m" {
		t.Errorf("method shorthand = %q, want %q", methodFlag.Shorthand, "m")
	}

	if methodFlag.DefValue != testMethodHTTP {
		t.Errorf("method default = %q, want %q", methodFlag.DefValue, testMethodHTTP)
	}
}

func TestNewCommand_SubnetFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	subnetFlag := cmd.Flags().Lookup("subnet")
	if subnetFlag == nil {
		t.Error("subnet flag not found")
		return
	}

	// Default should be empty array (auto-detected)
	if subnetFlag.DefValue != "[]" {
		t.Errorf("subnet default = %q, want %q", subnetFlag.DefValue, "[]")
	}
}

func TestNewCommand_SkipPluginsFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	skipPluginsFlag := cmd.Flags().Lookup("skip-plugins")
	if skipPluginsFlag == nil {
		t.Error("skip-plugins flag not found")
		return
	}

	if skipPluginsFlag.DefValue != "false" {
		t.Errorf("skip-plugins default = %q, want %q", skipPluginsFlag.DefValue, "false")
	}
}

func TestNewCommand_PlatformFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	platformFlag := cmd.Flags().Lookup("platform")
	if platformFlag == nil {
		t.Error("platform flag not found")
		return
	}

	if platformFlag.Shorthand != "p" {
		t.Errorf("platform shorthand = %q, want %q", platformFlag.Shorthand, "p")
	}

	// Default should be empty
	if platformFlag.DefValue != "" {
		t.Errorf("platform default = %q, want empty string", platformFlag.DefValue)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Timeout != 0 {
		t.Errorf("default timeout = %v, want 0", opts.Timeout)
	}

	if opts.Register {
		t.Error("default register = true, want false")
	}

	if opts.SkipExisting {
		t.Error("default skipExisting = true, want false (from zero value)")
	}

	if len(opts.Subnets) != 0 {
		t.Errorf("default subnets = %v, want empty", opts.Subnets)
	}

	if opts.Method != "" {
		t.Errorf("default method = %q, want empty string", opts.Method)
	}
}

func TestOptions_AllFields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Timeout:      30 * time.Second,
		Register:     true,
		SkipExisting: true,
		Subnets:      []string{"192.168.1.0/24"},
		Method:       "http",
		SkipPlugins:  true,
		Platform:     "tasmota",
	}

	if opts.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}

	if !opts.Register {
		t.Error("register = false, want true")
	}

	if !opts.SkipExisting {
		t.Error("skipExisting = false, want true")
	}

	if len(opts.Subnets) != 1 || opts.Subnets[0] != "192.168.1.0/24" {
		t.Errorf("subnets = %v, want [192.168.1.0/24]", opts.Subnets)
	}

	if opts.Method != testMethodHTTP {
		t.Errorf("method = %q, want %q", opts.Method, testMethodHTTP)
	}

	if !opts.SkipPlugins {
		t.Error("skipPlugins = false, want true")
	}

	if opts.Platform != testPlatformTasmota {
		t.Errorf("platform = %q, want %q", opts.Platform, testPlatformTasmota)
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

	// Verify the factory's IOStreams are used
	if f.IOStreams() == nil {
		t.Error("Factory IOStreams is nil")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check for key information in long description
	if !strings.Contains(cmd.Long, "HTTP subnet scanning") {
		t.Error("Long should mention HTTP subnet scanning")
	}

	if !strings.Contains(cmd.Long, "mDNS") {
		t.Error("Long should mention mDNS")
	}

	if !strings.Contains(cmd.Long, "ble") && !strings.Contains(cmd.Long, "Bluetooth") {
		t.Error("Long should mention ble or Bluetooth")
	}

	if !strings.Contains(cmd.Long, "CoIoT") {
		t.Error("Long should mention CoIoT")
	}
}

func TestNewCommand_AllFlagsExist(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	requiredFlags := []string{
		"timeout",
		"register",
		"skip-existing",
		"subnet",
		"method",
		"skip-plugins",
		"platform",
	}

	for _, flag := range requiredFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag %q not found", flag)
		}
	}
}

// TestTermPluginDiscoveredDevice tests the plugin discovered device type.
func TestTermPluginDiscoveredDevice(t *testing.T) {
	t.Parallel()

	device := term.PluginDiscoveredDevice{
		ID:       "tasmota-123",
		Name:     "Kitchen Light",
		Model:    "Sonoff Basic",
		Address:  "192.168.1.100",
		Platform: "tasmota",
		Firmware: "12.0.0",
		Components: []term.PluginComponentInfo{
			{Type: "switch", ID: 0, Name: "Relay"},
		},
	}

	if device.ID != "tasmota-123" {
		t.Errorf("ID = %q, want %q", device.ID, "tasmota-123")
	}

	if device.Name != "Kitchen Light" {
		t.Errorf("Name = %q, want %q", device.Name, "Kitchen Light")
	}

	if device.Model != "Sonoff Basic" {
		t.Errorf("Model = %q, want %q", device.Model, "Sonoff Basic")
	}

	if device.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want %q", device.Address, "192.168.1.100")
	}

	if device.Platform != "tasmota" {
		t.Errorf("Platform = %q, want %q", device.Platform, "tasmota")
	}

	if len(device.Components) != 1 {
		t.Errorf("Components count = %d, want 1", len(device.Components))
	}
}

// TestTermPluginComponentInfo tests the plugin component info type.
func TestTermPluginComponentInfo(t *testing.T) {
	t.Parallel()

	component := term.PluginComponentInfo{
		Type: "switch",
		ID:   0,
		Name: "Relay 1",
	}

	if component.Type != "switch" {
		t.Errorf("Type = %q, want %q", component.Type, "switch")
	}

	if component.ID != 0 {
		t.Errorf("ID = %d, want %d", component.ID, 0)
	}

	if component.Name != "Relay 1" {
		t.Errorf("Name = %q, want %q", component.Name, "Relay 1")
	}
}

// TestTermDisplayPluginDiscoveredDevices verifies the display function doesn't panic.
func TestTermDisplayPluginDiscoveredDevices(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	devices := []term.PluginDiscoveredDevice{
		{
			ID:       "tasmota-123",
			Name:     "Kitchen Light",
			Model:    "Sonoff Basic",
			Address:  "192.168.1.100",
			Platform: "tasmota",
		},
		{
			ID:       "esphome-456",
			Name:     "Garage Door",
			Model:    "ESP32",
			Address:  "192.168.1.101",
			Platform: "esphome",
		},
	}

	// Should not panic
	term.DisplayPluginDiscoveredDevices(ios, devices)

	// Verify some output was produced
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}
}

// TestTermDisplayPluginDiscoveredDevices_Empty handles empty list.
func TestTermDisplayPluginDiscoveredDevices_Empty(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Should not panic with empty devices
	term.DisplayPluginDiscoveredDevices(ios, []term.PluginDiscoveredDevice{})
}

// TestTermDisplayDiscoveredDevices_Empty handles empty list.
func TestTermDisplayDiscoveredDevices_Empty(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Import discovery package to access types
	// Should not panic with empty devices
	term.DisplayDiscoveredDevices(ios, nil)
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Discover Shelly devices on the network"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_SubcommandAliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	for _, subcmd := range cmd.Commands() {
		// Each subcommand should have at least one alias
		if subcmd.Use == "mdns" && len(subcmd.Aliases) == 0 {
			t.Error("mdns subcommand should have aliases")
		}
	}
}

// TestRun_UnknownMethod tests the run function with an unknown method.
func TestRun_UnknownMethod(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Method:  "invalid-method",
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for unknown method")
	}

	if !strings.Contains(err.Error(), "unknown discovery method") {
		t.Errorf("Error = %q, should contain 'unknown discovery method'", err.Error())
	}
}

// TestRun_HTTPMethod tests the run function with HTTP method.
func TestRun_HTTPMethod(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use cancelled context to prevent actual scan
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:     f,
		Method:      "http",
		Timeout:     1 * time.Millisecond,
		SkipPlugins: true,
	}

	// Should fail due to cancelled context or subnet detection
	err := run(ctx, opts)
	if err != nil {
		t.Logf("HTTP discovery error (expected): %v", err)
	}
}

// TestRun_MDNSMethod tests the run function with mDNS method.
func TestRun_MDNSMethod(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: f,
		Method:  "mdns",
		Timeout: 1 * time.Millisecond,
	}

	err := run(ctx, opts)
	if err != nil {
		t.Logf("mDNS discovery error (expected): %v", err)
	}
}

// TestRun_CoIoTMethod tests the run function with CoIoT method.
func TestRun_CoIoTMethod(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: f,
		Method:  "coiot",
		Timeout: 1 * time.Millisecond,
	}

	err := run(ctx, opts)
	if err != nil {
		t.Logf("CoIoT discovery error (expected): %v", err)
	}
}

// TestRun_BLEMethod tests the run function with BLE method.
// NOTE: No t.Parallel() - BLE library has race when multiple tests init concurrently.
func TestRun_BLEMethod(t *testing.T) { //nolint:paralleltest // Intentional - BLE library race
	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: f,
		Method:  "ble",
		Timeout: 1 * time.Millisecond,
	}

	err := run(ctx, opts)
	if err != nil {
		t.Logf("BLE discovery error (expected - BLE might not be available): %v", err)
	}
}

// TestRun_PlatformFilter tests the run function with platform filter.
func TestRun_PlatformFilter(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  f,
		Platform: "tasmota",
		Timeout:  1 * time.Millisecond,
	}

	err := run(ctx, opts)
	if err != nil {
		t.Logf("platform filter error (expected - no plugin for tasmota likely): %v", err)
	}
}
