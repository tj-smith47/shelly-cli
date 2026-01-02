package set

import (
	"bytes"
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

	if cmd.Use != "set <key>=<value>..." {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <key>=<value>...")
	}

	wantAliases := []string{"write", "update"}
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
		{"one arg valid", []string{"key=value"}, false},
		{"multiple args", []string{"key1=value1", "key2=value2"}, false},
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
		"shelly config set",
		"defaults.timeout",
		"defaults.output",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for setting completion")
	}
}

func TestRun_InvalidFormat(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	opts := &Options{Factory: tf.Factory, Args: []string{"invalid-no-equals"}}
	err := run(opts)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Error should mention invalid format: %v", err)
	}
}

func TestRun_ValidSetting(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// This will fail because config.SetSetting needs a real config setup
	// but we're testing the run function's parsing logic
	opts := &Options{Factory: tf.Factory, Args: []string{"defaults.timeout=30s"}}
	err := run(opts)
	// We expect either success or a config-related error, not a parsing error
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Should not get invalid format error for valid format: %v", err)
	}
}

func TestRun_MultipleSettings(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	opts := &Options{Factory: tf.Factory, Args: []string{"key1=value1", "key2=value2"}}
	err := run(opts)
	// Check that parsing is correct
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Should not get invalid format error for valid formats: %v", err)
	}
}

func TestRun_EmptyValue(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// key= should be valid (empty value)
	opts := &Options{Factory: tf.Factory, Args: []string{"key="}}
	err := run(opts)
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Should not get invalid format error for key=: %v", err)
	}
}

func TestRun_ValueWithEquals(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// key=value=with=equals should be valid
	opts := &Options{Factory: tf.Factory, Args: []string{"key=value=with=equals"}}
	err := run(opts)
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Should not get invalid format error for value with equals: %v", err)
	}
}
