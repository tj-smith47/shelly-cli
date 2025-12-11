package config

import (
	"fmt"
	"regexp"
)

// sceneNameRegex validates scene names (alphanumeric, hyphens, underscores).
var sceneNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// CreateScene creates a new scene.
func CreateScene(name, description string) error {
	if err := ValidateSceneName(name); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	if _, exists := c.Scenes[name]; exists {
		return fmt.Errorf("scene %q already exists", name)
	}

	c.Scenes[name] = Scene{
		Name:        name,
		Description: description,
		Actions:     []SceneAction{},
	}

	return Save()
}

// DeleteScene removes a scene.
func DeleteScene(name string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	if _, exists := c.Scenes[name]; !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	delete(c.Scenes, name)
	return Save()
}

// GetScene returns a scene by name.
func GetScene(name string) (Scene, bool) {
	c := Get()

	cfgMu.RLock()
	defer cfgMu.RUnlock()

	scene, exists := c.Scenes[name]
	return scene, exists
}

// ListScenes returns all scenes.
func ListScenes() map[string]Scene {
	c := Get()

	cfgMu.RLock()
	defer cfgMu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]Scene, len(c.Scenes))
	for k, v := range c.Scenes {
		result[k] = v
	}
	return result
}

// AddActionToScene adds an action to a scene.
func AddActionToScene(sceneName string, action SceneAction) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	scene, exists := c.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = append(scene.Actions, action)
	c.Scenes[sceneName] = scene

	return Save()
}

// SetSceneActions replaces all actions in a scene.
func SetSceneActions(sceneName string, actions []SceneAction) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	scene, exists := c.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = actions
	c.Scenes[sceneName] = scene

	return Save()
}

// UpdateScene updates a scene's name and/or description.
func UpdateScene(oldName, newName, description string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	scene, exists := c.Scenes[oldName]
	if !exists {
		return fmt.Errorf("scene %q not found", oldName)
	}

	if newName != "" && newName != oldName {
		if err := ValidateSceneName(newName); err != nil {
			return err
		}
		if _, exists := c.Scenes[newName]; exists {
			return fmt.Errorf("scene %q already exists", newName)
		}
		delete(c.Scenes, oldName)
		scene.Name = newName
	}

	if description != "" {
		scene.Description = description
	}

	c.Scenes[scene.Name] = scene
	return Save()
}

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
