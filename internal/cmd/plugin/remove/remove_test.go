package remove

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "remove <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "remove <name>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"rm", "uninstall", "delete"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Remove an installed extension"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "Remove an installed extension") {
		t.Error("Long description should explain remove functionality")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly extension remove") {
		t.Error("Example should contain 'shelly extension remove'")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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
		{
			name:    "no args returns error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg succeeds",
			args:    []string{"myext"},
			wantErr: false,
		},
		{
			name:    "two args returns error",
			args:    []string{"ext1", "ext2"},
			wantErr: true,
		},
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func() bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Use != ""
			},
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Short != ""
			},
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Long != ""
			},
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Example != ""
			},
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return len(cmd.Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.RunE != nil
			},
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc() {
				t.Error(tt.errMsg)
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

func TestExecute_MissingArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestExecute_PluginNotFound(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestExecute_Success(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create a fake plugin in the plugins directory
	pluginsDir := filepath.Join(configDir, "shelly", "plugins")
	pluginDir := filepath.Join(pluginsDir, "shelly-testplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a fake executable
	pluginPath := filepath.Join(pluginDir, "shelly-testplugin")
	//nolint:gosec // G306: test executable needs to be executable
	if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil {
		t.Fatalf("failed to create plugin file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"testplugin"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "Removed") || !strings.Contains(combined, "testplugin") {
		t.Errorf("expected success message about removing testplugin, got: %q", combined)
	}

	// Verify plugin directory was removed
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("plugin directory should be removed")
	}
}

func TestExecute_OldFormat(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create a fake plugin as a bare binary (old format)
	pluginsDir := filepath.Join(configDir, "shelly", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create a fake executable directly in plugins dir (old format)
	pluginPath := filepath.Join(pluginsDir, "shelly-oldplugin")
	//nolint:gosec // G306: test executable needs to be executable
	if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho old"), 0o750); err != nil {
		t.Fatalf("failed to create plugin file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"oldplugin"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify plugin file was removed
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		t.Error("plugin file should be removed")
	}
}

func TestExecute_PluginNotInUserDir(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create the plugins directory (empty)
	pluginsDir := filepath.Join(configDir, "shelly", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create a fake plugin in a different directory and add it to PATH
	otherDir := t.TempDir()
	pluginPath := filepath.Join(otherDir, "shelly-pathplugin")
	//nolint:gosec // G306: test executable needs to be executable
	if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho path"), 0o750); err != nil {
		t.Fatalf("failed to create plugin file: %v", err)
	}

	// Prepend otherDir to PATH
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", otherDir+string(os.PathListSeparator)+origPath)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"pathplugin"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when trying to remove plugin not in user directory")
	}

	if !strings.Contains(err.Error(), "not installed in user plugins directory") {
		t.Errorf("error should mention plugin is not in user directory, got: %v", err)
	}

	// Verify plugin file still exists (wasn't removed)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Error("plugin file should NOT be removed (it's not in user dir)")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for shell completion")
	}
}

func TestRun_Success(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create and install a fake plugin using the registry
	pluginsDir := filepath.Join(configDir, "shelly", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create plugin directory and file (new format)
	pluginDir := filepath.Join(pluginsDir, "shelly-testrun")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}
	pluginPath := filepath.Join(pluginDir, "shelly-testrun")
	//nolint:gosec // G306: test executable needs to be executable
	if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil {
		t.Fatalf("failed to create plugin file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	err := run(f, "testrun")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "Removed") {
		t.Errorf("expected success message, got: %q", combined)
	}
}

func TestRun_PluginNotFound(t *testing.T) {
	// Set XDG_CONFIG_HOME to temp directory
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(nil, stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	err := run(f, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// Verify plugins package constants are as expected.
func TestPluginsConstants(t *testing.T) {
	t.Parallel()

	if plugins.PluginPrefix != "shelly-" {
		t.Errorf("PluginPrefix = %q, want %q", plugins.PluginPrefix, "shelly-")
	}
}
