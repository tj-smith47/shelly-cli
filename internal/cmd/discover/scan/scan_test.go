// Package scan provides subnet scanning discovery command.
package scan

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "scan [subnet]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "scan [subnet]")
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

	// Test register flag exists
	register := cmd.Flags().Lookup("register")
	if register == nil {
		t.Error("register flag not found")
	} else if register.DefValue != "false" {
		t.Errorf("register default = %q, want %q", register.DefValue, "false")
	}

	// Test skip-existing flag exists
	skipExisting := cmd.Flags().Lookup("skip-existing")
	if skipExisting == nil {
		t.Error("skip-existing flag not found")
	} else if skipExisting.DefValue != "true" {
		t.Errorf("skip-existing default = %q, want %q", skipExisting.DefValue, "true")
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
