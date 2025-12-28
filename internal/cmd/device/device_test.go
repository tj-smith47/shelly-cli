package device

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "device" {
		t.Errorf("Use = %q, want device", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "dev" {
		t.Errorf("Aliases = %v, want [dev]", cmd.Aliases)
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	expected := []string{"alias", "config", "factory-reset", "info", "list", "ping", "reboot", "status", "ui"}
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
