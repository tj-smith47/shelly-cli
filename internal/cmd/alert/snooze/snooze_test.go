package snooze

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// setupTest initializes the test environment with isolated filesystem.
// Returns a cleanup function that should be deferred.
//
//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func setupTest(t *testing.T) {
	t.Helper()
	factory.SetupTestFs(t)
}

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

	if cmd.Use != "snooze <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "snooze <name>")
	}

	wantAliases := []string{"mute", "silence"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"alert-name"}, false},
		{"two args", []string{"alert1", "alert2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"duration", "d", "1h"},
		{"clear", "", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("--%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("--%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("--%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly alert snooze",
		"--duration",
		"--clear",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestExecute_AlertNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-alert"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent alert")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found': %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_InvalidDuration(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Set up config with an alert (in-memory)
	tf.Config.Alerts = map[string]config.Alert{
		"test-alert": {
			Name:      "test-alert",
			Condition: "offline",
			Enabled:   true,
		},
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-alert", "--duration", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid duration")
	}
	if !strings.Contains(err.Error(), "invalid duration") {
		t.Errorf("Error should mention 'invalid duration': %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_SnoozeSuccess(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Set up config with an alert (in-memory)
	tf.Config.Alerts = map[string]config.Alert{
		"test-alert": {
			Name:      "test-alert",
			Condition: "offline",
			Enabled:   true,
		},
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-alert", "--duration", "1h"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_ClearSnooze(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Set up config with a snoozed alert (in-memory)
	tf.Config.Alerts = map[string]config.Alert{
		"test-alert": {
			Name:         "test-alert",
			Condition:    "offline",
			Enabled:      true,
			SnoozedUntil: "2099-01-01T00:00:00Z",
		},
	}

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-alert", "--clear"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Duration: "2h",
		Clear:    true,
	}

	if opts.Duration != "2h" {
		t.Errorf("Duration = %q, want %q", opts.Duration, "2h")
	}

	if !opts.Clear {
		t.Error("Clear should be true")
	}
}
