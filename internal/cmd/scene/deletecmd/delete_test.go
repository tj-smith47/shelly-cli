package deletecmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// setupTestManagerWithScenes creates a test manager with a temp path and populates it with scenes.
func setupTestManagerWithScenes(t *testing.T, scenes map[string]config.Scene) *config.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	mgr := config.NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// Populate scenes directly in config
	cfg := mgr.Get()
	for k, v := range scenes {
		cfg.Scenes[k] = v
	}
	return mgr
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "delete <scene>" {
		t.Errorf("Use = %q, want \"delete <scene>\"", cmd.Use)
	}
	aliases := []string{"rm", "del", "remove"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "delete <scene>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <scene>")
	}

	// Test Aliases
	wantAliases := []string{"rm", "del", "remove"}
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"scene-name"}, false},
		{"two args", []string{"scene1", "scene2"}, true},
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

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want false", yesFlag.DefValue)
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for scene completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly scene delete",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"scene",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

// =============================================================================
// Execute-based Tests
// =============================================================================

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute --help failed: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "delete") {
		t.Errorf("expected help to contain 'delete', got: %s", output)
	}
	if !strings.Contains(output, "scene") {
		t.Errorf("expected help to contain 'scene', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_SceneNotFound(t *testing.T) {
	mgr := setupTestManagerWithScenes(t, map[string]config.Scene{})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent scene")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_DeleteSuccess(t *testing.T) {
	scenes := map[string]config.Scene{
		"movie-night": {
			Name:        "movie-night",
			Description: "Turn off all lights for movies",
			Actions: []config.SceneAction{
				{Device: "living-room", Method: "Switch.Set", Params: map[string]any{"id": 0, "on": false}},
			},
		},
	}
	mgr := setupTestManagerWithScenes(t, scenes)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"movie-night", "--yes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify success message
	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}

	// Verify scene is removed from config
	cfg := mgr.Get()
	if _, exists := cfg.Scenes["movie-night"]; exists {
		t.Error("scene should have been deleted from config")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_DeleteCancelled(t *testing.T) {
	scenes := map[string]config.Scene{
		"movie-night": {
			Name:        "movie-night",
			Description: "Turn off all lights for movies",
			Actions:     []config.SceneAction{},
		},
	}
	mgr := setupTestManagerWithScenes(t, scenes)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	// Without --yes flag, in non-TTY mode confirmation returns false
	cmd.SetArgs([]string{"movie-night"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify cancellation message
	output := out.String()
	if !strings.Contains(output, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %s", output)
	}

	// Verify scene is NOT removed from config
	cfg := mgr.Get()
	if _, exists := cfg.Scenes["movie-night"]; !exists {
		t.Error("scene should NOT have been deleted when cancelled")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_DeleteWithMultipleActions(t *testing.T) {
	scenes := map[string]config.Scene{
		"party-mode": {
			Name:        "party-mode",
			Description: "Turn on party lights",
			Actions: []config.SceneAction{
				{Device: "living-room", Method: "Light.Set", Params: map[string]any{"id": 0, "on": true}},
				{Device: "kitchen", Method: "Light.Set", Params: map[string]any{"id": 0, "on": true}},
				{Device: "bedroom", Method: "Light.Set", Params: map[string]any{"id": 0, "on": true}},
			},
		},
	}
	mgr := setupTestManagerWithScenes(t, scenes)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"party-mode", "--yes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify success message
	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}

	// Verify scene is removed
	cfg := mgr.Get()
	if _, exists := cfg.Scenes["party-mode"]; exists {
		t.Error("scene should have been deleted from config")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error with no arguments")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_DeleteSceneWithNoActions(t *testing.T) {
	scenes := map[string]config.Scene{
		"empty-scene": {
			Name:        "empty-scene",
			Description: "An empty scene",
			Actions:     []config.SceneAction{},
		},
	}
	mgr := setupTestManagerWithScenes(t, scenes)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"empty-scene", "--yes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify success message
	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}

	// Verify scene is removed
	cfg := mgr.Get()
	if _, exists := cfg.Scenes["empty-scene"]; exists {
		t.Error("scene should have been deleted from config")
	}
}
