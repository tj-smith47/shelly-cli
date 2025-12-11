// Package status provides the input status subcommand.
package status

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "status <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device>")
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
		if alias == "st" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected alias 'st' not found")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	// Test id flag exists
	idFlag := cmd.Flags().Lookup("id")
	switch {
	case idFlag == nil:
		t.Error("id flag not found")
	case idFlag.Shorthand != "i":
		t.Errorf("id shorthand = %q, want %q", idFlag.Shorthand, "i")
	case idFlag.DefValue != "0":
		t.Errorf("id default = %q, want %q", idFlag.DefValue, "0")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	// The command should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}
