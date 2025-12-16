package deletecmd

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "delete <scene>" {
		t.Errorf("Use = %q, want \"delete <scene>\"", cmd.Use)
	}
	aliases := []string{"rm", "del", "remove"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
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

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "extra"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want false", yesFlag.DefValue)
	}
}
