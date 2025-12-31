package importcmd

import (
	"os"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want \"import <file>\"", cmd.Use)
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
	if err := cmd.Args(cmd, []string{"file.yaml"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"file1", "file2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("name flag not found")
	}
	if nameFlag.Shorthand != "n" {
		t.Errorf("name shorthand = %q, want n", nameFlag.Shorthand)
	}

	overwriteFlag := cmd.Flags().Lookup("overwrite")
	if overwriteFlag == nil {
		t.Fatal("overwrite flag not found")
	}
	if overwriteFlag.DefValue != "false" {
		t.Errorf("overwrite default = %q, want false", overwriteFlag.DefValue)
	}
}

func TestImportScene_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := importScene("/nonexistent/scene.yaml", "", false)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestImportScene_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/invalid.yaml"
	if err := os.WriteFile(file, []byte("name: [invalid yaml"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importScene(file, "", false)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestImportScene_MissingName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/no-name.yaml"
	content := `actions:
  - device: "living-room"
    action: "on"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importScene(file, "", false)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error = %v, want to contain 'name is required'", err)
	}
}

func TestImportScene_NameOverride(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/scene.yaml"
	content := `name: "original-name"
actions:
  - device: "living-room"
    action: "on"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importScene(file, "overridden-name", false)
	if err != nil {
		t.Fatalf("importScene() error = %v", err)
	}
	if !strings.Contains(result, "overridden-name") {
		t.Errorf("result = %q, want to contain 'overridden-name'", result)
	}

	// Verify scene was saved with overridden name
	scene, exists := config.GetScene("overridden-name")
	if !exists {
		t.Error("scene 'overridden-name' should exist")
	}
	if scene.Name != "overridden-name" {
		t.Errorf("scene.Name = %q, want %q", scene.Name, "overridden-name")
	}
}

func TestImportScene_Success(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/scene.yaml"
	content := `name: "movie-night"
description: "Turn off lights for movie time"
actions:
  - device: "living-room"
    action: "off"
  - device: "kitchen"
    action: "off"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importScene(file, "", false)
	if err != nil {
		t.Fatalf("importScene() error = %v", err)
	}
	if !strings.Contains(result, "movie-night") {
		t.Errorf("result = %q, want to contain 'movie-night'", result)
	}
	if !strings.Contains(result, "2 action(s)") {
		t.Errorf("result = %q, want to contain '2 action(s)'", result)
	}

	// Verify scene was saved
	scene, exists := config.GetScene("movie-night")
	if !exists {
		t.Error("scene 'movie-night' should exist")
	}
	if len(scene.Actions) != 2 {
		t.Errorf("len(scene.Actions) = %d, want 2", len(scene.Actions))
	}
}

func TestImportScene_AlreadyExists(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/scene.yaml"
	content := `name: "existing-scene"
actions:
  - device: "device1"
    action: "on"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Import first time
	_, err := importScene(file, "", false)
	if err != nil {
		t.Fatalf("first import error = %v", err)
	}

	// Import second time without overwrite
	_, err = importScene(file, "", false)
	if err == nil {
		t.Error("expected error when importing existing scene without overwrite")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want to contain 'already exists'", err)
	}
}

func TestImportScene_OverwriteWithForce(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/scene.yaml"
	content := `name: "force-scene"
description: "Original"
actions:
  - device: "device1"
    action: "on"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Import first time
	_, err := importScene(file, "", false)
	if err != nil {
		t.Fatalf("first import error = %v", err)
	}

	// Update file
	updatedContent := `name: "force-scene"
description: "Updated"
actions:
  - device: "device1"
    action: "off"
`
	if err := os.WriteFile(file, []byte(updatedContent), 0o600); err != nil {
		t.Fatalf("failed to write updated file: %v", err)
	}

	// Import with overwrite
	result, err := importScene(file, "", true)
	if err != nil {
		t.Fatalf("overwrite import error = %v", err)
	}
	if !strings.Contains(result, "force-scene") {
		t.Errorf("result = %q, want to contain 'force-scene'", result)
	}

	// Verify updated description
	scene, exists := config.GetScene("force-scene")
	if !exists {
		t.Error("scene should exist")
	}
	if scene.Description != "Updated" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "Updated")
	}
}

func TestImportScene_JSONFile(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/scene.json"
	content := `{
  "name": "json-scene",
  "actions": [
    {"device": "light1", "action": "toggle"}
  ]
}`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importScene(file, "", false)
	if err != nil {
		t.Fatalf("importScene() error = %v", err)
	}
	if !strings.Contains(result, "json-scene") {
		t.Errorf("result = %q, want to contain 'json-scene'", result)
	}

	// Verify scene was saved
	_, exists := config.GetScene("json-scene")
	if !exists {
		t.Error("scene 'json-scene' should exist")
	}
}
