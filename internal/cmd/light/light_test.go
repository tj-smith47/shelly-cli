// Package light provides the light command and its subcommands.
package light

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "light" {
		t.Errorf("Use = %q, want %q", cmd.Use, "light")
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
	cmd := NewCommand()

	if len(cmd.Aliases) == 0 {
		t.Error("No aliases defined")
	}

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "lt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'lt' not found")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

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
