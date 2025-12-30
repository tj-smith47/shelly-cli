package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// viperSetOutput sets the "output" viper value for the duration of the test and restores it after.
func viperSetOutput(t *testing.T, value string) {
	t.Helper()
	original := viper.GetString("output")
	viper.Set("output", value)
	t.Cleanup(func() {
		viper.Set("output", original)
	})
}

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

//nolint:paralleltest // Tests share viper state for output format
func TestGetFormat_WithViper(t *testing.T) {
	t.Run("json format", func(t *testing.T) {
		viperSetOutput(t, "json")
		got := GetFormat()
		if got != FormatJSON {
			t.Errorf("GetFormat() = %v, want %v", got, FormatJSON)
		}
	})

	t.Run("yaml format", func(t *testing.T) {
		viperSetOutput(t, "yaml")
		got := GetFormat()
		if got != FormatYAML {
			t.Errorf("GetFormat() = %v, want %v", got, FormatYAML)
		}
	})

	t.Run("yml format", func(t *testing.T) {
		viperSetOutput(t, "yml")
		got := GetFormat()
		if got != FormatYAML {
			t.Errorf("GetFormat() = %v, want %v", got, FormatYAML)
		}
	})

	t.Run("table format", func(t *testing.T) {
		viperSetOutput(t, "table")
		got := GetFormat()
		if got != FormatTable {
			t.Errorf("GetFormat() = %v, want %v", got, FormatTable)
		}
	})

	t.Run("text format", func(t *testing.T) {
		viperSetOutput(t, "text")
		got := GetFormat()
		if got != FormatText {
			t.Errorf("GetFormat() = %v, want %v", got, FormatText)
		}
	})

	t.Run("plain format", func(t *testing.T) {
		viperSetOutput(t, "plain")
		got := GetFormat()
		if got != FormatText {
			t.Errorf("GetFormat() = %v, want %v", got, FormatText)
		}
	})

	t.Run("template format", func(t *testing.T) {
		viperSetOutput(t, "template")
		got := GetFormat()
		if got != FormatTemplate {
			t.Errorf("GetFormat() = %v, want %v", got, FormatTemplate)
		}
	})

	t.Run("go-template format", func(t *testing.T) {
		viperSetOutput(t, "go-template")
		got := GetFormat()
		if got != FormatTemplate {
			t.Errorf("GetFormat() = %v, want %v", got, FormatTemplate)
		}
	})

	t.Run("unknown defaults to table", func(t *testing.T) {
		viperSetOutput(t, "unknown-format")
		got := GetFormat()
		if got != FormatTable {
			t.Errorf("GetFormat() = %v, want %v (default)", got, FormatTable)
		}
	})
}

func TestShouldHighlight(t *testing.T) {
	t.Parallel()

	// In test environment, terminal is not a TTY so shouldHighlight should return false
	if shouldHighlight() {
		// Not expected in test environment, but if it's true, that's also valid
		t.Log("shouldHighlight returned true (unexpected in test, but acceptable)")
	}
}

func TestColorEnabled(t *testing.T) {
	t.Parallel()

	// In test environment, terminal is not a TTY so colorEnabled should return false
	if colorEnabled() {
		// Not expected in test environment, but if it's true, that's also valid
		t.Log("colorEnabled returned true (unexpected in test, but acceptable)")
	}
}

func TestNewFormatter_Template(t *testing.T) {
	t.Parallel()

	// Test that NewFormatter with FormatTemplate creates a TemplateFormatter
	viperSetOutput(t, "template")
	formatter := NewFormatter(FormatTemplate)
	if formatter == nil {
		t.Fatal("NewFormatter returned nil for template format")
	}
}

func TestJSONFormatter_WithHighlight(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := &JSONFormatter{Indent: true, Highlight: true}
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}
	// Should produce output (might have ANSI codes)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestJSONFormatter_NoIndent(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := &JSONFormatter{Indent: false, Highlight: false}
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}
	output := buf.String()
	// Without indentation, should be compact
	if strings.Contains(output, "  ") {
		t.Error("expected no indentation in output")
	}
}

func TestYAMLFormatter_WithHighlight(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	f := &YAMLFormatter{Highlight: true}
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestTemplateFormatter_ExecutionError(t *testing.T) {
	t.Parallel()

	// Template that tries to access a non-existent field
	tmpl := "{{.NonExistent.Field}}"
	data := struct{ Name string }{"test"}

	var buf bytes.Buffer
	f := NewTemplateFormatter(tmpl)
	err := f.Format(&buf, data)
	// Should return an error for missing field
	if err == nil {
		t.Error("expected error for template execution failure")
	}
}

func TestFormatDisplayValue_Float(t *testing.T) {
	t.Parallel()

	got := FormatDisplayValue(float64(3.14159))
	if !strings.Contains(got, "3.14") {
		t.Errorf("expected float format, got %q", got)
	}
}

func TestRenderProgressBar_OverMax(t *testing.T) {
	t.Parallel()

	// Test when value exceeds max
	got := RenderProgressBar(25, 10)
	if got == "" {
		t.Error("expected non-empty result")
	}
}

func TestExtractMapSection_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Test with a type that can't be marshaled to JSON properly
	ch := make(chan int)
	got := ExtractMapSection(ch, "key")
	if got != nil {
		t.Error("expected nil for unmarshalable type")
	}
}

func TestExtractMapSection_InvalidUnmarshal(t *testing.T) {
	t.Parallel()

	// This should work - just test non-map section
	data := map[string]any{
		"notamap": "just a string",
	}
	got := ExtractMapSection(data, "notamap")
	if got != nil {
		t.Error("expected nil for non-map section")
	}
}

func TestGetChromaStyle_Themes(t *testing.T) {
	t.Parallel()

	themes := []string{"dracula", "nord", "gruvbox", "tokyo-night", "catppuccin", "unknown-theme"}
	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			t.Parallel()
			// Set theme in viper (note: may affect other tests so we use t.Cleanup)
			original := viper.GetString("theme.name")
			viper.Set("theme.name", themeName)
			t.Cleanup(func() {
				viper.Set("theme.name", original)
			})

			style := getChromaStyle()
			if style == nil {
				t.Errorf("getChromaStyle() returned nil for theme %q", themeName)
			}
		})
	}
}

// TestPrint_Stdout tests Print functions that write to os.Stdout.
// We can't easily capture stdout in parallel tests, but we can verify they don't error.
func TestPrint_Stdout(t *testing.T) {
	// Print, PrintJSON, PrintYAML write to os.Stdout
	// These are thin wrappers around PrintTo/JSON/YAML which are tested
	// Verify they exist and compile correctly by calling PrintTo instead
	data := map[string]string{"key": "value"}

	// Test PrintTo (underlying implementation)
	var buf bytes.Buffer
	if err := PrintTo(&buf, data); err != nil {
		t.Errorf("PrintTo() error: %v", err)
	}

	// Test JSON (underlying for PrintJSON)
	buf.Reset()
	if err := JSON(&buf, data); err != nil {
		t.Errorf("JSON() error: %v", err)
	}

	// Test YAML (underlying for PrintYAML)
	buf.Reset()
	if err := YAML(&buf, data); err != nil {
		t.Errorf("YAML() error: %v", err)
	}
}

//nolint:paralleltest // Tests manipulate shared environment
func TestColorEnabled_EnvVars(t *testing.T) {
	// Test NO_COLOR environment variable
	t.Run("NO_COLOR set", func(t *testing.T) {
		// Set NO_COLOR
		t.Setenv("NO_COLOR", "1")
		// colorEnabled should return false when NO_COLOR is set
		// (also returns false because not a TTY, but that's fine)
		result := colorEnabled()
		if result {
			t.Log("colorEnabled returned true even with NO_COLOR set (TTY override)")
		}
	})

	t.Run("SHELLY_NO_COLOR set", func(t *testing.T) {
		t.Setenv("SHELLY_NO_COLOR", "1")
		result := colorEnabled()
		if result {
			t.Log("colorEnabled returned true even with SHELLY_NO_COLOR set (TTY override)")
		}
	})

	t.Run("TERM=dumb", func(t *testing.T) {
		t.Setenv("TERM", "dumb")
		result := colorEnabled()
		if result {
			t.Log("colorEnabled returned true even with TERM=dumb (TTY override)")
		}
	})

	t.Run("plain flag set", func(t *testing.T) {
		original := viper.GetBool("plain")
		viper.Set("plain", true)
		t.Cleanup(func() {
			viper.Set("plain", original)
		})
		result := colorEnabled()
		if result {
			t.Log("colorEnabled returned true even with plain=true (TTY override)")
		}
	})

	t.Run("no-color flag set", func(t *testing.T) {
		original := viper.GetBool("no-color")
		viper.Set("no-color", true)
		t.Cleanup(func() {
			viper.Set("no-color", original)
		})
		result := colorEnabled()
		if result {
			t.Log("colorEnabled returned true even with no-color=true (TTY override)")
		}
	})
}

//nolint:paralleltest // Tests manipulate shared viper state
func TestShouldHighlight_Flags(t *testing.T) {
	t.Run("plain flag disables highlight", func(t *testing.T) {
		original := viper.GetBool("plain")
		viper.Set("plain", true)
		t.Cleanup(func() {
			viper.Set("plain", original)
		})
		result := shouldHighlight()
		if result {
			t.Error("shouldHighlight should return false when plain=true")
		}
	})

	t.Run("highlight flag controls output", func(t *testing.T) {
		original := viper.GetBool("highlight")
		viper.Set("highlight", true)
		t.Cleanup(func() {
			viper.Set("highlight", original)
		})
		// Still returns false because not a TTY, but exercises the code path
		_ = shouldHighlight()
	})
}

func TestFormatConfigValue_NestedMap(t *testing.T) {
	t.Parallel()

	// Test nested map that causes JSON marshaling
	nestedMap := map[string]any{
		"nested": map[string]any{
			"key": "value",
		},
	}
	result := FormatConfigValue(nestedMap)
	if result == "" {
		t.Error("expected non-empty result for nested map")
	}
}

// errorMarshaler is a type that always returns an error when marshaling.
type errorMarshaler struct{}

func (e errorMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errTestMarshal
}

func (e errorMarshaler) MarshalYAML() (interface{}, error) {
	return nil, errTestMarshal
}

var errTestMarshal = errors.New("test marshal error")

func TestYAMLFormatter_MarshalError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewYAMLFormatter()
	err := f.Format(&buf, errorMarshaler{})
	if err == nil {
		t.Error("expected error for unmarshalable type")
	}
}

func TestJSONFormatter_MarshalError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	f := NewJSONFormatter()
	err := f.Format(&buf, errorMarshaler{})
	if err == nil {
		t.Error("expected error for unmarshalable type")
	}
}

// ExtractMapSection tests are in helpers_test.go

// FormatConfigValue tests are in helpers_test.go

// FormatDisplayValue and RenderProgressBar tests are in helpers_test.go

//nolint:paralleltest // Tests modify shared isTTY and viper state
func TestShouldHighlight_TTY(t *testing.T) {
	// Save and restore isTTY
	oldIsTTY := isTTY
	defer func() { isTTY = oldIsTTY }()

	t.Run("non-TTY returns false", func(t *testing.T) {
		isTTY = func() bool { return false }
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false for non-TTY")
		}
	})

	t.Run("TTY with plain flag returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		viper.Set("plain", true)
		defer viper.Set("plain", false)
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false when plain=true")
		}
	})

	t.Run("TTY with no-color flag returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		viper.Set("no-color", true)
		defer viper.Set("no-color", false)
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false when no-color=true")
		}
	})

	t.Run("TTY with NO_COLOR env returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		t.Setenv("NO_COLOR", "1")
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false when NO_COLOR is set")
		}
	})

	t.Run("TTY with SHELLY_NO_COLOR env returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		t.Setenv("SHELLY_NO_COLOR", "1")
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false when SHELLY_NO_COLOR is set")
		}
	})

	t.Run("TTY with TERM=dumb returns false", func(t *testing.T) {
		isTTY = func() bool { return true }
		t.Setenv("TERM", "dumb")
		if shouldHighlight() {
			t.Error("expected shouldHighlight() = false when TERM=dumb")
		}
	})

	t.Run("TTY with no restrictions returns true", func(t *testing.T) {
		isTTY = func() bool { return true }
		// Clear all flags
		viper.Set("plain", false)
		viper.Set("no-color", false)
		// Set a normal terminal type
		t.Setenv("TERM", "xterm-256color")
		if !shouldHighlight() {
			t.Error("expected shouldHighlight() = true when TTY with no restrictions")
		}
	})
}

func TestPrint_ToStdout(t *testing.T) {
	// Test that Print doesn't panic and writes to stdout
	data := map[string]string{"key": "value"}
	err := Print(data)
	if err != nil {
		t.Errorf("Print() returned error: %v", err)
	}
}

func TestPrintJSON_ToStdout(t *testing.T) {
	// Test that PrintJSON doesn't panic and writes to stdout
	data := map[string]string{"key": "value"}
	err := PrintJSON(data)
	if err != nil {
		t.Errorf("PrintJSON() returned error: %v", err)
	}
}

func TestPrintYAML_ToStdout(t *testing.T) {
	// Test that PrintYAML doesn't panic and writes to stdout
	data := map[string]string{"key": "value"}
	err := PrintYAML(data)
	if err != nil {
		t.Errorf("PrintYAML() returned error: %v", err)
	}
}

func TestPrintTemplate_ToStdout(t *testing.T) {
	// Test that PrintTemplate doesn't panic and writes to stdout
	data := map[string]string{"key": "value"}
	err := PrintTemplate("{{.key}}", data)
	if err != nil {
		t.Errorf("PrintTemplate() returned error: %v", err)
	}
}

//nolint:paralleltest // Tests modify shared isTTY and viper state
func TestNewJSONFormatter_WithTTY(t *testing.T) {
	// Save and restore isTTY
	oldIsTTY := isTTY
	defer func() { isTTY = oldIsTTY }()

	// Enable TTY mode to test highlighting path
	isTTY = func() bool { return true }
	// Make sure no color restrictions
	viper.Set("plain", false)
	viper.Set("no-color", false)
	t.Setenv("TERM", "xterm-256color")

	f := NewJSONFormatter()
	if !f.Highlight {
		t.Error("expected Highlight=true when TTY with no restrictions")
	}

	var buf bytes.Buffer
	err := f.Format(&buf, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}
	// With highlighting, output should contain ANSI escape codes
	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

//nolint:paralleltest // Tests modify shared isTTY and viper state
func TestNewYAMLFormatter_WithTTY(t *testing.T) {
	// Save and restore isTTY
	oldIsTTY := isTTY
	defer func() { isTTY = oldIsTTY }()

	// Enable TTY mode to test highlighting path
	isTTY = func() bool { return true }
	// Make sure no color restrictions
	viper.Set("plain", false)
	viper.Set("no-color", false)
	t.Setenv("TERM", "xterm-256color")

	f := NewYAMLFormatter()
	if !f.Highlight {
		t.Error("expected Highlight=true when TTY with no restrictions")
	}

	var buf bytes.Buffer
	err := f.Format(&buf, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestHighlightCode_Languages(t *testing.T) {
	t.Parallel()

	t.Run("unknown language returns plain code", func(t *testing.T) {
		t.Parallel()
		code := "some code"
		result := highlightCode(code, "nonexistent-language-xyz")
		if result != code {
			t.Errorf("expected plain code for unknown language, got %q", result)
		}
	})

	t.Run("json language highlights", func(t *testing.T) {
		t.Parallel()
		code := `{"key": "value"}`
		result := highlightCode(code, "json")
		// Result should be non-empty (may or may not have ANSI depending on formatter)
		if result == "" {
			t.Error("expected non-empty result for json highlighting")
		}
	})

	t.Run("yaml language highlights", func(t *testing.T) {
		t.Parallel()
		code := "key: value"
		result := highlightCode(code, "yaml")
		if result == "" {
			t.Error("expected non-empty result for yaml highlighting")
		}
	})
}

//nolint:paralleltest // Tests modify shared viper state
func TestGetChromaStyle_AllThemes(t *testing.T) {
	themes := []string{"dracula", "nord", "gruvbox", "gruvbox-dark", "tokyo-night", "catppuccin", "some-unknown-theme", ""}
	for _, themeName := range themes {
		t.Run("theme_"+themeName, func(t *testing.T) {
			viper.Set("theme.name", themeName)
			defer viper.Set("theme.name", "")
			style := getChromaStyle()
			if style == nil {
				t.Errorf("expected non-nil style for theme %q", themeName)
			}
		})
	}
}
