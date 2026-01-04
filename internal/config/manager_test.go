package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
)

const (
	testModelSHSW1 = "SHSW-1"
	testScriptName = "my-script"
	testScriptCode = "console.log('hello');"
)

// setupTestManager sets up an isolated Manager for testing.
// It uses an in-memory filesystem to avoid touching real files.
func setupTestManager(t *testing.T) *Manager {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	m := NewManager("/test/config/config.yaml")
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return m
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_Path(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	path := "/test/config/config.yaml"
	m := NewManager(path)

	if m.Path() != path {
		t.Errorf("Path() = %q, want %q", m.Path(), path)
	}
}

// TestManager_Reload tests real file persistence across Reload().
// Uses t.TempDir() because the test specifically validates disk I/O.
func TestManager_Reload(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device
	if err := m.RegisterDevice("test", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Reload should reset and reload from file
	if err := m.Reload(); err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	// Device should still exist (saved to file)
	if _, ok := m.GetDevice("test"); !ok {
		t.Error("device should exist after reload")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SaveWithoutLoad(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	m := NewManager("/test/config/config.yaml")

	// Try to save without loading first (don't call Load())
	err := m.Save()
	if err == nil {
		t.Error("Save() should fail when config is not loaded")
	}
}

//nolint:gocyclo,paralleltest // test function with many assertions; modifies global state via SetFs
func TestManager_SceneOperations(t *testing.T) {
	m := setupTestManager(t)

	// Create a scene
	if err := m.CreateScene("movie-night", "Dim lights for movies"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Get scene
	scene, ok := m.GetScene("movie-night")
	if !ok {
		t.Fatal("GetScene() returned false")
	}
	if scene.Description != "Dim lights for movies" {
		t.Errorf("Description = %q, want %q", scene.Description, "Dim lights for movies")
	}

	// List scenes
	scenes := m.ListScenes()
	if len(scenes) != 1 {
		t.Errorf("ListScenes() returned %d scenes, want 1", len(scenes))
	}

	// Add action to scene
	action := SceneAction{
		Device: "light1",
		Method: "Switch.Set",
		Params: map[string]any{"on": true},
	}
	if err := m.AddActionToScene("movie-night", action); err != nil {
		t.Fatalf("AddActionToScene() error: %v", err)
	}

	// Verify action was added
	scene, _ = m.GetScene("movie-night")
	if len(scene.Actions) != 1 {
		t.Errorf("scene has %d actions, want 1", len(scene.Actions))
	}

	// Set scene actions (replace all)
	newActions := []SceneAction{
		{Device: "light2", Method: "Switch.Set", Params: map[string]any{"on": false}},
	}
	if err := m.SetSceneActions("movie-night", newActions); err != nil {
		t.Fatalf("SetSceneActions() error: %v", err)
	}
	scene, _ = m.GetScene("movie-night")
	if len(scene.Actions) != 1 || scene.Actions[0].Device != "light2" {
		t.Error("SetSceneActions did not replace actions correctly")
	}

	// Update scene
	if err := m.UpdateScene("movie-night", "movie-time", "Updated description"); err != nil {
		t.Fatalf("UpdateScene() error: %v", err)
	}
	_, ok = m.GetScene("movie-night")
	if ok {
		t.Error("old scene name should not exist")
	}
	scene, ok = m.GetScene("movie-time")
	if !ok {
		t.Fatal("renamed scene should exist")
	}
	if scene.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", scene.Description, "Updated description")
	}

	// Delete scene
	if err := m.DeleteScene("movie-time"); err != nil {
		t.Fatalf("DeleteScene() error: %v", err)
	}
	_, ok = m.GetScene("movie-time")
	if ok {
		t.Error("scene should not exist after delete")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SceneOperations_Errors(t *testing.T) {
	m := setupTestManager(t)

	// Create scene to test duplicate error
	if err := m.CreateScene("test-scene", ""); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Try to create duplicate
	if err := m.CreateScene("test-scene", ""); err == nil {
		t.Error("expected error creating duplicate scene")
	}

	// Add action to nonexistent scene
	if err := m.AddActionToScene("nonexistent", SceneAction{}); err == nil {
		t.Error("expected error adding action to nonexistent scene")
	}

	// Set actions on nonexistent scene
	if err := m.SetSceneActions("nonexistent", nil); err == nil {
		t.Error("expected error setting actions on nonexistent scene")
	}

	// Update nonexistent scene
	if err := m.UpdateScene("nonexistent", "", ""); err == nil {
		t.Error("expected error updating nonexistent scene")
	}

	// Delete nonexistent scene
	if err := m.DeleteScene("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent scene")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_DeviceTemplateOperations(t *testing.T) {
	m := setupTestManager(t)

	// Create template
	cfg := map[string]any{"setting1": "value1"}
	if err := m.CreateDeviceTemplate("my-template", "Test template", testModelSHSW1, "switch", 2, cfg, "source-device"); err != nil {
		t.Fatalf("CreateDeviceTemplate() error: %v", err)
	}

	// Get template
	tpl, ok := m.GetDeviceTemplate("my-template")
	if !ok {
		t.Fatal("GetDeviceTemplate() returned false")
	}
	if tpl.Description != "Test template" {
		t.Errorf("Description = %q, want %q", tpl.Description, "Test template")
	}
	if tpl.Model != testModelSHSW1 {
		t.Errorf("Model = %q, want %q", tpl.Model, testModelSHSW1)
	}
	if tpl.App != "switch" {
		t.Errorf("App = %q, want %q", tpl.App, "switch")
	}

	// List templates
	templates := m.ListDeviceTemplates()
	if len(templates) != 1 {
		t.Errorf("ListDeviceTemplates() returned %d, want 1", len(templates))
	}

	// Update template
	if err := m.UpdateDeviceTemplate("my-template", "New description"); err != nil {
		t.Fatalf("UpdateDeviceTemplate() error: %v", err)
	}
	tpl, _ = m.GetDeviceTemplate("my-template")
	if tpl.Description != "New description" {
		t.Errorf("Description = %q, want %q", tpl.Description, "New description")
	}

	// Save template (overwrite)
	newTpl := DeviceTemplate{
		Name:        "my-template",
		Description: "Replaced",
		Model:       "SHSW-2",
		Generation:  3,
		Config:      map[string]any{},
	}
	if err := m.SaveDeviceTemplate(newTpl); err != nil {
		t.Fatalf("SaveDeviceTemplate() error: %v", err)
	}
	tpl, _ = m.GetDeviceTemplate("my-template")
	if tpl.Model != "SHSW-2" {
		t.Errorf("Model = %q, want %q", tpl.Model, "SHSW-2")
	}

	// Delete template
	if err := m.DeleteDeviceTemplate("my-template"); err != nil {
		t.Fatalf("DeleteDeviceTemplate() error: %v", err)
	}
	_, ok = m.GetDeviceTemplate("my-template")
	if ok {
		t.Error("template should not exist after delete")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_DeviceTemplateOperations_Errors(t *testing.T) {
	m := setupTestManager(t)

	// Create template
	if err := m.CreateDeviceTemplate("test", "", "Model", "", 2, map[string]any{}, ""); err != nil {
		t.Fatalf("CreateDeviceTemplate() error: %v", err)
	}

	// Try to create duplicate
	if err := m.CreateDeviceTemplate("test", "", "Model", "", 2, map[string]any{}, ""); err == nil {
		t.Error("expected error creating duplicate template")
	}

	// Update nonexistent
	if err := m.UpdateDeviceTemplate("nonexistent", ""); err == nil {
		t.Error("expected error updating nonexistent template")
	}

	// Delete nonexistent
	if err := m.DeleteDeviceTemplate("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent template")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ScriptTemplateOperations(t *testing.T) {
	m := setupTestManager(t)

	// Save script template
	tpl := ScriptTemplate{
		Name:        "my-script",
		Description: "Test script",
		Code:        testScriptCode,
		Category:    "utility",
	}
	if err := m.SaveScriptTemplate(tpl); err != nil {
		t.Fatalf("SaveScriptTemplate() error: %v", err)
	}

	// Get script template
	got, ok := m.GetScriptTemplate("my-script")
	if !ok {
		t.Fatal("GetScriptTemplate() returned false")
	}
	if got.Code != testScriptCode {
		t.Errorf("Code = %q, want %q", got.Code, testScriptCode)
	}

	// List script templates
	templates := m.ListScriptTemplates()
	if len(templates) != 1 {
		t.Errorf("ListScriptTemplates() returned %d, want 1", len(templates))
	}

	// Delete script template
	if err := m.DeleteScriptTemplate("my-script"); err != nil {
		t.Fatalf("DeleteScriptTemplate() error: %v", err)
	}
	_, ok = m.GetScriptTemplate("my-script")
	if ok {
		t.Error("template should not exist after delete")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ScriptTemplateOperations_Errors(t *testing.T) {
	m := setupTestManager(t)

	// Delete nonexistent
	if err := m.DeleteScriptTemplate("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent script template")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AlertOperations(t *testing.T) {
	m := setupTestManager(t)

	// Create alert
	if err := m.CreateAlert("power-alert", "High power warning", "device1", "power>100", "notify", true); err != nil {
		t.Fatalf("CreateAlert() error: %v", err)
	}

	// Get alert
	alert, ok := m.GetAlert("power-alert")
	if !ok {
		t.Fatal("GetAlert() returned false")
	}
	if alert.Description != "High power warning" {
		t.Errorf("Description = %q, want %q", alert.Description, "High power warning")
	}
	if alert.Device != "device1" {
		t.Errorf("Device = %q, want %q", alert.Device, "device1")
	}
	if alert.Condition != "power>100" {
		t.Errorf("Condition = %q, want %q", alert.Condition, "power>100")
	}
	if !alert.Enabled {
		t.Error("Enabled should be true")
	}

	// List alerts
	alerts := m.ListAlerts()
	if len(alerts) != 1 {
		t.Errorf("ListAlerts() returned %d, want 1", len(alerts))
	}

	// Update alert
	enabled := false
	snoozed := time.Now().Add(time.Hour).Format(time.RFC3339)
	if err := m.UpdateAlert("power-alert", &enabled, snoozed); err != nil {
		t.Fatalf("UpdateAlert() error: %v", err)
	}
	alert, _ = m.GetAlert("power-alert")
	if alert.Enabled {
		t.Error("Enabled should be false after update")
	}
	if alert.SnoozedUntil != snoozed {
		t.Errorf("SnoozedUntil = %q, want %q", alert.SnoozedUntil, snoozed)
	}

	// Delete alert
	if err := m.DeleteAlert("power-alert"); err != nil {
		t.Fatalf("DeleteAlert() error: %v", err)
	}
	_, ok = m.GetAlert("power-alert")
	if ok {
		t.Error("alert should not exist after delete")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AlertOperations_Errors(t *testing.T) {
	m := setupTestManager(t)

	// Create alert
	if err := m.CreateAlert("test", "", "", "", "", true); err != nil {
		t.Fatalf("CreateAlert() error: %v", err)
	}

	// Try to create duplicate
	if err := m.CreateAlert("test", "", "", "", "", true); err == nil {
		t.Error("expected error creating duplicate alert")
	}

	// Update nonexistent
	if err := m.UpdateAlert("nonexistent", nil, ""); err == nil {
		t.Error("expected error updating nonexistent alert")
	}

	// Delete nonexistent
	if err := m.DeleteAlert("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent alert")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ListAliasesMap(t *testing.T) {
	m := setupTestManager(t)

	if err := m.AddAlias("test1", "cmd1", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}
	if err := m.AddAlias("test2", "cmd2", true); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	aliases := m.ListAliasesMap()
	if len(aliases) != 2 {
		t.Errorf("ListAliasesMap() returned %d aliases, want 2", len(aliases))
	}

	if aliases["test1"].Command != "cmd1" {
		t.Errorf("test1 command = %q, want %q", aliases["test1"].Command, "cmd1")
	}
	if aliases["test2"].Shell != true {
		t.Error("test2 should have Shell=true")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SetDeviceAuth(t *testing.T) {
	m := setupTestManager(t)

	// Register device
	if err := m.RegisterDevice("test", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Set auth
	if err := m.SetDeviceAuth("test", "admin", "password123"); err != nil {
		t.Fatalf("SetDeviceAuth() error: %v", err)
	}

	// Verify auth was set
	dev, _ := m.GetDevice("test")
	if dev.Auth == nil {
		t.Fatal("Auth should not be nil")
	}
	if dev.Auth.Username != "admin" {
		t.Errorf("Username = %q, want %q", dev.Auth.Username, "admin")
	}
	if dev.Auth.Password != "password123" {
		t.Errorf("Password = %q, want %q", dev.Auth.Password, "password123")
	}

	// Get all credentials
	creds := m.GetAllDeviceCredentials()
	if len(creds) != 1 {
		t.Errorf("GetAllDeviceCredentials() returned %d, want 1", len(creds))
	}
	if creds["test"].Username != "admin" {
		t.Errorf("creds[test].Username = %q, want %q", creds["test"].Username, "admin")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SetDeviceAuth_NotFound(t *testing.T) {
	m := setupTestManager(t)

	if err := m.SetDeviceAuth("nonexistent", "user", "pass"); err == nil {
		t.Error("expected error setting auth on nonexistent device")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CreateScene_ValidationError(t *testing.T) {
	m := setupTestManager(t)

	// Try to create scene with invalid name
	if err := m.CreateScene("", "description"); err == nil {
		t.Error("expected error creating scene with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SaveScene_Overwrite(t *testing.T) {
	m := setupTestManager(t)

	// Create initial scene
	scene1 := Scene{
		Name:        "test-scene",
		Description: "Original",
	}
	if err := m.SaveScene(scene1); err != nil {
		t.Fatalf("SaveScene() error: %v", err)
	}

	// Overwrite with new scene
	scene2 := Scene{
		Name:        "test-scene",
		Description: "Updated",
	}
	if err := m.SaveScene(scene2); err != nil {
		t.Fatalf("SaveScene() overwrite error: %v", err)
	}

	// Verify overwrite
	got, ok := m.GetScene("test-scene")
	if !ok {
		t.Fatal("scene should exist")
	}
	if got.Description != "Updated" {
		t.Errorf("Description = %q, want %q", got.Description, "Updated")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SaveScene_ValidationError(t *testing.T) {
	m := setupTestManager(t)

	// Try to save scene with invalid name
	scene := Scene{Name: ""}
	if err := m.SaveScene(scene); err == nil {
		t.Error("expected error saving scene with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateScene_DescriptionOnly(t *testing.T) {
	m := setupTestManager(t)

	if err := m.CreateScene("test-scene", "Original"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Update description only (empty new name)
	if err := m.UpdateScene("test-scene", "", "New Description"); err != nil {
		t.Fatalf("UpdateScene() error: %v", err)
	}

	got, ok := m.GetScene("test-scene")
	if !ok {
		t.Fatal("scene should exist")
	}
	if got.Description != "New Description" {
		t.Errorf("Description = %q, want %q", got.Description, "New Description")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CreateDeviceTemplate_ValidationError(t *testing.T) {
	m := setupTestManager(t)

	// Try to create template with invalid name
	if err := m.CreateDeviceTemplate("", "desc", "model", "app", 2, nil, ""); err == nil {
		t.Error("expected error creating template with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SaveDeviceTemplate_ValidationError(t *testing.T) {
	m := setupTestManager(t)

	// Try to save template with invalid name
	tpl := DeviceTemplate{Name: ""}
	if err := m.SaveDeviceTemplate(tpl); err == nil {
		t.Error("expected error saving template with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_SaveScriptTemplate_ValidationError(t *testing.T) {
	m := setupTestManager(t)

	// Try to save script template with invalid name
	tpl := ScriptTemplate{Name: ""}
	if err := m.SaveScriptTemplate(tpl); err == nil {
		t.Error("expected error saving script template with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_Load_InvalidYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	SetFs(fs)
	t.Cleanup(func() { SetFs(nil) })
	configPath := "/test/config/config.yaml"

	// Write invalid YAML
	if err := afero.WriteFile(fs, configPath, []byte(":\ninvalid yaml content"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	m := NewManager(configPath)
	if err := m.Load(); err == nil {
		t.Error("expected error loading invalid YAML")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddAlias_NilMap(t *testing.T) {
	m := setupTestManager(t)

	// Force nil aliases map
	m.mu.Lock()
	m.config.Aliases = nil
	m.mu.Unlock()

	// AddAlias should initialize the map
	if err := m.AddAlias("test", "device info $1", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	alias, ok := m.GetAlias("test")
	if !ok {
		t.Error("alias should exist")
	}
	if alias.Command != "device info $1" {
		t.Errorf("Command = %q, want %q", alias.Command, "device info $1")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateScene_RenameValidationError(t *testing.T) {
	m := setupTestManager(t)

	if err := m.CreateScene("original", "description"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Try to rename with invalid name
	err := m.UpdateScene("original", "", "new description")
	if err != nil {
		t.Errorf("UpdateScene with empty newName should succeed (description only): %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateScene_RenameAlreadyExists(t *testing.T) {
	m := setupTestManager(t)

	if err := m.CreateScene("scene1", "description1"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}
	if err := m.CreateScene("scene2", "description2"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Try to rename scene1 to scene2 (already exists)
	err := m.UpdateScene("scene1", "scene2", "")
	if err == nil {
		t.Error("UpdateScene should error when new name already exists")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateScene_RenameWithValidation(t *testing.T) {
	m := setupTestManager(t)

	if err := m.CreateScene("original", "description"); err != nil {
		t.Fatalf("CreateScene() error: %v", err)
	}

	// Rename to a valid new name
	if err := m.UpdateScene("original", "new-name", "new description"); err != nil {
		t.Fatalf("UpdateScene() error: %v", err)
	}

	// Verify rename
	_, ok := m.GetScene("original")
	if ok {
		t.Error("old scene name should not exist")
	}
	got, ok := m.GetScene("new-name")
	if !ok {
		t.Error("new scene name should exist")
	}
	if got.Description != "new description" {
		t.Errorf("Description = %q, want %q", got.Description, "new description")
	}
}
