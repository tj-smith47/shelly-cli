package inspect

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}
	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are required")
	}
	if cmd.Example == "" {
		t.Error("Example is required")
	}
}

func TestNewCommand_APIPFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Flags().Lookup("ap-ip") == nil {
		t.Error("--ap-ip flag not registered")
	}
}

func TestNewCommand_RequiresOneArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error with two args")
	}
	if err := cmd.Args(cmd, []string{"ShellyBulbDuo-D0DCFF"}); err != nil {
		t.Errorf("expected no error with one arg, got %v", err)
	}
}

func TestNewCommand_LongMentionsAP(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Long), "access point") {
		t.Error("Long description should explain the AP hop")
	}
}
