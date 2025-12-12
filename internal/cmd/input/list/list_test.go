// Package list provides the input list subcommand.
package list

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

	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// The command should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}
