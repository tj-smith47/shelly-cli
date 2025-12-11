// Package set provides the light set subcommand.
package set

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "set <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device>")
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

	// Test brightness flag exists
	brightnessFlag := cmd.Flags().Lookup("brightness")
	switch {
	case brightnessFlag == nil:
		t.Error("brightness flag not found")
	case brightnessFlag.Shorthand != "b":
		t.Errorf("brightness shorthand = %q, want %q", brightnessFlag.Shorthand, "b")
	case brightnessFlag.DefValue != "-1":
		t.Errorf("brightness default = %q, want %q", brightnessFlag.DefValue, "-1")
	}

	// Test on flag exists
	onFlag := cmd.Flags().Lookup("on")
	switch {
	case onFlag == nil:
		t.Error("on flag not found")
	case onFlag.DefValue != "false":
		t.Errorf("on default = %q, want %q", onFlag.DefValue, "false")
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
