package batch

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "batch" {
		t.Errorf("Use = %q, want batch", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "b" {
		t.Errorf("Aliases = %v, want [b]", cmd.Aliases)
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

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expected := []string{"on", "off", "toggle", "command"}
	subCmds := cmd.Commands()

	if len(subCmds) != len(expected) {
		t.Errorf("got %d subcommands, want %d", len(subCmds), len(expected))
	}

	// Check each expected subcommand exists
	for _, name := range expected {
		found := false
		for _, sub := range subCmds {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}
