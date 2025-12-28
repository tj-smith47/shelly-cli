package config

import (
	"os"
	"path/filepath"
	"testing"
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
