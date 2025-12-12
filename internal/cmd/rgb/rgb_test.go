// Package rgb provides the rgb command and its subcommands.
package rgb

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "rgb" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rgb")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) != 6 {
		t.Errorf("Subcommand count = %d, want 6", len(subcommands))
	}

	expectedSubcommands := map[string]bool{
		"on":     false,
		"off":    false,
		"toggle": false,
		"status": false,
		"set":    false,
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
