// Package open provides the cover open subcommand.
package open

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "open <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "open <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
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

	// Test duration flag exists
	durationFlag := cmd.Flags().Lookup("duration")
	switch {
	case durationFlag == nil:
		t.Error("duration flag not found")
	case durationFlag.Shorthand != "d":
		t.Errorf("duration shorthand = %q, want %q", durationFlag.Shorthand, "d")
	case durationFlag.DefValue != "0":
		t.Errorf("duration default = %q, want %q", durationFlag.DefValue, "0")
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
