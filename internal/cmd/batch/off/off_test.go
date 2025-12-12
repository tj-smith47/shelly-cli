package off

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "off [device...]" {
		t.Errorf("Use = %q, want \"off [device...]\"", cmd.Use)
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

	flags := []struct {
		name      string
		shorthand string
	}{
		{"group", "g"},
		{"all", "a"},
		{"timeout", "t"},
		{"switch", "s"},
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
}
