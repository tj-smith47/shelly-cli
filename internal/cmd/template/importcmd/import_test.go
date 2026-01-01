package importcmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "import <file> [name]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import <file> [name]")
	}

	wantAliases := []string{"load"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"template.yaml"}, false},
		{"two args valid", []string{"template.yaml", "my-config"}, false},
		{"three args", []string{"template.yaml", "name", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("--force flag not found")
	}
	if flag.Shorthand != "f" {
		t.Errorf("--force shorthand = %q, want %q", flag.Shorthand, "f")
	}
	if flag.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", flag.DefValue, "false")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly template import",
		".yaml",
		"--force",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestImportTemplate_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := importTemplate("/nonexistent/path/template.yaml", "", false)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("error = %v, want to contain 'failed to read file'", err)
	}
}

func TestImportTemplate_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/invalid.yaml"
	if err := os.WriteFile(file, []byte("name: [invalid yaml"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("error = %v, want to contain 'failed to parse YAML'", err)
	}
}

func TestImportTemplate_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/invalid.json"
	if err := os.WriteFile(file, []byte("{invalid json}"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("error = %v, want to contain 'failed to parse JSON'", err)
	}
}

func TestImportTemplate_MissingName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/missing-name.yaml"
	content := `model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "missing required field: name") {
		t.Errorf("error = %v, want to contain 'missing required field: name'", err)
	}
}

func TestImportTemplate_MissingModel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/missing-model.yaml"
	content := `name: "test-template"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for missing model")
	}
	if !strings.Contains(err.Error(), "missing required field: model") {
		t.Errorf("error = %v, want to contain 'missing required field: model'", err)
	}
}

func TestImportTemplate_MissingConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/missing-config.yaml"
	content := `name: "test-template"
model: "SNSW-001P16EU"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for missing config")
	}
	if !strings.Contains(err.Error(), "missing required field: config") {
		t.Errorf("error = %v, want to contain 'missing required field: config'", err)
	}
}

func TestImportTemplate_InvalidName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := tmpDir + "/invalid-name.yaml"
	content := `name: "invalid name with spaces!"
model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_NameOverride(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.yaml"
	content := `name: "original-name"
model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importTemplate(file, "overridden-name", false)
	if err != nil {
		t.Fatalf("importTemplate() error = %v", err)
	}
	if !strings.Contains(result, "overridden-name") {
		t.Errorf("result = %q, want to contain 'overridden-name'", result)
	}

	// Verify template was saved with overridden name
	tpl, exists := config.GetDeviceTemplate("overridden-name")
	if !exists {
		t.Error("template 'overridden-name' should exist")
	}
	if tpl.Name != "overridden-name" {
		t.Errorf("template.Name = %q, want %q", tpl.Name, "overridden-name")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_Success(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.yaml"
	content := `name: "my-test-template"
model: "SNSW-001P16EU"
description: "A test template"
generation: 2
config:
  switch:0:
    name: "Test Switch"
    default_state: "on"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importTemplate(file, "", false)
	if err != nil {
		t.Fatalf("importTemplate() error = %v", err)
	}
	if !strings.Contains(result, "my-test-template") {
		t.Errorf("result = %q, want to contain 'my-test-template'", result)
	}
	if !strings.Contains(result, "imported from") {
		t.Errorf("result = %q, want to contain 'imported from'", result)
	}

	// Verify template was saved
	tpl, exists := config.GetDeviceTemplate("my-test-template")
	if !exists {
		t.Error("template 'my-test-template' should exist")
	}
	if tpl.Model != "SNSW-001P16EU" {
		t.Errorf("template.Model = %q, want %q", tpl.Model, "SNSW-001P16EU")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_AlreadyExists(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.yaml"
	content := `name: "existing-template"
model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Import first time
	_, err := importTemplate(file, "", false)
	if err != nil {
		t.Fatalf("first import error = %v", err)
	}

	// Import second time without --force
	_, err = importTemplate(file, "", false)
	if err == nil {
		t.Error("expected error when importing existing template without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want to contain 'already exists'", err)
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_OverwriteWithForce(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.yaml"
	content := `name: "force-template"
model: "SNSW-001P16EU"
description: "Original"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Import first time
	_, err := importTemplate(file, "", false)
	if err != nil {
		t.Fatalf("first import error = %v", err)
	}

	// Update file with new description
	updatedContent := `name: "force-template"
model: "SNSW-001P16EU"
description: "Updated"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(updatedContent), 0o600); err != nil {
		t.Fatalf("failed to write updated file: %v", err)
	}

	// Import with --force
	result, err := importTemplate(file, "", true)
	if err != nil {
		t.Fatalf("force import error = %v", err)
	}
	if !strings.Contains(result, "force-template") {
		t.Errorf("result = %q, want to contain 'force-template'", result)
	}

	// Verify updated description
	tpl, exists := config.GetDeviceTemplate("force-template")
	if !exists {
		t.Error("template should exist")
	}
	if tpl.Description != "Updated" {
		t.Errorf("template.Description = %q, want %q", tpl.Description, "Updated")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_JSONFile(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.json"
	content := `{
  "name": "json-template",
  "model": "SNSW-001P16EU",
  "config": {
    "switch:0": {
      "name": "JSON Switch"
    }
  }
}`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importTemplate(file, "", false)
	if err != nil {
		t.Fatalf("importTemplate() error = %v", err)
	}
	if !strings.Contains(result, "json-template") {
		t.Errorf("result = %q, want to contain 'json-template'", result)
	}

	// Verify template was saved
	_, exists := config.GetDeviceTemplate("json-template")
	if !exists {
		t.Error("template 'json-template' should exist")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_InvalidNameOverride(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.yaml"
	content := `name: "valid-name"
model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Override with invalid name
	_, err := importTemplate(file, "invalid/override/name", false)
	if err == nil {
		t.Error("expected error for invalid name override")
	}
}

//nolint:paralleltest // Modifies global config state
func TestImportTemplate_UnknownFileExtension(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tmpDir := t.TempDir()
	file := tmpDir + "/template.txt"
	// Valid YAML content in .txt file
	content := `name: "txt-template"
model: "SNSW-001P16EU"
config:
  switch:0:
    name: "Test Switch"
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := importTemplate(file, "", false)
	if err != nil {
		t.Fatalf("importTemplate() error = %v", err)
	}
	if !strings.Contains(result, "txt-template") {
		t.Errorf("result = %q, want to contain 'txt-template'", result)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for tab completion")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"Import",
		"template",
		"JSON",
		"YAML",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}
