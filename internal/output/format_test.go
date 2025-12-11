package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestGetFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want Format
	}{
		// Default case is tested separately as it requires viper setup
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Test would require viper setup
		})
	}
}

func TestParseFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"YAML", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"table", FormatTable, false},
		{"TABLE", FormatTable, false},
		{"text", FormatText, false},
		{"plain", FormatText, false},
		{"template", FormatTemplate, false},
		{"TEMPLATE", FormatTemplate, false},
		{"go-template", FormatTemplate, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidFormats(t *testing.T) {
	t.Parallel()

	formats := ValidFormats()
	if len(formats) != 5 {
		t.Errorf("expected 5 formats, got %d", len(formats))
	}

	expected := []string{"json", "yaml", "table", "text", "template"}
	for i, f := range expected {
		if formats[i] != f {
			t.Errorf("expected format[%d] = %q, got %q", i, f, formats[i])
		}
	}
}

func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := NewJSONFormatter()
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"key"`) {
		t.Error("expected JSON output to contain key")
	}
	if !strings.Contains(output, `"value"`) {
		t.Error("expected JSON output to contain value")
	}
}

func TestYAMLFormatter(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := NewYAMLFormatter()
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "key:") {
		t.Error("expected YAML output to contain key:")
	}
	if !strings.Contains(output, "value") {
		t.Error("expected YAML output to contain value")
	}
}

func TestTextFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     any
		contains string
	}{
		{"string", "hello world", "hello world"},
		{"string slice", []string{"a", "b", "c"}, "a\nb\nc"},
		{"struct", struct{ Name string }{"test"}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			f := NewTextFormatter()
			err := f.Format(&buf, tt.data)
			if err != nil {
				t.Fatalf("Format() error: %v", err)
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, buf.String())
			}
		})
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	data := map[string]int{"count": 42}
	var buf bytes.Buffer

	err := JSON(&buf, data)
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	if !strings.Contains(buf.String(), "42") {
		t.Error("expected JSON output to contain 42")
	}
}

func TestYAML(t *testing.T) {
	t.Parallel()

	data := map[string]int{"count": 42}
	var buf bytes.Buffer

	err := YAML(&buf, data)
	if err != nil {
		t.Fatalf("YAML() error: %v", err)
	}

	if !strings.Contains(buf.String(), "42") {
		t.Error("expected YAML output to contain 42")
	}
}

func TestTemplateFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		data     any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple field access",
			template: "Name: {{.Name}}",
			data:     struct{ Name string }{"TestDevice"},
			want:     "Name: TestDevice\n",
			wantErr:  false,
		},
		{
			name:     "map access",
			template: "Value: {{.key}}",
			data:     map[string]string{"key": "value"},
			want:     "Value: value\n",
			wantErr:  false,
		},
		{
			name:     "multiple fields",
			template: "{{.Name}} - {{.Status}}",
			data:     struct{ Name, Status string }{"Device1", "online"},
			want:     "Device1 - online\n",
			wantErr:  false,
		},
		{
			name:     "range over slice",
			template: "{{range .}}{{.}}\n{{end}}",
			data:     []string{"a", "b", "c"},
			want:     "a\nb\nc\n\n",
			wantErr:  false,
		},
		{
			name:     "conditional",
			template: "{{if .Active}}ON{{else}}OFF{{end}}",
			data:     struct{ Active bool }{true},
			want:     "ON\n",
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			data:     "anything",
			wantErr:  true,
		},
		{
			name:     "invalid template syntax",
			template: "{{.Invalid",
			data:     "anything",
			wantErr:  true,
		},
		{
			name:     "template with trailing newline",
			template: "Hello\n",
			data:     nil,
			want:     "Hello\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			f := NewTemplateFormatter(tt.template)
			err := f.Format(&buf, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && buf.String() != tt.want {
				t.Errorf("Format() output = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	data := struct {
		ID   int
		Name string
	}{42, "Test"}

	var buf bytes.Buffer
	err := Template(&buf, "ID={{.ID}}, Name={{.Name}}", data)
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	expected := "ID=42, Name=Test\n"
	if buf.String() != expected {
		t.Errorf("Template() output = %q, want %q", buf.String(), expected)
	}
}

func TestTemplateFormatter_ComplexData(t *testing.T) {
	t.Parallel()

	type Device struct {
		ID     int
		Name   string
		Online bool
		Power  float64
	}

	devices := []Device{
		{1, "Living Room", true, 45.5},
		{2, "Bedroom", false, 0.0},
	}

	tmpl := `{{range .}}{{.Name}}: {{if .Online}}ON ({{.Power}}W){{else}}OFF{{end}}
{{end}}`

	var buf bytes.Buffer
	f := NewTemplateFormatter(tmpl)
	err := f.Format(&buf, devices)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Living Room: ON (45.5W)") {
		t.Error("expected output to contain 'Living Room: ON (45.5W)'")
	}
	if !strings.Contains(output, "Bedroom: OFF") {
		t.Error("expected output to contain 'Bedroom: OFF'")
	}
}
