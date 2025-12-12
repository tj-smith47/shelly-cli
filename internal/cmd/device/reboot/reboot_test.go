package reboot

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "reboot <device>" {
		t.Errorf("Use = %q, want 'reboot <device>'", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "restart" {
		t.Errorf("Aliases = %v, want [restart]", cmd.Aliases)
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	delayFlag := cmd.Flags().Lookup("delay")
	if delayFlag == nil {
		t.Fatal("delay flag not found")
	}
	if delayFlag.Shorthand != "d" {
		t.Errorf("delay shorthand = %q, want d", delayFlag.Shorthand)
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}
}
