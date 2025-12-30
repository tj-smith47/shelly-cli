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

// =============================================================================
// Manager Scene Methods
// =============================================================================

// CreateScene creates a new scene.
func (m *Manager) CreateScene(name, description string) error {
	if err := ValidateSceneName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Scenes[name]; exists {
		return fmt.Errorf("scene %q already exists", name)
	}

	m.config.Scenes[name] = Scene{
		Name:        name,
		Description: description,
		Actions:     []SceneAction{},
	}
	return m.saveWithoutLock()
}

// DeleteScene removes a scene.
func (m *Manager) DeleteScene(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Scenes[name]; !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	delete(m.config.Scenes, name)
	return m.saveWithoutLock()
}

// GetScene returns a scene by name.
func (m *Manager) GetScene(name string) (Scene, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scene, ok := m.config.Scenes[name]
	return scene, ok
}

// ListScenes returns all scenes.
func (m *Manager) ListScenes() map[string]Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Scene, len(m.config.Scenes))
	for k, v := range m.config.Scenes {
		result[k] = v
	}
	return result
}

// AddActionToScene adds an action to a scene.
func (m *Manager) AddActionToScene(sceneName string, action SceneAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = append(scene.Actions, action)
	m.config.Scenes[sceneName] = scene
	return m.saveWithoutLock()
}

// SetSceneActions replaces all actions in a scene.
func (m *Manager) SetSceneActions(sceneName string, actions []SceneAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = actions
	m.config.Scenes[sceneName] = scene
	return m.saveWithoutLock()
}

// UpdateScene updates a scene's name and/or description.
func (m *Manager) UpdateScene(oldName, newName, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[oldName]
	if !exists {
		return fmt.Errorf("scene %q not found", oldName)
	}

	if newName != "" && newName != oldName {
		if err := ValidateSceneName(newName); err != nil {
			return err
		}
		if _, exists := m.config.Scenes[newName]; exists {
			return fmt.Errorf("scene %q already exists", newName)
		}
		delete(m.config.Scenes, oldName)
		scene.Name = newName
	}

	if description != "" {
		scene.Description = description
	}

	m.config.Scenes[scene.Name] = scene
	return m.saveWithoutLock()
}

// SaveScene saves or updates a scene (used by import).
func (m *Manager) SaveScene(scene Scene) error {
	if err := ValidateSceneName(scene.Name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Scenes[scene.Name] = scene
	return m.saveWithoutLock()
}
