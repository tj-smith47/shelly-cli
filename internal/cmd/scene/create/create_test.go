package create

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "create <name>" {
		t.Errorf("Use = %q, want \"create <name>\"", cmd.Use)
	}
	aliases := []string{"new"}
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "extra"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	descFlag := cmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Fatal("description flag not found")
	}
	if descFlag.Shorthand != "d" {
		t.Errorf("description shorthand = %q, want d", descFlag.Shorthand)
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_Success(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "new-scene",
	}

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify scene was created
	scene, exists := config.GetScene("new-scene")
	if !exists {
		t.Error("scene should have been created")
	}
	if scene.Name != "new-scene" {
		t.Errorf("scene.Name = %q, want %q", scene.Name, "new-scene")
	}

	// Check output
	allOutput := tf.OutString() + tf.ErrString()
	if !strings.Contains(allOutput, "created") {
		t.Errorf("output = %q, want to contain 'created'", allOutput)
	}
	if !strings.Contains(allOutput, "add-action") {
		t.Errorf("output = %q, want to contain 'add-action'", allOutput)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_WithDescription(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:     tf.Factory,
		Name:        "scene-with-desc",
		Description: "My test scene description",
	}

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify scene was created with description
	scene, exists := config.GetScene("scene-with-desc")
	if !exists {
		t.Error("scene should have been created")
	}
	if scene.Description != "My test scene description" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "My test scene description")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_SceneAlreadyExists(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"existing-scene": {
				Name:        "existing-scene",
				Description: "Already exists",
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "existing-scene",
	}

	err := run(opts)

	if err == nil {
		t.Error("expected error for existing scene")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want to contain 'already exists'", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_InvalidSceneName(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "invalid name with spaces",
	}

	err := run(opts)

	if err == nil {
		t.Error("expected error for invalid scene name")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_Execute(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-scene-exec"})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify scene was created
	_, exists := config.GetScene("test-scene-exec")
	if !exists {
		t.Error("scene should have been created")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_ExecuteWithDescription(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-scene-desc", "-d", "Test description"})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify scene was created with description
	scene, exists := config.GetScene("test-scene-desc")
	if !exists {
		t.Error("scene should have been created")
	}
	if scene.Description != "Test description" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "Test description")
	}
}
