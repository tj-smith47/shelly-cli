package wait

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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

	if cmd.Use != "wait <duration>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "wait <duration>")
	}

	wantAliases := []string{"delay", "pause"}
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
		{"one arg valid", []string{"5s"}, false},
		{"two args", []string{"5s", "extra"}, true},
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
		"shelly wait",
		"5s",
		"2m",
		"1h",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestRun_ShortDuration(t *testing.T) {
	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, "1ms")
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	out := tf.OutString()
	if !strings.Contains(out, "Done") {
		t.Errorf("Output should contain 'Done', got: %s", out)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, tf.Factory, "1h")
	// With cancelled context, the function should return quickly
	if err != nil {
		t.Logf("run() with cancelled context error = %v (may be expected)", err)
	}
}

func TestRun_InvalidDuration(t *testing.T) {
	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, "invalid")
	if err == nil {
		t.Error("Expected error for invalid duration")
	}
}

func TestRun_VariousDurations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration string
		wantErr  bool
	}{
		{"milliseconds", "1ms", false},
		{"seconds", "1s", false},
		{"minutes", "1m", false},
		{"combined", "1m30s", false},
		{"invalid", "abc", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			ctx, cancel := context.WithCancel(context.Background())
			if !tt.wantErr {
				// For valid durations, cancel immediately to not wait
				cancel()
			}
			defer cancel()

			err := run(ctx, tf.Factory, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("run(%q) error = %v, wantErr %v", tt.duration, err, tt.wantErr)
			}
		})
	}
}
