package activate

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	if cmd.Use != "activate <name>" {
		t.Errorf("Use = %q, want \"activate <name>\"", cmd.Use)
	}
	aliases := []string{"run", "exec", "play"}
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

	cmd := NewCommand()

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

	cmd := NewCommand()

	flags := []struct {
		name      string
		shorthand string
	}{
		{"timeout", "t"},
		{"concurrent", "c"},
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

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("dry-run flag not found")
	}
}

func TestFormatParams(t *testing.T) {
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
		result := formatParams(tt.params)
		if tt.empty && result != "" {
			t.Errorf("%s: formatParams() = %q, want empty", tt.name, result)
		}
		if !tt.empty && result == "" {
			t.Errorf("%s: formatParams() = empty, want non-empty", tt.name)
		}
	}
}
