// Package input provides the input command and its subcommands.
package input

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

	if cmd.Use != "input" {
		t.Errorf("Use = %q, want %q", cmd.Use, "input")
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
		if alias == "in" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'in' not found")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) != 3 {
		t.Errorf("Subcommand count = %d, want 3", len(subcommands))
	}

	expectedSubcommands := map[string]bool{
		"list":    false,
		"status":  false,
		"trigger": false,
	}

	for _, sub := range subcommands {
		switch sub.Use {
		case "list", "list <device>":
			expectedSubcommands["list"] = true
		case "status", "status <device>":
			expectedSubcommands["status"] = true
		case "trigger", "trigger <device>":
			expectedSubcommands["trigger"] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}
