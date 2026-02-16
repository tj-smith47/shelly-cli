package factories

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewInputToggleCommand_Enable(t *testing.T) {
	t.Parallel()

	cmd := NewInputToggleCommand(cmdutil.NewFactory(), InputToggleOpts{
		Enable:  true,
		Long:    "Enable an input component.",
		Example: "  shelly input enable kitchen",
	})

	if cmd.Use != "enable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enable <device>")
	}

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "on" {
		t.Errorf("Aliases = %v, want [on]", cmd.Aliases)
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("--id flag not found")
	}
}

func TestNewInputToggleCommand_Disable(t *testing.T) {
	t.Parallel()

	cmd := NewInputToggleCommand(cmdutil.NewFactory(), InputToggleOpts{
		Enable:  false,
		Long:    "Disable an input component.",
		Example: "  shelly input disable kitchen",
	})

	if cmd.Use != "disable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "disable <device>")
	}

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "off" {
		t.Errorf("Aliases = %v, want [off]", cmd.Aliases)
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewInputToggleCommand_MissingArg(t *testing.T) {
	t.Parallel()

	cmd := NewInputToggleCommand(cmdutil.NewFactory(), InputToggleOpts{
		Enable:  true,
		Long:    "Enable.",
		Example: "  shelly input enable dev",
	})

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
}
