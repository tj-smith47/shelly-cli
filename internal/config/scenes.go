package config

import (
	"fmt"
	"regexp"
)

// sceneNameRegex validates scene names (alphanumeric, hyphens, underscores).
var sceneNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateSceneName checks if a scene name is valid.
func ValidateSceneName(name string) error {
	if name == "" {
		return fmt.Errorf("scene name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("scene name too long (max 64 characters)")
	}
	if !sceneNameRegex.MatchString(name) {
		return fmt.Errorf("scene name contains invalid characters")
	}
	return nil
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
