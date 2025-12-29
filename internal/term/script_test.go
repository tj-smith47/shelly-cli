package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

func TestDisplayScriptEvalResult_Nil(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptEvalResult(ios, nil)

	output := out.String()
	if !strings.Contains(output, "no result") {
		t.Error("expected no result message")
	}
}

func TestDisplayScriptEvalResult_String(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptEvalResult(ios, "Hello World")

	output := out.String()
	if !strings.Contains(output, "Hello World") {
		t.Error("expected string result")
	}
}

func TestDisplayScriptEvalResult_WholeNumber(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptEvalResult(ios, float64(42))

	output := out.String()
	if !strings.Contains(output, "42") {
		t.Error("expected integer display")
	}
	// Should not contain decimal point for whole numbers
	if strings.Contains(output, ".0") {
		t.Error("should not show decimal for whole numbers")
	}
}

func TestDisplayScriptEvalResult_Float(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptEvalResult(ios, 3.14159)

	output := out.String()
	if !strings.Contains(output, "3.14") {
		t.Error("expected float value")
	}
}

func TestDisplayScriptEvalResult_Bool(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptEvalResult(ios, true)

	output := out.String()
	if !strings.Contains(output, "true") {
		t.Error("expected boolean result")
	}
}

func TestDisplayScriptEvalResult_Object(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := map[string]any{
		"key":   "value",
		"count": float64(10),
	}
	DisplayScriptEvalResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "key") {
		t.Error("expected key in JSON output")
	}
	if !strings.Contains(output, "value") {
		t.Error("expected value in JSON output")
	}
}

func TestDisplayScriptStatus_Running(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &automation.ScriptStatus{
		ID:       1,
		Running:  true,
		MemUsage: 1024,
		MemPeak:  2048,
		MemFree:  8192,
		Errors:   []string{},
	}
	DisplayScriptStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Script Status") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "ID:      1") {
		t.Error("expected script ID")
	}
	if !strings.Contains(output, "Memory") {
		t.Error("expected memory section")
	}
	if !strings.Contains(output, "1024 bytes") {
		t.Error("expected memory usage")
	}
}

func TestDisplayScriptStatus_WithErrors(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &automation.ScriptStatus{
		ID:      2,
		Running: false,
		Errors:  []string{"Syntax error at line 10", "Undefined variable 'x'"},
	}
	DisplayScriptStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Errors") {
		t.Error("expected errors section")
	}
	if !strings.Contains(output, "Syntax error") {
		t.Error("expected first error")
	}
	if !strings.Contains(output, "Undefined variable") {
		t.Error("expected second error")
	}
}

func TestDisplayScriptCode_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptCode(ios, "")

	output := out.String()
	if !strings.Contains(output, "no code") {
		t.Error("expected no code message")
	}
}

func TestDisplayScriptCode_WithCode(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	code := "function main() {\n  print('Hello');\n}"
	DisplayScriptCode(ios, code)

	output := out.String()
	if !strings.Contains(output, "function main()") {
		t.Error("expected code output")
	}
	if !strings.Contains(output, "print('Hello')") {
		t.Error("expected code content")
	}
}

func TestDisplayScriptTemplateList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	templates := []config.ScriptTemplate{
		{Name: "timer", Category: "automation", Description: "Timer script", BuiltIn: true},
		{Name: "webhook", Category: "integration", Description: "Webhook handler", BuiltIn: false},
	}
	DisplayScriptTemplateList(ios, templates)

	output := out.String()
	if !strings.Contains(output, "timer") {
		t.Error("expected template name")
	}
	if !strings.Contains(output, "automation") {
		t.Error("expected category")
	}
	if !strings.Contains(output, "built-in") {
		t.Error("expected built-in source")
	}
	if !strings.Contains(output, "user") {
		t.Error("expected user source")
	}
	if !strings.Contains(output, "2") {
		t.Error("expected count")
	}
}

func TestDisplayScriptTemplate_Full(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	tpl := config.ScriptTemplate{
		Name:        "my-template",
		Description: "A test template",
		Category:    "test",
		Author:      "Test Author",
		Version:     "1.0.0",
		MinGen:      2,
		BuiltIn:     true,
		Variables: []config.ScriptVariable{
			{Name: "interval", Type: "number", Required: true, Description: "Interval in seconds"},
			{Name: "name", Type: "string", Required: false, Default: "default"},
		},
		Code: "// Template code here",
	}
	DisplayScriptTemplate(ios, tpl)

	output := out.String()
	if !strings.Contains(output, "my-template") {
		t.Error("expected template name")
	}
	if !strings.Contains(output, "A test template") {
		t.Error("expected description")
	}
	if !strings.Contains(output, "Test Author") {
		t.Error("expected author")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("expected version")
	}
	if !strings.Contains(output, "Min Gen:      2") {
		t.Error("expected min gen")
	}
	if !strings.Contains(output, "Variables") {
		t.Error("expected variables section")
	}
	if !strings.Contains(output, "interval") {
		t.Error("expected variable name")
	}
	if !strings.Contains(output, "(required)") {
		t.Error("expected required marker")
	}
	if !strings.Contains(output, "[default:") {
		t.Error("expected default value")
	}
	if !strings.Contains(output, "// Template code here") {
		t.Error("expected code")
	}
}

func TestDisplayScriptTemplate_Minimal(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	tpl := config.ScriptTemplate{
		Name: "minimal",
		Code: "print('hello');",
	}
	DisplayScriptTemplate(ios, tpl)

	output := out.String()
	if !strings.Contains(output, "minimal") {
		t.Error("expected template name")
	}
	if !strings.Contains(output, "user-defined") {
		t.Error("expected user-defined source")
	}
}
