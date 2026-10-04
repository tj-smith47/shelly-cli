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

// newAliasDeleteTest returns an alias delete command backed by an in-memory
// config holding the "lights" alias, with that config installed as the default
// manager so the test never reads or writes the live config.
func newAliasDeleteTest(t *testing.T, args ...string) (*config.Manager, *bytes.Buffer, error) {
	t.Helper()
	mgr := config.NewTestManager(&config.Config{
		Aliases: map[string]config.Alias{"lights": {Command: "batch on living-room"}},
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	return mgr, out, cmd.Execute()
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_AliasNotFound(t *testing.T) {
	_, _, err := newAliasDeleteTest(t, "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for non-existent alias")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_YesDeletesWithoutPrompt(t *testing.T) {
	mgr, out, err := newAliasDeleteTest(t, "lights", "--yes")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if mgr.IsAlias("lights") {
		t.Error("alias should have been deleted")
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out.String())
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ShortYesFlag(t *testing.T) {
	mgr, _, err := newAliasDeleteTest(t, "lights", "-y")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if mgr.IsAlias("lights") {
		t.Error("alias should have been deleted")
	}
}

// Without --yes the command asks first; a non-interactive terminal answers
// with the prompt's default (no), so nothing is deleted.
//
//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithoutYesConfirmsFirst(t *testing.T) {
	mgr, out, err := newAliasDeleteTest(t, "lights")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !mgr.IsAlias("lights") {
		t.Error("alias must not be deleted without confirmation")
	}
	if !strings.Contains(out.String(), "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %s", out.String())
	}
}

func TestNewCommand_YesFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("--yes flag not found")
	}
	if flag.Shorthand != "y" {
		t.Errorf("--yes shorthand = %q, want y", flag.Shorthand)
	}
	if !strings.Contains(cmd.Example, "--yes") {
		t.Error("Example should show --yes")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for alias completion")
	}
}
