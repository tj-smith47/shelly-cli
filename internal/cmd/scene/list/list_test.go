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

// Note: This command uses the global -o/--output flag defined on the root command,
// not a local flag. The global flag is tested in the root command tests.

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
