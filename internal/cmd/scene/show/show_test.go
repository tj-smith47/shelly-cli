package show

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "show <name>" {
		t.Errorf("Use = %q, want \"show <name>\"", cmd.Use)
	}
	aliases := []string{"info", "get"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
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

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "extra"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want o", outputFlag.Shorthand)
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("output default = %q, want table", outputFlag.DefValue)
	}
}

func TestFormatParamsInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params map[string]any
		empty  bool
	}{
		{"nil params", nil, true},
		{"empty params", map[string]any{}, true},
		{"with params", map[string]any{"id": 0}, false},
	}

	for _, tt := range tests {
		result := output.FormatParamsInline(tt.params)
		if tt.empty && result != "" {
			t.Errorf("%s: FormatParamsInline() = %q, want empty", tt.name, result)
		}
		if !tt.empty && result == "" {
			t.Errorf("%s: FormatParamsInline() = empty, want non-empty", tt.name)
		}
	}
}
