package config

import (
	"testing"

	"github.com/spf13/afero"
)

const testTemplateName = "my-template"

// testModelSHSW1 and testScriptCode are defined in manager_test.go

func TestValidateTemplateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "my-template", false},
		{"valid with underscore", "my_template", false},
		{"valid alphanumeric", "template123", false},
		{"empty", "", true},
		{"too long", string(make([]byte, 65)), true},
		{"starts with hyphen", "-template", true},
		{"contains space", "my template", true},
		{"contains special char", "template@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateTemplateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseDeviceTemplateFile_JSON(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"name": "my-template",
		"description": "Test template",
		"model": "SHSW-1",
		"generation": 2,
		"config": {"key": "value"}
	}`)

	tpl, err := ParseDeviceTemplateFile("template.json", content)
	if err != nil {
		t.Fatalf("ParseDeviceTemplateFile() error: %v", err)
	}

	if tpl.Name != testTemplateName {
		t.Errorf("Name = %q, want %q", tpl.Name, testTemplateName)
	}
	if tpl.Model != testModelSHSW1 {
		t.Errorf("Model = %q, want %q", tpl.Model, testModelSHSW1)
	}
	if tpl.Generation != 2 {
		t.Errorf("Generation = %d, want %d", tpl.Generation, 2)
	}
}

func TestParseDeviceTemplateFile_YAML(t *testing.T) {
	t.Parallel()

	content := []byte(`name: my-template
description: Test template
model: SHSW-1
generation: 2
config:
  key: value
`)

	tpl, err := ParseDeviceTemplateFile("template.yaml", content)
	if err != nil {
		t.Fatalf("ParseDeviceTemplateFile() error: %v", err)
	}

	if tpl.Name != testTemplateName {
		t.Errorf("Name = %q, want %q", tpl.Name, testTemplateName)
	}
	if tpl.Model != testModelSHSW1 {
		t.Errorf("Model = %q, want %q", tpl.Model, testModelSHSW1)
	}
}

func TestParseDeviceTemplateFile_UnknownExtension(t *testing.T) {
	t.Parallel()

	// Should try YAML first, then JSON
	content := []byte(`name: my-template
model: SHSW-1
config:
  key: value
`)

	tpl, err := ParseDeviceTemplateFile("template.txt", content)
	if err != nil {
		t.Fatalf("ParseDeviceTemplateFile() error: %v", err)
	}

	if tpl.Name != testTemplateName {
		t.Errorf("Name = %q, want %q", tpl.Name, testTemplateName)
	}
}

func TestParseDeviceTemplateFile_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing name",
			content: `{"model": "SHSW-1", "config": {}}`,
			wantErr: "template missing required field: name",
		},
		{
			name:    "missing model",
			content: `{"name": "test", "config": {}}`,
			wantErr: "template missing required field: model",
		},
		{
			name:    "missing config",
			content: `{"name": "test", "model": "SHSW-1"}`,
			wantErr: "template missing required field: config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseDeviceTemplateFile("test.json", []byte(tt.content))
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseDeviceTemplateFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseDeviceTemplateFile("test.json", []byte("not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseDeviceTemplateFile_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := ParseDeviceTemplateFile("test.yaml", []byte(":\ninvalid yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseScriptTemplateFile_JSON(t *testing.T) {
	t.Parallel()

	content := `{
		"name": "my-script",
		"description": "Test script",
		"code": "console.log('hello');",
		"category": "utility"
	}`

	tpl, err := ParseScriptTemplateFile("test.json", []byte(content))
	if err != nil {
		t.Fatalf("ParseScriptTemplateFile() error: %v", err)
	}

	if tpl.Name != testScriptName {
		t.Errorf("Name = %q, want %q", tpl.Name, testScriptName)
	}
	if tpl.Code != testScriptCode {
		t.Errorf("Code = %q, want %q", tpl.Code, testScriptCode)
	}
}

func TestParseScriptTemplateFile_YAML(t *testing.T) {
	t.Parallel()

	content := `name: my-script
description: Test script
code: |
  console.log('hello');
category: utility
`

	tpl, err := ParseScriptTemplateFile("test.yaml", []byte(content))
	if err != nil {
		t.Fatalf("ParseScriptTemplateFile() error: %v", err)
	}

	if tpl.Name != testScriptName {
		t.Errorf("Name = %q, want %q", tpl.Name, testScriptName)
	}
}

func TestParseScriptTemplateFile_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing name",
			content: `{"code": "test"}`,
			wantErr: "script template missing required field: name",
		},
		{
			name:    "missing code",
			content: `{"name": "test"}`,
			wantErr: "script template missing required field: code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseScriptTemplateFile("test.json", []byte(tt.content))
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIsCompatibleModel(t *testing.T) {
	t.Parallel()

	tpl := DeviceTemplate{
		Name:  "test",
		Model: testModelSHSW1,
	}

	if !IsCompatibleModel(tpl, testModelSHSW1) {
		t.Error("IsCompatibleModel() should return true for matching model")
	}
	if IsCompatibleModel(tpl, "SHSW-2") {
		t.Error("IsCompatibleModel() should return false for non-matching model")
	}
}

func TestIsCompatibleGeneration(t *testing.T) {
	t.Parallel()

	tpl := DeviceTemplate{
		Name:       "test",
		Generation: 2,
	}

	if !IsCompatibleGeneration(tpl, 2) {
		t.Error("IsCompatibleGeneration() should return true for matching generation")
	}
	if IsCompatibleGeneration(tpl, 1) {
		t.Error("IsCompatibleGeneration() should return false for non-matching generation")
	}
}

func TestParseDeviceTemplateFile_UnknownExtBothFail(t *testing.T) {
	t.Parallel()

	// Invalid data that fails both YAML and JSON parsing
	_, err := ParseDeviceTemplateFile("template.txt", []byte("{invalid{"))
	if err == nil {
		t.Error("expected error for invalid content with unknown extension")
	}
}

func TestParseScriptTemplateFile_UnknownExt(t *testing.T) {
	t.Parallel()

	// Should try YAML first, then JSON
	content := `name: my-script
code: console.log('hello');
`
	tpl, err := ParseScriptTemplateFile("script.txt", []byte(content))
	if err != nil {
		t.Fatalf("ParseScriptTemplateFile() error: %v", err)
	}
	if tpl.Name != testScriptName {
		t.Errorf("Name = %q, want %q", tpl.Name, testScriptName)
	}
}

func TestParseScriptTemplateFile_UnknownExtBothFail(t *testing.T) {
	t.Parallel()

	// Invalid data that fails both YAML and JSON parsing
	_, err := ParseScriptTemplateFile("script.txt", []byte("{invalid{"))
	if err == nil {
		t.Error("expected error for invalid content with unknown extension")
	}
}

func TestParseScriptTemplateFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseScriptTemplateFile("test.json", []byte("not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseScriptTemplateFile_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := ParseScriptTemplateFile("test.yaml", []byte(":\ninvalid yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestPackageLevelDeviceTemplateFunctions(t *testing.T) {
	// Use in-memory filesystem for test isolation
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()

	// Test CreateDeviceTemplate
	err := CreateDeviceTemplate("test-tpl", "Test template", testModelSHSW1, "", 2, map[string]any{"key": "value"}, "")
	if err != nil {
		t.Fatalf("CreateDeviceTemplate() error: %v", err)
	}

	// Test GetDeviceTemplate
	tpl, ok := GetDeviceTemplate("test-tpl")
	if !ok {
		t.Fatal("GetDeviceTemplate() returned false for existing template")
	}
	if tpl.Name != "test-tpl" {
		t.Errorf("GetDeviceTemplate() Name = %q, want %q", tpl.Name, "test-tpl")
	}
	if tpl.Model != testModelSHSW1 {
		t.Errorf("GetDeviceTemplate() Model = %q, want %q", tpl.Model, testModelSHSW1)
	}

	// Test UpdateDeviceTemplate
	if err := UpdateDeviceTemplate("test-tpl", "Updated description"); err != nil {
		t.Fatalf("UpdateDeviceTemplate() error: %v", err)
	}
	tpl, _ = GetDeviceTemplate("test-tpl")
	if tpl.Description != "Updated description" {
		t.Errorf("UpdateDeviceTemplate() Description = %q, want %q", tpl.Description, "Updated description")
	}

	// Test ListDeviceTemplates
	templates := ListDeviceTemplates()
	if len(templates) != 1 {
		t.Errorf("ListDeviceTemplates() returned %d templates, want 1", len(templates))
	}
	if _, ok := templates["test-tpl"]; !ok {
		t.Error("ListDeviceTemplates() missing test-tpl")
	}

	// Test SaveDeviceTemplate
	newTpl := DeviceTemplate{
		Name:        "saved-tpl",
		Description: "Saved template",
		Model:       "SHSW-2",
		Generation:  2,
		Config:      map[string]any{"foo": "bar"},
	}
	if err := SaveDeviceTemplate(newTpl); err != nil {
		t.Fatalf("SaveDeviceTemplate() error: %v", err)
	}
	saved, ok := GetDeviceTemplate("saved-tpl")
	if !ok {
		t.Fatal("SaveDeviceTemplate() template not found after save")
	}
	if saved.Model != "SHSW-2" {
		t.Errorf("SaveDeviceTemplate() Model = %q, want %q", saved.Model, "SHSW-2")
	}

	// Test DeleteDeviceTemplate
	if err := DeleteDeviceTemplate("test-tpl"); err != nil {
		t.Fatalf("DeleteDeviceTemplate() error: %v", err)
	}
	if _, ok := GetDeviceTemplate("test-tpl"); ok {
		t.Error("template still exists after DeleteDeviceTemplate()")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestPackageLevelScriptTemplateFunctions(t *testing.T) {
	// Use in-memory filesystem for test isolation
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()

	// Test SaveScriptTemplate
	tpl := ScriptTemplate{
		Name:        "test-script",
		Description: "Test script",
		Code:        "console.log('hello');",
		Category:    "utility",
	}
	if err := SaveScriptTemplate(tpl); err != nil {
		t.Fatalf("SaveScriptTemplate() error: %v", err)
	}

	// Test GetScriptTemplate
	saved, ok := GetScriptTemplate("test-script")
	if !ok {
		t.Fatal("GetScriptTemplate() returned false for existing template")
	}
	if saved.Name != "test-script" {
		t.Errorf("GetScriptTemplate() Name = %q, want %q", saved.Name, "test-script")
	}
	if saved.Code != "console.log('hello');" {
		t.Errorf("GetScriptTemplate() Code = %q, want %q", saved.Code, "console.log('hello');")
	}

	// Test ListScriptTemplates
	templates := ListScriptTemplates()
	if len(templates) != 1 {
		t.Errorf("ListScriptTemplates() returned %d templates, want 1", len(templates))
	}
	if _, ok := templates["test-script"]; !ok {
		t.Error("ListScriptTemplates() missing test-script")
	}

	// Test DeleteScriptTemplate
	if err := DeleteScriptTemplate("test-script"); err != nil {
		t.Fatalf("DeleteScriptTemplate() error: %v", err)
	}
	if _, ok := GetScriptTemplate("test-script"); ok {
		t.Error("template still exists after DeleteScriptTemplate()")
	}
}

//nolint:paralleltest // Tests modify global state
func TestExportDeviceTemplateToFile(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()

	// Create a template to export
	err := CreateDeviceTemplate("export-test", "Test export", "SHSW-1", "switch", 2, map[string]any{"key": "value"}, "source-dev")
	if err != nil {
		t.Fatalf("CreateDeviceTemplate() error: %v", err)
	}

	// Create export directory
	if err := Fs().MkdirAll("/export/templates", 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}

	// Export to file
	outputPath := "/export/templates/export-test.json"
	filePath, err := ExportDeviceTemplateToFile("export-test", outputPath)
	if err != nil {
		t.Fatalf("ExportDeviceTemplateToFile() error: %v", err)
	}
	if filePath != outputPath {
		t.Errorf("filePath = %q, want %q", filePath, outputPath)
	}

	// Verify file was created and is valid JSON
	data, err := afero.ReadFile(Fs(), outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	// Parse exported template
	exported, err := ParseDeviceTemplateFile(outputPath, data)
	if err != nil {
		t.Fatalf("ParseDeviceTemplateFile() error: %v", err)
	}
	if exported.Name != "export-test" {
		t.Errorf("exported.Name = %q, want %q", exported.Name, "export-test")
	}
	if exported.Model != "SHSW-1" {
		t.Errorf("exported.Model = %q, want %q", exported.Model, "SHSW-1")
	}
}

//nolint:paralleltest // Tests modify global state
func TestExportDeviceTemplateToFile_NotFound(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()

	_, err := ExportDeviceTemplateToFile("nonexistent", "/export/nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}
