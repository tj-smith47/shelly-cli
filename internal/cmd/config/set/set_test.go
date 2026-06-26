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
		"discovery.timeout",
		"output=json",
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

//nolint:paralleltest // run() touches global viper state; keep serial like siblings
func TestRun_BareKeyMissingValue(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{Factory: tf.Factory, Args: []string{"telemetry-no-separator"}}
	err := run(opts)
	if err == nil {
		t.Fatal("Expected error for a bare key with no value")
	}
	if !strings.Contains(err.Error(), "missing value") {
		t.Errorf("Error should mention the missing value: %v", err)
	}
}

// parseErr reports whether err is a key/value parse error (as opposed to a
// downstream config-write error, which is acceptable in these tests because
// SetSetting needs a real config file).
func parseErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "missing value") ||
		strings.Contains(msg, "empty key") ||
		strings.Contains(msg, "no key/value pairs")
}

//nolint:paralleltest // modifies global viper state
func TestRun_SeparatorEquivalence(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// "=", ":", and a space must all be accepted for the same assignment.
	cases := [][]string{
		{"discovery.timeout=30s"},
		{"discovery.timeout:30s"},
		{"discovery.timeout", "30s"},
	}
	for _, args := range cases {
		opts := &Options{Factory: tf.Factory, Args: args}
		if err := run(opts); parseErr(err) {
			t.Errorf("args %q should parse, got parse error: %v", args, err)
		}
	}
}

//nolint:paralleltest // modifies global viper state
func TestRun_MultipleSettings(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{Factory: tf.Factory, Args: []string{"editor=vim", "theme.name=dark"}}
	if err := run(opts); parseErr(err) {
		t.Errorf("multiple inline pairs should parse: %v", err)
	}
}

//nolint:paralleltest // modifies global viper state
func TestRun_ValueWithEquals(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// editor=value=with=equals should keep everything after the first separator.
	opts := &Options{Factory: tf.Factory, Args: []string{"editor=code --wait=true"}}
	if err := run(opts); parseErr(err) {
		t.Errorf("value containing '=' should parse: %v", err)
	}
}
