package reset

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
	if cmd.Use != "reset" {
		t.Errorf("Use = %q, want %q", cmd.Use, "reset")
	}

	// Test Aliases
	wantAliases := []string{"clear"}
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

	// Test Args - should accept no args
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("Args should accept no arguments: %v", err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("Args should reject extra arguments")
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
		t.Errorf("--yes shorthand = %q, want %q", flag.Shorthand, "y")
	}
	if flag.DefValue != "false" {
		t.Errorf("--yes default = %q, want %q", flag.DefValue, "false")
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

// Note: TestRun_WithYes is skipped because config.ResetSettings() accesses the
// filesystem directly and can't be easily mocked. The cancelled path tests the
// run function logic without hitting the filesystem.

func TestRun_Cancelled(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	// Non-TTY mode - confirmation will be denied
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning goes to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "cancelled") {
		t.Errorf("expected 'cancelled' in stderr, got: %s", errOutput)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain useful patterns
	wantPatterns := []string{
		"shelly config reset",
		"--yes",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Yes {
		t.Error("Default Yes should be false")
	}
	if opts.Factory != nil {
		t.Error("Default Factory should be nil")
	}
}

func TestOptions_FieldsSet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
	}
	opts.Yes = true

	if !opts.Yes {
		t.Error("Yes should be true")
	}
	if opts.Factory != f {
		t.Error("Factory should be set")
	}
}
