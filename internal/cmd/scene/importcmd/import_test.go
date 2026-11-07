package importcmd

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want \"import <file>\"", cmd.Use)
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

	// Test requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"file.yaml"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"file1", "file2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("name flag not found")
	}
	if nameFlag.Shorthand != "n" {
		t.Errorf("name shorthand = %q, want n", nameFlag.Shorthand)
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("force flag not found")
	}
	if forceFlag.Shorthand != "f" {
		t.Errorf("force shorthand = %q, want f", forceFlag.Shorthand)
	}
	if forceFlag.DefValue != "false" {
		t.Errorf("force default = %q, want false", forceFlag.DefValue)
	}

	// --overwrite stays as a hidden alias so existing scripts keep working.
	overwriteFlag := cmd.Flags().Lookup("overwrite")
	if overwriteFlag == nil {
		t.Fatal("overwrite alias not found")
	}
	if !overwriteFlag.Hidden {
		t.Error("overwrite alias should be hidden")
	}
	if err := cmd.Flags().Set("overwrite", "true"); err != nil {
		t.Fatalf("set overwrite: %v", err)
	}
	if forceFlag.Value.String() != "true" {
		t.Error("--overwrite should set the same value as --force")
	}
}
