package next

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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

	if cmd.Use != "next" {
		t.Errorf("Use = %q, want %q", cmd.Use, "next")
	}

	wantAliases := []string{"n"}
	if len(cmd.Aliases) != len(wantAliases) || cmd.Aliases[0] != wantAliases[0] {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

//nolint:paralleltest // Tests modify global theme state
func TestRun_CyclesToNextTheme(t *testing.T) {
	// Save and restore original theme
	originalTheme := theme.Current()
	t.Cleanup(func() {
		theme.SetTheme(originalTheme.ID)
	})

	tf := factory.NewTestFactory(t)
	initialTheme := theme.Current().ID

	opts := &Options{Factory: tf.Factory}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Theme should have changed
	newTheme := theme.Current()
	if newTheme == nil {
		t.Fatal("theme.Current() returned nil")
	}

	// With 280+ themes, it should have cycled to a different one
	if len(theme.ListThemes()) > 1 && newTheme.ID == initialTheme {
		t.Error("expected theme to change")
	}

	// Check output message
	output := tf.TestIO.OutString()
	if !strings.Contains(output, "Theme changed to") {
		t.Errorf("output = %q, want to contain 'Theme changed to'", output)
	}
}

//nolint:paralleltest // Tests modify global theme state
func TestRun_MultipleCycles(t *testing.T) {
	// Save and restore original theme
	originalTheme := theme.Current()
	t.Cleanup(func() {
		theme.SetTheme(originalTheme.ID)
	})

	tf := factory.NewTestFactory(t)

	// Cycle multiple times
	for i := range 5 {
		opts := &Options{Factory: tf.Factory}
		err := run(opts)
		if err != nil {
			t.Fatalf("run() cycle %d error = %v", i+1, err)
		}
	}

	// Verify we end up with a valid theme
	current := theme.Current()
	if current == nil {
		t.Error("theme.Current() returned nil after multiple cycles")
	}
}

//nolint:paralleltest // Test modifies global theme state
func TestNewCommand_Execute(t *testing.T) {
	originalTheme := theme.Current()
	t.Cleanup(func() {
		theme.SetTheme(originalTheme.ID)
	})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
