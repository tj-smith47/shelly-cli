package status

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "status [child-device]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status [child-device]")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Accepts 0 or 1 arg
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("unexpected error with no args: %v", err)
	}
	if err := cmd.Args(cmd, []string{"device"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error with 2 args")
	}
}
