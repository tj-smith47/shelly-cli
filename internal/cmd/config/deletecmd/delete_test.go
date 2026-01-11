// Package deletecmd provides the config delete subcommand for CLI settings.
package deletecmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

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

	if cmd.Use != "delete <key>..." {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <key>...")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	wantAliases := []string{"del", "rm", "remove", "unset"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}
	for i, want := range wantAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_YesFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test --yes flag exists
	yesFlag := cmd.Flags().Lookup("yes")
	switch {
	case yesFlag == nil:
		t.Error("--yes flag not found")
	case yesFlag.Shorthand != "y":
		t.Errorf("--yes shorthand = %q, want %q", yesFlag.Shorthand, "y")
	case yesFlag.DefValue != "false":
		t.Errorf("--yes default = %q, want %q", yesFlag.DefValue, "false")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require at least 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with no args
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Test with 1 arg
	err = cmd.Args(cmd, []string{"key1"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}

	// Test with multiple args
	err = cmd.Args(cmd, []string{"key1", "key2", "key3"})
	if err != nil {
		t.Errorf("Expected no error with multiple args, got: %v", err)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for setting key completion")
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
		"shelly config delete",
		"defaults.timeout",
		"--yes",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Keys:    []string{"key1", "key2"},
	}

	if len(opts.Keys) != 2 {
		t.Errorf("Keys length = %d, want 2", len(opts.Keys))
	}

	// Verify ConfirmFlags is embedded and accessible
	opts.Yes = true
	if !opts.Yes {
		t.Error("Yes should be true")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs and viper
func TestRun_DeleteSingleKey(t *testing.T) {
	// Set up in-memory filesystem
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create config directory and file
	configDir := "/tmp/shelly-test"
	configFile := filepath.Join(configDir, "config.yaml")
	if err := fs.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write initial config
	initialConfig := `editor: vim
defaults:
  timeout: 30s
  output: json
`
	if err := afero.WriteFile(fs, configFile, []byte(initialConfig), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Set up viper to use this config file
	viper.Reset()
	viper.SetFs(fs)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	// Verify the key exists
	if !viper.IsSet("editor") {
		t.Fatal("editor key should be set before deletion")
	}

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Keys:    []string{"editor"},
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Read back the config file and verify key was removed
	data, err := afero.ReadFile(fs, configFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if strings.Contains(string(data), "editor:") {
		t.Error("editor key should be removed from config file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs and viper
func TestRun_KeyNotSet(t *testing.T) {
	// Set up in-memory filesystem
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create config directory and file
	configDir := "/tmp/shelly-test"
	configFile := filepath.Join(configDir, "config.yaml")
	if err := fs.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write minimal config
	if err := afero.WriteFile(fs, configFile, []byte("theme: dracula\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Set up viper
	viper.Reset()
	viper.SetFs(fs)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Keys:    []string{"nonexistent.key"},
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for nonexistent key")
	}
	if !strings.Contains(err.Error(), "not set") {
		t.Errorf("Error should mention key not set: %v", err)
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"Delete",
		"configuration",
		"dot notation",
		"nested",
		"--yes",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}
