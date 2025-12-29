package term

import (
	"context"
	"strings"
	"testing"
)

func TestDisplayShellHelp(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayShellHelp(ios)

	output := out.String()
	if !strings.Contains(output, "Shell Commands") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Built-in") {
		t.Error("expected built-in section")
	}
	if !strings.Contains(output, "help") {
		t.Error("expected help command")
	}
	if !strings.Contains(output, "info") {
		t.Error("expected info command")
	}
	if !strings.Contains(output, "status") {
		t.Error("expected status command")
	}
	if !strings.Contains(output, "config") {
		t.Error("expected config command")
	}
	if !strings.Contains(output, "methods") {
		t.Error("expected methods command")
	}
	if !strings.Contains(output, "components") {
		t.Error("expected components command")
	}
	if !strings.Contains(output, "exit") {
		t.Error("expected exit command")
	}
	if !strings.Contains(output, "RPC Methods") {
		t.Error("expected RPC methods section")
	}
	if !strings.Contains(output, "Examples") {
		t.Error("expected examples section")
	}
	if !strings.Contains(output, "Switch.GetStatus") {
		t.Error("expected example command")
	}
}

func TestNewShellSession(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewShellSession(ios, nil, "test-device")

	if session == nil {
		t.Fatal("expected session to be created")
	}
	if session.Device != "test-device" {
		t.Errorf("device = %q, want %q", session.Device, "test-device")
	}
	if session.IOS != ios {
		t.Error("expected ios to be set")
	}
	if session.Conn != nil {
		t.Error("expected conn to be nil")
	}
}

func TestShellSession_ExecuteCommand_Exit(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	session := NewShellSession(ios, nil, "test-device")

	exitCommands := []string{"exit", "quit", "q", "EXIT", "QUIT", "Q"}
	for _, cmd := range exitCommands {
		shouldExit := session.ExecuteCommand(context.Background(), cmd)
		if !shouldExit {
			t.Errorf("ExecuteCommand(%q) should return true for exit", cmd)
		}
	}
}

func TestShellSession_ExecuteCommand_Help(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	session := NewShellSession(ios, nil, "test-device")

	helpCommands := []string{"help", "h", "?", "HELP", "H"}
	for _, cmd := range helpCommands {
		out.Reset()
		shouldExit := session.ExecuteCommand(context.Background(), cmd)
		if shouldExit {
			t.Errorf("ExecuteCommand(%q) should not exit", cmd)
		}
		if !strings.Contains(out.String(), "Shell Commands") {
			t.Errorf("ExecuteCommand(%q) should display help", cmd)
		}
	}
}
