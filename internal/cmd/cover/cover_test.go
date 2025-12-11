// Package cover provides the cover command and its subcommands.
package cover

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "cover" {
		t.Errorf("Use = %q, want %q", cmd.Use, "cover")
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
		if alias == "cv" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'cv' not found")
	}
}

func TestNewCommand_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	subcommands := cmd.Commands()
	if len(subcommands) != 7 {
		t.Errorf("Subcommand count = %d, want 7", len(subcommands))
	}

	expectedSubcommands := map[string]bool{
		"open":      false,
		"close":     false,
		"stop":      false,
		"status":    false,
		"position":  false,
		"calibrate": false,
		"list":      false,
	}

	for _, sub := range subcommands {
		for name := range expectedSubcommands {
			if sub.Use == name || sub.Use == name+" <device>" || sub.Use == name+" <device> <percent>" {
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
