package remove

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "remove <group> <device>..." {
		t.Errorf("Use = %q, want \"remove <group> <device>...\"", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "rm" {
		t.Errorf("Aliases = %v, want [rm]", cmd.Aliases)
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

	cmd := NewCommand(cmdutil.NewFactory())

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
