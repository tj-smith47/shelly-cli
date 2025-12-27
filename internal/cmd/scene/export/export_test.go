package export

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "export <name> [file]" {
		t.Errorf("Use = %q, want \"export <name> [file]\"", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires 1-2 arguments
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "file.yaml"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "file", "extra"}); err == nil {
		t.Error("expected error with 3 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("format shorthand = %q, want f", formatFlag.Shorthand)
	}
	if formatFlag.DefValue != "yaml" {
		t.Errorf("format default = %q, want yaml", formatFlag.DefValue)
	}
}
