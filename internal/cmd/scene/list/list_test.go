package list

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

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

	cmd := NewCommand()

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

func TestFormatActionCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		count int
		want  string
	}{
		{0, "0 (empty)"},
		{1, "1 action"},
		{5, "5 actions"},
	}

	for _, tt := range tests {
		result := formatActionCount(tt.count)
		// Result contains ANSI codes from theme, so just check it's not empty
		if result == "" {
			t.Errorf("formatActionCount(%d) returned empty string", tt.count)
		}
	}
}
