package set

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
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

	// Test Use
	if cmd.Use != "set <name> <command>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <name> <command>")
	}

	// Test Aliases
	wantAliases := []string{"add", "create"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
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

	// Should require at least 2 args
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Args should reject 0 arguments")
	}
	if err := cmd.Args(cmd, []string{"name"}); err == nil {
		t.Error("Args should reject 1 argument")
	}
	if err := cmd.Args(cmd, []string{"name", "command"}); err != nil {
		t.Errorf("Args should accept 2 arguments: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "command", "with", "spaces"}); err != nil {
		t.Errorf("Args should accept more than 2 arguments: %v", err)
	}
}

func TestNewCommand_ShellFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("shell")
	if flag == nil {
		t.Fatal("--shell flag not found")
	}
	if flag.Shorthand != "s" {
		t.Errorf("--shell shorthand = %q, want %q", flag.Shorthand, "s")
	}
	if flag.DefValue != "false" {
		t.Errorf("--shell default = %q, want %q", flag.DefValue, "false")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_CreateAlias(t *testing.T) {
	factory.SetupTestFs(t)
	config.ResetDefaultManagerForTesting()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"myalias", "device", "list"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "myalias") {
		t.Errorf("expected alias name in output, got: %s", output)
	}
	if !strings.Contains(output, "device list") {
		t.Errorf("expected command in output, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_CreateShellAlias(t *testing.T) {
	factory.SetupTestFs(t)
	config.ResetDefaultManagerForTesting()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--shell", "mybackup", "tar -czf backup.tar.gz"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "shell") {
		t.Errorf("expected 'shell' in output for shell alias, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_ShellAliasWithBangPrefix(t *testing.T) {
	factory.SetupTestFs(t)
	config.ResetDefaultManagerForTesting()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"bangalias", "!echo hello"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	// The ! prefix should make it a shell alias
	if !strings.Contains(output, "shell") {
		t.Errorf("expected 'shell' for ! prefixed command, got: %s", output)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain useful patterns
	wantPatterns := []string{
		"shelly alias set",
		"$1",
		"$@",
		"--shell",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}
