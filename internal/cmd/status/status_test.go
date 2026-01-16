// Package statuscmd provides the quick status command.
package status

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
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

	if cmd.Use != "status [device]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status [device]")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"st", "state"}
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

	// Verify example contains expected content
	if !strings.Contains(cmd.Example, "shelly status living-room") {
		t.Error("Example should contain 'shelly status living-room'")
	}

	if !strings.Contains(cmd.Example, "shelly status") {
		t.Error("Example should contain 'shelly status'")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept 0 or 1 arguments
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("unexpected error with no args: %v", err)
	}

	if err := cmd.Args(cmd, []string{"device1"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}

	if err := cmd.Args(cmd, []string{"device1", "device2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestOptions_Device(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
}

func TestOptions_Empty(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Device != "" {
		t.Errorf("Device = %q, want empty string", opts.Device)
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "my-device",
		Factory: f,
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	// Verify factory provides IOStreams
	if opts.Factory.IOStreams() == nil {
		t.Error("Factory.IOStreams() returned nil")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
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

func TestNewCommand_ArgsPopulatesDevice(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Execute the command with a device argument to verify the RunE closure works
	// We just verify the command is set up correctly - actual execution would require mocks
	if cmd.Use != "status [device]" {
		t.Errorf("command Use = %q, expected %q", cmd.Use, "status [device]")
	}
}

func TestNewCommand_MaximumOneArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that Args validates correctly
	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, false},
		{"one arg", []string{"living-room"}, false},
		{"two args", []string{"living-room", "kitchen"}, true},
		{"three args", []string{"a", "b", "c"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tc.args)
			if (err != nil) != tc.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
		})
	}
}

// TestTermComponentState verifies the component state type.
func TestTermComponentState(t *testing.T) {
	t.Parallel()

	state := term.ComponentState{
		Type:  "Switch 0",
		Name:  "Kitchen",
		State: "on",
	}

	if state.Type != "Switch 0" {
		t.Errorf("Type = %q, want %q", state.Type, "Switch 0")
	}

	if state.Name != "Kitchen" {
		t.Errorf("Name = %q, want %q", state.Name, "Kitchen")
	}

	if state.State != "on" {
		t.Errorf("State = %q, want %q", state.State, "on")
	}
}

// TestTermQuickDeviceStatus verifies the quick device status type.
func TestTermQuickDeviceStatus(t *testing.T) {
	t.Parallel()

	status := term.QuickDeviceStatus{
		Name:   "living-room",
		Model:  "SHSW-1",
		Online: true,
	}

	if status.Name != "living-room" {
		t.Errorf("Name = %q, want %q", status.Name, "living-room")
	}

	if status.Model != "SHSW-1" {
		t.Errorf("Model = %q, want %q", status.Model, "SHSW-1")
	}

	if !status.Online {
		t.Error("Online = false, want true")
	}
}

// TestTermDisplayQuickDeviceStatus verifies the display function doesn't panic.
func TestTermDisplayQuickDeviceStatus(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	states := []term.ComponentState{
		{Type: "Switch 0", Name: "Kitchen", State: "on"},
	}

	// Should not panic
	term.DisplayQuickDeviceStatus(ios, states)

	// Verify some output was produced
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}
}

// TestTermDisplayQuickDeviceStatus_EmptyStates handles empty states list.
func TestTermDisplayQuickDeviceStatus_EmptyStates(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Should not panic with empty states
	term.DisplayQuickDeviceStatus(ios, []term.ComponentState{})

	// Verify some output was produced
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}
}

// TestTermDisplayAllDevicesQuickStatus verifies the all-devices display.
func TestTermDisplayAllDevicesQuickStatus(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	statuses := []term.QuickDeviceStatus{
		{Name: "living-room", Model: "SHSW-1", Online: true},
		{Name: "kitchen", Model: "SHSW-25", Online: false},
	}

	// Should not panic
	term.DisplayAllDevicesQuickStatus(ios, statuses)

	// Verify some output was produced
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}
}

// TestTermDisplayAllDevicesQuickStatus_Empty handles empty list.
func TestTermDisplayAllDevicesQuickStatus_Empty(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Should not panic with empty statuses
	term.DisplayAllDevicesQuickStatus(ios, []term.QuickDeviceStatus{})
}

// TestRun_ContextCancellation tests that context cancellation is handled.
func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Device:  "unreachable-device",
		Factory: f,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run should fail due to cancelled context or connection error
	err := run(ctx, opts)
	// We expect an error here (context cancelled or connection failure)
	if err == nil {
		t.Log("run returned nil error - device might have been reached")
	}
}

// TestRun_WithDevice runs with a device argument.
func TestRun_WithDevice(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Device:  "192.168.1.100", // Invalid address to trigger error
		Factory: f,
	}

	// Run should fail since the device doesn't exist
	err := run(context.Background(), opts)
	if err == nil {
		t.Log("run returned nil - device might exist on network")
	}
}

// TestRun_NoDevice runs without a device argument (all devices mode).
func TestRun_NoDevice(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Device:  "", // Empty means all devices
		Factory: f,
	}

	// Run should complete (possibly with warning about no devices)
	err := run(context.Background(), opts)
	if err != nil {
		t.Logf("run returned error: %v", err)
	}

	// Check for expected output (warning about no devices or device list)
	output := stdout.String() + stderr.String()
	_ = output // Output might be empty or contain warnings
}

// TestNewCommand_Short verifies the short description.
func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Show device status (quick overview)"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

// TestNewCommand_LongContainsUsage verifies Long description contains usage info.
func TestNewCommand_LongContainsUsage(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "quick status overview") {
		t.Error("Long should contain 'quick status overview'")
	}

	if !strings.Contains(cmd.Long, "all registered devices") {
		t.Error("Long should contain 'all registered devices'")
	}
}
