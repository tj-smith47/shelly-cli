package list

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want list", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "ls" {
		t.Errorf("Aliases = %v, want [ls]", cmd.Aliases)
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	genFlag := cmd.Flags().Lookup("generation")
	if genFlag == nil {
		t.Fatal("generation flag not found")
	}
	if genFlag.Shorthand != "g" {
		t.Errorf("generation shorthand = %q, want g", genFlag.Shorthand)
	}

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("type flag not found")
	}
	if typeFlag.Shorthand != "t" {
		t.Errorf("type shorthand = %q, want t", typeFlag.Shorthand)
	}

	platformFlag := cmd.Flags().Lookup("platform")
	if platformFlag == nil {
		t.Fatal("platform flag not found")
	}
	if platformFlag.Shorthand != "p" {
		t.Errorf("platform shorthand = %q, want p", platformFlag.Shorthand)
	}
}
