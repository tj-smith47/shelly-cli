package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateSceneName checks if a scene name is valid.
func ValidateSceneName(name string) error {
	return ValidateName(name, "scene")
}

// Package-level functions delegate to the default manager.

// CreateScene creates a new scene.
func CreateScene(name, description string) error {
	return getDefaultManager().CreateScene(name, description)
}

// DeleteScene removes a scene.
func DeleteScene(name string) error {
	return getDefaultManager().DeleteScene(name)
}

// GetScene returns a scene by name.
func GetScene(name string) (Scene, bool) {
	return getDefaultManager().GetScene(name)
}

// ListScenes returns all scenes.
func ListScenes() map[string]Scene {
	return getDefaultManager().ListScenes()
}

// AddActionToScene adds an action to a scene.
func AddActionToScene(sceneName string, action SceneAction) error {
	return getDefaultManager().AddActionToScene(sceneName, action)
}

// SetSceneActions replaces all actions in a scene.
func SetSceneActions(sceneName string, actions []SceneAction) error {
	return getDefaultManager().SetSceneActions(sceneName, actions)
}

// UpdateScene updates a scene's name and/or description.
func UpdateScene(oldName, newName, description string) error {
	return getDefaultManager().UpdateScene(oldName, newName, description)
}

// SaveScene saves or updates a scene (used by import).
func SaveScene(scene Scene) error {
	return getDefaultManager().SaveScene(scene)
}

// ParseSceneFile parses a scene definition from a file.
// Format is auto-detected from file extension (.json, .yaml, .yml).
// If extension is unknown, it tries YAML first, then JSON.
func ParseSceneFile(file string) (*Scene, error) {
	// #nosec G304 -- file path comes from user CLI argument
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var scene Scene

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ExtJSON:
		if err := json.Unmarshal(data, &scene); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ExtYAML, ExtYML:
		if err := yaml.Unmarshal(data, &scene); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		if err := parseSceneUnknownFormat(data, &scene); err != nil {
			return nil, err
		}
	}

	return &scene, nil
}

func parseSceneUnknownFormat(data []byte, scene *Scene) error {
	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, scene); err != nil {
		if jsonErr := json.Unmarshal(data, scene); jsonErr != nil {
			return fmt.Errorf("failed to parse file (tried YAML and JSON)")
		}
	}
	return nil
}

// ImportScene imports a scene, optionally overwriting an existing one.
// Returns an error if the scene exists and overwrite is false.
func ImportScene(scene *Scene, overwrite bool) error {
	if scene.Name == "" {
		return fmt.Errorf("scene name is required")
	}

	// Check if scene exists
	if _, exists := GetScene(scene.Name); exists {
		if !overwrite {
			return fmt.Errorf("scene %q already exists (use --overwrite to replace)", scene.Name)
		}
		// Delete existing scene first
		if err := DeleteScene(scene.Name); err != nil {
			return fmt.Errorf("failed to delete existing scene: %w", err)
		}
	}

	// Create scene
	if err := CreateScene(scene.Name, scene.Description); err != nil {
		return fmt.Errorf("failed to create scene: %w", err)
	}

	// Add actions
	if len(scene.Actions) > 0 {
		if err := SetSceneActions(scene.Name, scene.Actions); err != nil {
			return fmt.Errorf("failed to set scene actions: %w", err)
		}
	}

	return nil
}
