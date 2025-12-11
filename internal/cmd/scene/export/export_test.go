package export

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

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

	cmd := NewCommand()

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

	cmd := NewCommand()

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want o", outputFlag.Shorthand)
	}
	if outputFlag.DefValue != "yaml" {
		t.Errorf("output default = %q, want yaml", outputFlag.DefValue)
	}
}
