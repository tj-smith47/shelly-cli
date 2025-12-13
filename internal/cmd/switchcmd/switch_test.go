// Package switchcmd provides the switch command group for controlling relay switches.
package switchcmd

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "switch" {
		t.Errorf("Use = %q, want %q", cmd.Use, "switch")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("No aliases defined")
	}

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "sw" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'sw' not found")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) != 5 {
		t.Errorf("Subcommand count = %d, want 5", len(subcommands))
	}

	expectedSubcommands := map[string]bool{
		"on":     false,
		"off":    false,
		"toggle": false,
		"status": false,
		"list":   false,
	}

	for _, sub := range subcommands {
		for name := range expectedSubcommands {
			if sub.Use == name || sub.Use == name+" <device>" {
				expectedSubcommands[name] = true
			}
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}
