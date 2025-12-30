package deletecmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use - ConfigDeleteCommand uses "delete <name>"
	if !strings.HasPrefix(cmd.Use, "delete") {
		t.Errorf("Use = %q, want prefix 'delete'", cmd.Use)
	}

	// Test Aliases - ConfigDeleteCommand adds standard aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 arg
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Args should reject 0 arguments")
	}
	if err := cmd.Args(cmd, []string{"alias1"}); err != nil {
		t.Errorf("Args should accept 1 argument: %v", err)
	}
	if err := cmd.Args(cmd, []string{"alias1", "alias2"}); err == nil {
		t.Error("Args should reject 2 arguments")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestRun_AliasNotFound(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Aliases: map[string]config.Alias{},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent alias")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for alias completion")
	}
}
