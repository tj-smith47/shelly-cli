package add

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	if cmd.Use != "add <group> <device>..." {
		t.Errorf("Use = %q, want \"add <group> <device>...\"", cmd.Use)
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
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	// Test requires minimum 2 arguments
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"group"}); err == nil {
		t.Error("expected error with only 1 arg")
	}
	if err := cmd.Args(cmd, []string{"group", "device"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
	if err := cmd.Args(cmd, []string{"group", "device1", "device2"}); err != nil {
		t.Errorf("unexpected error with 3 args: %v", err)
	}
}
