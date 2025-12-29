package term

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestDisplayREPLHelp(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayREPLHelp(ios)

	output := out.String()
	if !strings.Contains(output, "Available Commands") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Navigation") {
		t.Error("expected navigation section")
	}
	if !strings.Contains(output, "Device Control") {
		t.Error("expected device control section")
	}
	if !strings.Contains(output, "Advanced") {
		t.Error("expected advanced section")
	}
	if !strings.Contains(output, "help") {
		t.Error("expected help command")
	}
	if !strings.Contains(output, "devices") {
		t.Error("expected devices command")
	}
	if !strings.Contains(output, "connect") {
		t.Error("expected connect command")
	}
	if !strings.Contains(output, "disconnect") {
		t.Error("expected disconnect command")
	}
	if !strings.Contains(output, "status") {
		t.Error("expected status command")
	}
	if !strings.Contains(output, "rpc") {
		t.Error("expected rpc command")
	}
}

func TestDisplayRPCMethods(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	methods := []string{
		"Shelly.GetDeviceInfo",
		"Switch.GetStatus",
		"Switch.Set",
		"Script.List",
	}
	DisplayRPCMethods(ios, methods)

	output := out.String()
	if !strings.Contains(output, "Available RPC Methods") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Shelly.GetDeviceInfo") {
		t.Error("expected first method")
	}
	if !strings.Contains(output, "Switch.GetStatus") {
		t.Error("expected second method")
	}
}

func TestDisplayControlResults_AllSuccess(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []shelly.ComponentControlResult{
		{Type: "switch", ID: 0, Success: true},
		{Type: "switch", ID: 1, Success: true},
	}
	DisplayControlResults(ios, results, "turned on")

	output := out.String()
	if !strings.Contains(output, "switch:0 turned on") {
		t.Error("expected success message for switch:0")
	}
	if !strings.Contains(output, "switch:1 turned on") {
		t.Error("expected success message for switch:1")
	}
}

func TestDisplayControlResults_WithFailures(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	results := []shelly.ComponentControlResult{
		{Type: "switch", ID: 0, Success: true},
		{Type: "switch", ID: 1, Success: false},
	}
	DisplayControlResults(ios, results, "toggled")

	// Success goes to stdout
	if !strings.Contains(out.String(), "switch:0 toggled") {
		t.Error("expected success message")
	}
	// Warning goes to stderr
	if !strings.Contains(errOut.String(), "Failed on") {
		t.Error("expected failure message")
	}
}

func TestNewREPLSession(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "kitchen")

	if session == nil {
		t.Fatal("expected session to be created")
	}
	if session.ActiveDevice != "kitchen" {
		t.Errorf("active device = %q, want %q", session.ActiveDevice, "kitchen")
	}
	if session.IOS != ios {
		t.Error("expected ios to be set")
	}
}

func TestREPLSession_ExecuteCommand_Exit(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "")

	ctx := context.Background()
	exitCommands := []string{"exit", "quit", "q"}
	for _, cmd := range exitCommands {
		shouldExit := session.ExecuteCommand(ctx, cmd, nil)
		if !shouldExit {
			t.Errorf("ExecuteCommand(%q) should return true for exit", cmd)
		}
	}
}

func TestREPLSession_ExecuteCommand_Help(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "")

	helpCommands := []string{"help", "h", "?"}
	for _, cmd := range helpCommands {
		out.Reset()
		shouldExit := session.ExecuteCommand(context.Background(), cmd, nil)
		if shouldExit {
			t.Errorf("ExecuteCommand(%q) should not exit", cmd)
		}
		if !strings.Contains(out.String(), "Available Commands") {
			t.Errorf("ExecuteCommand(%q) should display help", cmd)
		}
	}
}

func TestREPLSession_ExecuteCommand_Devices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "")

	shouldExit := session.ExecuteCommand(context.Background(), "devices", nil)
	if shouldExit {
		t.Error("devices command should not exit")
	}
	// Output depends on config state
	_ = out.String()
}

func TestREPLSession_ExecuteCommand_Connect(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "")

	// Without device name
	out.Reset()
	session.ExecuteCommand(context.Background(), "connect", nil)
	// Should show error

	// With device name
	out.Reset()
	session.ExecuteCommand(context.Background(), "connect", []string{"my-device"})
	if session.ActiveDevice != "my-device" {
		t.Errorf("active device = %q, want %q", session.ActiveDevice, "my-device")
	}
	if !strings.Contains(out.String(), "Connected") {
		t.Error("expected connected message")
	}
}

func TestREPLSession_ExecuteCommand_Disconnect(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "current-device")

	session.ExecuteCommand(context.Background(), "disconnect", nil)
	if session.ActiveDevice != "" {
		t.Errorf("active device should be empty after disconnect")
	}
	// Info message "Disconnected from device" goes to stdout
	if !strings.Contains(out.String(), "Disconnected from device") {
		t.Error("expected disconnected message")
	}
}

func TestREPLSession_ExecuteCommand_Unknown(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	session := NewREPLSession(ios, nil, "")

	session.ExecuteCommand(context.Background(), "unknowncommand", nil)
	// Warning goes to stderr
	if !strings.Contains(errOut.String(), "Unknown command") {
		t.Error("expected unknown command message")
	}
}

func TestREPLSession_resolveDevice_WithArgs(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "active-device")

	// With args, should use arg
	device := session.resolveDevice([]string{"arg-device"})
	if device != "arg-device" {
		t.Errorf("device = %q, want %q", device, "arg-device")
	}
}

func TestREPLSession_resolveDevice_NoArgs(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewREPLSession(ios, nil, "active-device")

	// Without args, should use active device
	device := session.resolveDevice(nil)
	if device != "active-device" {
		t.Errorf("device = %q, want %q", device, "active-device")
	}
}
