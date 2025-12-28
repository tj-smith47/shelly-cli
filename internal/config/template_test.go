package config

import "testing"

const testTemplateName = "my-template"

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
	if tpl.Model != "SHSW-1" {
		t.Errorf("Model = %q, want %q", tpl.Model, "SHSW-1")
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
	if tpl.Model != "SHSW-1" {
		t.Errorf("Model = %q, want %q", tpl.Model, "SHSW-1")
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

	if tpl.Name != "my-script" {
		t.Errorf("Name = %q, want %q", tpl.Name, "my-script")
	}
	if tpl.Code != "console.log('hello');" {
		t.Errorf("Code = %q, want %q", tpl.Code, "console.log('hello');")
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

	if tpl.Name != "my-script" {
		t.Errorf("Name = %q, want %q", tpl.Name, "my-script")
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
		Model: "SHSW-1",
	}

	if !IsCompatibleModel(tpl, "SHSW-1") {
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
