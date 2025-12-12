// Package position provides the cover position subcommand.
package position

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

	if cmd.Use != "position <device> <percent>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "position <device> <percent>")
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
	cmd := NewCommand(cmdutil.NewFactory())

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
	cmd := NewCommand(cmdutil.NewFactory())

	// The command should require exactly 2 arguments
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}
