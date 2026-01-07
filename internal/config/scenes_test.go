package config

import (
	"testing"

	"github.com/spf13/afero"
)

func TestValidateSceneName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "movie-night", false},
		{"valid with underscore", "wake_up", false},
		{"valid alphanumeric", "scene123", false},
		{"empty name", "", true},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"starts with hyphen", "-invalid", true},
		{"contains space", "movie night", true},
		{"contains special char", "movie@night", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSceneName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSceneName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSceneStruct(t *testing.T) {
	t.Parallel()

	scene := Scene{
		Name:        "movie-night",
		Description: "Dim lights for movie watching",
		Actions: []SceneAction{
			{Device: "living-room-light", Method: "Light.Set", Params: map[string]any{"brightness": 20}},
			{Device: "tv-backlight", Method: "Switch.Set", Params: map[string]any{"on": true}},
		},
	}

	if scene.Name != "movie-night" {
		t.Errorf("Name = %q, want movie-night", scene.Name)
	}
	if len(scene.Actions) != 2 {
		t.Errorf("Actions length = %d, want 2", len(scene.Actions))
	}
	if scene.Actions[0].Device != "living-room-light" {
		t.Errorf("Action[0].Device = %q, want living-room-light", scene.Actions[0].Device)
	}
	if scene.Actions[0].Method != "Light.Set" {
		t.Errorf("Action[0].Method = %q, want Light.Set", scene.Actions[0].Method)
	}
}

func TestSceneActionStruct(t *testing.T) {
	t.Parallel()

	action := SceneAction{
		Device: "kitchen-switch",
		Method: "Switch.Toggle",
		Params: map[string]any{"id": 0},
	}

	if action.Device != "kitchen-switch" {
		t.Errorf("Device = %q, want kitchen-switch", action.Device)
	}
	if action.Method != "Switch.Toggle" {
		t.Errorf("Method = %q, want Switch.Toggle", action.Method)
	}
	if action.Params["id"] != 0 {
		t.Errorf("Params[id] = %v, want 0", action.Params["id"])
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_JSON(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	jsonFile := "/scene.json"
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
	if err := afero.WriteFile(Fs(), jsonFile, []byte(content), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_YAML(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	yamlFile := "/scene.yaml"
	content := `name: movie-night
description: Dim lights for movies
actions:
  - device: living-room-light
    method: Light.Set
    params:
      brightness: 20
`
	if err := afero.WriteFile(Fs(), yamlFile, []byte(content), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_YML(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	ymlFile := "/scene.yml"
	content := `name: test-scene
description: Test
actions: []
`
	if err := afero.WriteFile(Fs(), ymlFile, []byte(content), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_FileNotFound(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	_, err := ParseSceneFile("/nonexistent/path/scene.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_InvalidJSON(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	jsonFile := "/scene.json"
	if err := afero.WriteFile(Fs(), jsonFile, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := ParseSceneFile(jsonFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_InvalidYAML(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	yamlFile := "/scene.yaml"
	if err := afero.WriteFile(Fs(), yamlFile, []byte(":\ninvalid yaml"), 0o600); err != nil {
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

	_, ok = GetScene("renamed-scene")
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_UnknownExtension(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	// Use .txt extension which will trigger unknown format parsing
	txtFile := testSceneTxtFile
	content := `name: unknown-ext
description: Test unknown extension
actions: []
`
	if err := afero.WriteFile(Fs(), txtFile, []byte(content), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_UnknownExtensionFallbackJSON(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	// Use .txt extension with JSON content
	txtFile := testSceneTxtFile
	content := `{"name": "json-fallback", "actions": []}`
	if err := afero.WriteFile(Fs(), txtFile, []byte(content), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestParseSceneFile_UnknownExtensionInvalid(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	// Use .txt extension with invalid content
	txtFile := testSceneTxtFile
	content := `this is not valid yaml or json {{ }}`
	if err := afero.WriteFile(Fs(), txtFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := ParseSceneFile(txtFile)
	if err == nil {
		t.Error("expected error for invalid content in unknown extension")
	}
}

//nolint:paralleltest // Tests modify global state
func TestExportSceneToFile(t *testing.T) {
	setupScenesTest(t)

	// Create a scene to export
	scene := Scene{
		Name:        "export-test",
		Description: "Test export",
		Actions: []SceneAction{
			{Device: "dev1", Method: "Switch.On", Params: map[string]any{"id": 0}},
		},
	}
	if err := SaveScene(scene); err != nil {
		t.Fatalf("SaveScene() error = %v", err)
	}

	// Export to file
	outputPath := "/export/scenes/export-test.json"
	if err := Fs().MkdirAll("/export/scenes", 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	filePath, err := ExportSceneToFile("export-test", outputPath)
	if err != nil {
		t.Fatalf("ExportSceneToFile() error = %v", err)
	}
	if filePath != outputPath {
		t.Errorf("filePath = %q, want %q", filePath, outputPath)
	}

	// Verify file was created and is valid JSON
	data, err := afero.ReadFile(Fs(), outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Parse exported scene
	exported, err := ParseSceneFile(outputPath)
	if err != nil {
		t.Fatalf("ParseSceneFile() error = %v", err)
	}
	if exported.Name != "export-test" {
		t.Errorf("exported.Name = %q, want %q", exported.Name, "export-test")
	}
	if exported.Description != "Test export" {
		t.Errorf("exported.Description = %q, want %q", exported.Description, "Test export")
	}
	if len(exported.Actions) != 1 {
		t.Fatalf("len(exported.Actions) = %d, want 1", len(exported.Actions))
	}

	// Verify JSON is properly formatted with indentation
	if len(data) < 50 {
		t.Errorf("exported JSON seems too short: %s", string(data))
	}
}

//nolint:paralleltest // Tests modify global state
func TestExportSceneToFile_NotFound(t *testing.T) {
	setupScenesTest(t)

	_, err := ExportSceneToFile("nonexistent", "/export/nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent scene")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportSceneFromFile(t *testing.T) {
	setupScenesTest(t)

	// Create a scene file
	sceneFile := "/import/test-scene.json"
	if err := Fs().MkdirAll("/import", 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `{
		"name": "imported-scene",
		"description": "Imported from file",
		"actions": [
			{"device": "dev1", "method": "Switch.On"}
		]
	}`
	if err := afero.WriteFile(Fs(), sceneFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Import the scene
	msg, err := ImportSceneFromFile(sceneFile, "", false)
	if err != nil {
		t.Fatalf("ImportSceneFromFile() error = %v", err)
	}
	if msg == "" {
		t.Error("expected success message")
	}

	// Verify scene was imported
	scene, ok := GetScene("imported-scene")
	if !ok {
		t.Fatal("GetScene() should find imported scene")
	}
	if scene.Description != "Imported from file" {
		t.Errorf("scene.Description = %q, want %q", scene.Description, "Imported from file")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportSceneFromFile_NameOverride(t *testing.T) {
	setupScenesTest(t)

	// Create a scene file
	sceneFile := "/import/original.json"
	if err := Fs().MkdirAll("/import", 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `{"name": "original-name", "actions": []}`
	if err := afero.WriteFile(Fs(), sceneFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Import with name override
	_, err := ImportSceneFromFile(sceneFile, "overridden-name", false)
	if err != nil {
		t.Fatalf("ImportSceneFromFile() error = %v", err)
	}

	// Verify scene was imported with overridden name
	_, ok := GetScene("original-name")
	if ok {
		t.Error("should not find scene with original name")
	}

	_, ok = GetScene("overridden-name")
	if !ok {
		t.Fatal("should find scene with overridden name")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportSceneFromFile_NoName(t *testing.T) {
	setupScenesTest(t)

	// Create a scene file with no name
	sceneFile := "/import/noname.json"
	if err := Fs().MkdirAll("/import", 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `{"description": "No name", "actions": []}`
	if err := afero.WriteFile(Fs(), sceneFile, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Import should fail
	_, err := ImportSceneFromFile(sceneFile, "", false)
	if err == nil {
		t.Error("expected error for scene with no name")
	}
}
