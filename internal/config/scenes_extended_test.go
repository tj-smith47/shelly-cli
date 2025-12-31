package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

const testSceneMovieNight = "movie-night"

func TestParseSceneFile_JSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "scene.json")
	content := `{
		"name": "movie-night",
		"description": "Dim lights for movies",
		"actions": [
			{
				"device": "living-room-light",
				"method": "Light.Set",
				"params": {"brightness": 20}
			}
		]
	}`
	if err := os.WriteFile(jsonFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	scene, err := ParseSceneFile(jsonFile)
	if err != nil {
		t.Fatalf("ParseSceneFile() error: %v", err)
	}

	if scene.Name != testSceneMovieNight {
		t.Errorf("Name = %q, want %q", scene.Name, testSceneMovieNight)
	}
	if scene.Description != "Dim lights for movies" {
		t.Errorf("Description = %q, want %q", scene.Description, "Dim lights for movies")
	}
	if len(scene.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(scene.Actions))
	}
	if scene.Actions[0].Device != "living-room-light" {
		t.Errorf("Actions[0].Device = %q, want %q", scene.Actions[0].Device, "living-room-light")
	}
}

func TestParseSceneFile_YAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "scene.yaml")
	content := `name: movie-night
description: Dim lights for movies
actions:
  - device: living-room-light
    method: Light.Set
    params:
      brightness: 20
`
	if err := os.WriteFile(yamlFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	scene, err := ParseSceneFile(yamlFile)
	if err != nil {
		t.Fatalf("ParseSceneFile() error: %v", err)
	}

	if scene.Name != testSceneMovieNight {
		t.Errorf("Name = %q, want %q", scene.Name, testSceneMovieNight)
	}
	if len(scene.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(scene.Actions))
	}
}

func TestParseSceneFile_YML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	ymlFile := filepath.Join(tmpDir, "scene.yml")
	content := `name: test-scene
description: Test
actions: []
`
	if err := os.WriteFile(ymlFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	scene, err := ParseSceneFile(ymlFile)
	if err != nil {
		t.Fatalf("ParseSceneFile() error: %v", err)
	}

	if scene.Name != "test-scene" {
		t.Errorf("Name = %q, want %q", scene.Name, "test-scene")
	}
}

func TestParseSceneFile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ParseSceneFile("/nonexistent/path/scene.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseSceneFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "scene.json")
	if err := os.WriteFile(jsonFile, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := ParseSceneFile(jsonFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseSceneFile_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "scene.yaml")
	if err := os.WriteFile(yamlFile, []byte(":\ninvalid yaml"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := ParseSceneFile(yamlFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestImportScene_EmptyName(t *testing.T) {
	t.Parallel()

	scene := &Scene{
		Name: "",
	}

	// ImportScene uses the default manager which doesn't support isolation in tests,
	// but we can still test the validation logic
	if err := ImportScene(scene, false); err == nil {
		t.Error("expected error importing scene with empty name")
	}
}

// setupScenesTest sets up an isolated environment for scene package-level function tests.
func setupScenesTest(t *testing.T) {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_CreateScene(t *testing.T) {
	setupScenesTest(t)

	err := CreateScene("test-scene", "Test description")
	if err != nil {
		t.Errorf("CreateScene() error = %v", err)
	}

	// Verify scene was created
	scene, ok := GetScene("test-scene")
	if !ok {
		t.Fatal("GetScene() should find created scene")
	}
	if scene.Name != "test-scene" {
		t.Errorf("scene.Name = %q, want %q", scene.Name, "test-scene")
	}
	if scene.Description != "Test description" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "Test description")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_ListScenes(t *testing.T) {
	setupScenesTest(t)

	// Initially empty
	scenes := ListScenes()
	if len(scenes) != 0 {
		t.Errorf("ListScenes() should be empty, got %d", len(scenes))
	}

	// Create some scenes
	if err := CreateScene("scene1", ""); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}
	if err := CreateScene("scene2", ""); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}

	scenes = ListScenes()
	if len(scenes) != 2 {
		t.Errorf("ListScenes() should return 2 scenes, got %d", len(scenes))
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_DeleteScene(t *testing.T) {
	setupScenesTest(t)

	// Create a scene
	if err := CreateScene("to-delete", ""); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}

	// Delete it
	if err := DeleteScene("to-delete"); err != nil {
		t.Errorf("DeleteScene() error = %v", err)
	}

	// Verify it's gone
	_, ok := GetScene("to-delete")
	if ok {
		t.Error("GetScene() should not find deleted scene")
	}

	// Delete non-existent should error
	if err := DeleteScene("nonexistent"); err == nil {
		t.Error("DeleteScene() should error for non-existent scene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_AddActionToScene(t *testing.T) {
	setupScenesTest(t)

	// Create a scene
	if err := CreateScene("action-test", ""); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}

	// Add an action
	action := SceneAction{
		Device: "light1",
		Method: "Switch.Set",
		Params: map[string]any{"on": true},
	}
	if err := AddActionToScene("action-test", action); err != nil {
		t.Errorf("AddActionToScene() error = %v", err)
	}

	scene, ok := GetScene("action-test")
	if !ok {
		t.Fatal("GetScene() should find scene")
	}
	if len(scene.Actions) != 1 {
		t.Errorf("scene.Actions length = %d, want 1", len(scene.Actions))
	}
	if scene.Actions[0].Device != "light1" {
		t.Errorf("scene.Actions[0].Device = %q, want %q", scene.Actions[0].Device, "light1")
	}

	// Add to non-existent scene should error
	if err := AddActionToScene("nonexistent", action); err == nil {
		t.Error("AddActionToScene() should error for non-existent scene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_SetSceneActions(t *testing.T) {
	setupScenesTest(t)

	// Create a scene with actions
	if err := CreateScene("actions-test", ""); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}

	// Set actions
	actions := []SceneAction{
		{Device: "light1", Method: "Switch.Set", Params: map[string]any{"on": true}},
		{Device: "light2", Method: "Switch.Set", Params: map[string]any{"on": false}},
	}
	if err := SetSceneActions("actions-test", actions); err != nil {
		t.Errorf("SetSceneActions() error = %v", err)
	}

	scene, ok := GetScene("actions-test")
	if !ok {
		t.Fatal("GetScene() should find scene")
	}
	if len(scene.Actions) != 2 {
		t.Errorf("scene.Actions length = %d, want 2", len(scene.Actions))
	}

	// Set on non-existent scene should error
	if err := SetSceneActions("nonexistent", actions); err == nil {
		t.Error("SetSceneActions() should error for non-existent scene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_UpdateScene(t *testing.T) {
	setupScenesTest(t)

	// Create a scene
	if err := CreateScene("update-test", "Original description"); err != nil {
		t.Fatalf("CreateScene() error = %v", err)
	}

	// Update description only
	if err := UpdateScene("update-test", "", "New description"); err != nil {
		t.Errorf("UpdateScene() error = %v", err)
	}

	scene, ok := GetScene("update-test")
	if !ok {
		t.Fatal("GetScene() should find scene")
	}
	if scene.Description != "New description" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "New description")
	}

	// Update name
	if err := UpdateScene("update-test", "renamed-scene", ""); err != nil {
		t.Errorf("UpdateScene() rename error = %v", err)
	}

	_, ok = GetScene("update-test")
	if ok {
		t.Error("old scene name should not exist")
	}

	scene, ok = GetScene("renamed-scene")
	if !ok {
		t.Fatal("renamed scene should exist")
	}

	// Update non-existent should error
	if err := UpdateScene("nonexistent", "", ""); err == nil {
		t.Error("UpdateScene() should error for non-existent scene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_SaveScene(t *testing.T) {
	setupScenesTest(t)

	scene := Scene{
		Name:        "saved-scene",
		Description: "Saved via SaveScene",
		Actions: []SceneAction{
			{Device: "dev1", Method: "Switch.On"},
		},
	}

	if err := SaveScene(scene); err != nil {
		t.Errorf("SaveScene() error = %v", err)
	}

	loaded, ok := GetScene("saved-scene")
	if !ok {
		t.Fatal("GetScene() should find saved scene")
	}
	if loaded.Description != "Saved via SaveScene" {
		t.Errorf("scene.Description = %q, want %q", loaded.Description, "Saved via SaveScene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportScene_WithOverwrite(t *testing.T) {
	setupScenesTest(t)

	// Create initial scene
	scene1 := &Scene{
		Name:        "import-test",
		Description: "Original",
		Actions:     []SceneAction{},
	}
	if err := ImportScene(scene1, false); err != nil {
		t.Fatalf("ImportScene() initial error = %v", err)
	}

	// Try to import again without overwrite - should fail
	scene2 := &Scene{
		Name:        "import-test",
		Description: "Updated",
		Actions:     []SceneAction{},
	}
	if err := ImportScene(scene2, false); err == nil {
		t.Error("ImportScene() should error when scene exists and overwrite=false")
	}

	// Import with overwrite - should succeed
	if err := ImportScene(scene2, true); err != nil {
		t.Errorf("ImportScene() with overwrite error = %v", err)
	}

	scene, ok := GetScene("import-test")
	if !ok {
		t.Fatal("GetScene() should find scene")
	}
	if scene.Description != "Updated" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "Updated")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportScene_WithActions(t *testing.T) {
	setupScenesTest(t)

	scene := &Scene{
		Name: "with-actions",
		Actions: []SceneAction{
			{Device: "dev1", Method: "Switch.On"},
			{Device: "dev2", Method: "Switch.Off"},
		},
	}
	if err := ImportScene(scene, false); err != nil {
		t.Fatalf("ImportScene() error = %v", err)
	}

	loaded, ok := GetScene("with-actions")
	if !ok {
		t.Fatal("GetScene() should find scene")
	}
	if len(loaded.Actions) != 2 {
		t.Errorf("scene.Actions length = %d, want 2", len(loaded.Actions))
	}
}

func TestParseSceneFile_UnknownExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Use .txt extension which will trigger unknown format parsing
	txtFile := filepath.Join(tmpDir, "scene.txt")
	content := `name: unknown-ext
description: Test unknown extension
actions: []
`
	if err := os.WriteFile(txtFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	scene, err := ParseSceneFile(txtFile)
	if err != nil {
		t.Fatalf("ParseSceneFile() error: %v", err)
	}

	if scene.Name != "unknown-ext" {
		t.Errorf("Name = %q, want %q", scene.Name, "unknown-ext")
	}
}

func TestParseSceneFile_UnknownExtensionFallbackJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Use .txt extension with JSON content
	txtFile := filepath.Join(tmpDir, "scene.txt")
	content := `{"name": "json-fallback", "actions": []}`
	if err := os.WriteFile(txtFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	scene, err := ParseSceneFile(txtFile)
	if err != nil {
		t.Fatalf("ParseSceneFile() error: %v", err)
	}

	if scene.Name != "json-fallback" {
		t.Errorf("Name = %q, want %q", scene.Name, "json-fallback")
	}
}

func TestParseSceneFile_UnknownExtensionInvalid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Use .txt extension with invalid content
	txtFile := filepath.Join(tmpDir, "scene.txt")
	content := `this is not valid yaml or json {{ }}`
	if err := os.WriteFile(txtFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := ParseSceneFile(txtFile)
	if err == nil {
		t.Error("expected error for invalid content in unknown extension")
	}
}
