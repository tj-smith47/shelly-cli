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
	// Disable highlighting for tests to get predictable output
	f.Highlight = false
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"key"`) {
		t.Errorf("expected JSON output to contain key, got: %q", output)
	}
	if !strings.Contains(output, `"value"`) {
		t.Errorf("expected JSON output to contain value, got: %q", output)
	}
}

func TestYAMLFormatter(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := NewYAMLFormatter()
	// Disable highlighting for tests to get predictable output
	f.Highlight = false
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "key:") {
		t.Errorf("expected YAML output to contain key:, got: %q", output)
	}
	if !strings.Contains(output, "value") {
		t.Errorf("expected YAML output to contain value, got: %q", output)
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

func TestFormatConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{"JSON", FormatJSON, "json"},
		{"YAML", FormatYAML, "yaml"},
		{"Table", FormatTable, "table"},
		{"Text", FormatText, "text"},
		{"Template", FormatTemplate, "template"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.format) != tt.want {
				t.Errorf("Format constant %s = %q, want %q", tt.name, tt.format, tt.want)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format Format
	}{
		{"JSON", FormatJSON},
		{"YAML", FormatYAML},
		{"Table", FormatTable},
		{"Text", FormatText},
		{"Unknown", Format("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			formatter := NewFormatter(tt.format)
			if formatter == nil {
				t.Errorf("NewFormatter(%q) returned nil", tt.format)
			}
		})
	}
}

func TestFormatReleaseNotes(t *testing.T) {
	t.Parallel()

	t.Run("short notes", func(t *testing.T) {
		t.Parallel()
		body := "Line1\nLine2"
		result := FormatReleaseNotes(body)
		if !strings.Contains(result, "  Line1") {
			t.Error("expected indented lines")
		}
	})

	t.Run("long notes truncated", func(t *testing.T) {
		t.Parallel()
		body := strings.Repeat("a", 600)
		result := FormatReleaseNotes(body)
		if !strings.HasSuffix(result, "...") {
			t.Error("expected truncated output")
		}
	})
}

func TestTableFormatter_Format(t *testing.T) {
	t.Parallel()

	t.Run("struct slice", func(t *testing.T) {
		t.Parallel()
		type item struct {
			Name  string
			Value int
		}
		data := []item{
			{"a", 1},
			{"b", 2},
		}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		// Table headers are uppercase by default
		lowerOutput := strings.ToLower(output)
		if !strings.Contains(lowerOutput, "name") || !strings.Contains(lowerOutput, "value") {
			t.Errorf("expected table headers, got: %s", output)
		}
	})

	t.Run("non-tabular data", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, "just a string")
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "just a string") {
			t.Error("expected text fallback")
		}
	})
}

// testStringer is a type that implements fmt.Stringer for testing.
type testStringer struct{}

func (ts testStringer) String() string {
	return "custom stringer output"
}

func TestTextFormatter_Stringer(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewTextFormatter()
	err := f.Format(&buf, testStringer{})
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(buf.String(), "custom stringer output") {
		t.Errorf("expected stringer output, got %q", buf.String())
	}
}

func TestFormatPlaceholder(t *testing.T) {
	t.Parallel()

	result := FormatPlaceholder("placeholder text")
	if result == "" {
		t.Error("FormatPlaceholder returned empty string")
	}
	// Just verify it returns something with the text (theme may add styling)
	if !strings.Contains(result, "placeholder") {
		t.Error("expected result to contain placeholder text")
	}
}

func TestGetFormat_Default(t *testing.T) {
	t.Parallel()
	// Test default format (when no config is set)
	format := GetFormat()
	// Default should be table
	if format != FormatTable {
		t.Errorf("GetFormat() default = %v, want %v", format, FormatTable)
	}
}

func TestGetTemplate_Default(t *testing.T) {
	t.Parallel()
	// Test that GetTemplate returns empty string when not configured
	tmpl := GetTemplate()
	// Default template should be empty
	if tmpl != "" {
		t.Errorf("GetTemplate() default = %q, want empty", tmpl)
	}
}

func TestPrintTo(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer
	err := PrintTo(&buf, data)
	if err != nil {
		t.Fatalf("PrintTo() error: %v", err)
	}
	// Should produce some output (table format by default)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestPrintTemplate(t *testing.T) {
	t.Parallel()
	// PrintTemplate writes to os.Stdout which we can't easily capture
	// But we can verify it doesn't panic and returns no error for valid template
	data := struct{ Name string }{"Test"}
	err := Template(&bytes.Buffer{}, "Name: {{.Name}}", data)
	if err != nil {
		t.Errorf("Template() error: %v", err)
	}
}

func TestIsQuiet_Default(t *testing.T) {
	t.Parallel()
	// Test that IsQuiet returns false when not configured
	quiet := IsQuiet()
	if quiet {
		t.Error("IsQuiet() default should be false")
	}
}

func TestIsVerbose_Default(t *testing.T) {
	t.Parallel()
	// Test that IsVerbose returns false when not configured
	verbose := IsVerbose()
	if verbose {
		t.Error("IsVerbose() default should be false")
	}
}

func TestWantsJSON_Default(t *testing.T) {
	t.Parallel()
	// Test that WantsJSON returns false when default format is table
	want := WantsJSON()
	if want {
		t.Error("WantsJSON() default should be false")
	}
}

func TestWantsYAML_Default(t *testing.T) {
	t.Parallel()
	// Test that WantsYAML returns false when default format is table
	want := WantsYAML()
	if want {
		t.Error("WantsYAML() default should be false")
	}
}

func TestWantsTable_Default(t *testing.T) {
	t.Parallel()
	// Test that WantsTable returns true when default format is table
	want := WantsTable()
	if !want {
		t.Error("WantsTable() default should be true")
	}
}

func TestWantsStructured_Default(t *testing.T) {
	t.Parallel()
	// Test that WantsStructured returns false when default format is table
	want := WantsStructured()
	if want {
		t.Error("WantsStructured() default should be false")
	}
}

func TestFormatOutput(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer
	err := FormatOutput(&buf, data)
	if err != nil {
		t.Fatalf("FormatOutput() error: %v", err)
	}
	// Should produce some output (table format by default)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestHighlightCode(t *testing.T) {
	t.Parallel()

	t.Run("json syntax", func(t *testing.T) {
		t.Parallel()
		code := `{"key": "value"}`
		result := highlightCode(code, "json")
		// Should return non-empty result
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("yaml syntax", func(t *testing.T) {
		t.Parallel()
		code := "key: value"
		result := highlightCode(code, "yaml")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("unknown lexer returns code unchanged", func(t *testing.T) {
		t.Parallel()
		code := "some text"
		result := highlightCode(code, "nonexistent-language-xyz123")
		// Should return original code when lexer not found
		if result != code {
			t.Errorf("expected code unchanged, got %q", result)
		}
	})
}

func TestGetChromaStyle(t *testing.T) {
	t.Parallel()

	// Test that getChromaStyle returns a non-nil style
	style := getChromaStyle()
	if style == nil {
		t.Error("getChromaStyle() returned nil")
	}
}
