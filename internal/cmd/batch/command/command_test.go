package command

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "command <method> [params-json] [device...]" {
		t.Errorf("Use = %q, want \"command <method> [params-json] [device...]\"", cmd.Use)
	}
	aliases := []string{"cmd", "rpc"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, alias := range aliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
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

	// Test requires minimum 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"Shelly.GetStatus"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"Switch.Set", "{\"id\":0}"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flags := []struct {
		name      string
		shorthand string
	}{
		{"group", "g"},
		{"all", "a"},
		{"timeout", "t"},
		{"concurrent", "c"},
		{"output", "o"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not found", f.name)
			continue
		}
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag %q shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestIsJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{"{}", true},
		{`{"id":0}`, true},
		{`{"on":true,"id":0}`, true},
		{"", false},
		{"hello", false},
		{"[1,2,3]", false}, // Arrays don't count
		{"123", false},
	}

	for _, tt := range tests {
		got := isJSON(tt.input)
		if got != tt.want {
			t.Errorf("isJSON(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
