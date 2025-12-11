package deletecmd

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	if cmd.Use != "delete <name>" {
		t.Errorf("Use = %q, want \"delete <name>\"", cmd.Use)
	}
	aliases := []string{"rm", "del"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, alias := range aliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
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

	// Test requires exactly 1 argument
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

	cmd := NewCommand()

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
