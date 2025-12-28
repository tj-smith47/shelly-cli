package theme

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

// TestGetSyntaxColors verifies syntax colors are accessible.
func TestGetSyntaxColors(t *testing.T) {
	t.Parallel()

	colors := GetSyntaxColors()

	// Verify key colors are set (not nil)
	if colors.Background == nil {
		t.Error("SyntaxColors.Background is nil")
	}
	if colors.Text == nil {
		t.Error("SyntaxColors.Text is nil")
	}
	if colors.Keyword == nil {
		t.Error("SyntaxColors.Keyword is nil")
	}
	if colors.String == nil {
		t.Error("SyntaxColors.String is nil")
	}
	if colors.Number == nil {
		t.Error("SyntaxColors.Number is nil")
	}
	if colors.Comment == nil {
		t.Error("SyntaxColors.Comment is nil")
	}
	if colors.Function == nil {
		t.Error("SyntaxColors.Function is nil")
	}
	if colors.Operator == nil {
		t.Error("SyntaxColors.Operator is nil")
	}
	if colors.Type == nil {
		t.Error("SyntaxColors.Type is nil")
	}
	if colors.Constant == nil {
		t.Error("SyntaxColors.Constant is nil")
	}
}

// TestTokenTypeToColor verifies token type to color mapping.
func TestTokenTypeToColor(t *testing.T) {
	t.Parallel()

	colors := GetSyntaxColors()

	tests := []struct {
		name      string
		tokenType chroma.TokenType
		wantColor string // field name in SyntaxColors
	}{
		{"comment", chroma.Comment, "Comment"},
		{"comment_single", chroma.CommentSingle, "Comment"},
		{"comment_multiline", chroma.CommentMultiline, "Comment"},
		{"keyword", chroma.Keyword, "Keyword"},
		{"keyword_declaration", chroma.KeywordDeclaration, "Keyword"},
		{"keyword_type", chroma.KeywordType, "Keyword"},
		{"string", chroma.String, "String"},
		{"string_double", chroma.StringDouble, "String"},
		{"string_single", chroma.StringSingle, "String"},
		{"number", chroma.Number, "Number"},
		{"number_integer", chroma.NumberInteger, "Number"},
		{"number_float", chroma.NumberFloat, "Number"},
		{"operator", chroma.Operator, "Operator"},
		{"name_function", chroma.NameFunction, "Function"},
		{"name_builtin", chroma.NameBuiltin, "Function"},
		{"name_class", chroma.NameClass, "Type"},
		{"name_constant", chroma.NameConstant, "Constant"},
		{"keyword_constant", chroma.KeywordConstant, "Constant"},
		{"text", chroma.Text, "Text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tokenTypeToColor(tt.tokenType, colors)
			if result == nil {
				t.Errorf("tokenTypeToColor(%v) returned nil", tt.tokenType)
			}
		})
	}
}

// TestHighlightCode verifies code highlighting.
func TestHighlightCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		language string
	}{
		{"javascript_simple", "const x = 1;", "javascript"},
		{"javascript_function", "function hello() { return 'world'; }", "javascript"},
		{"json_object", `{"key": "value"}`, "json"},
		{"json_array", `[1, 2, 3]`, "json"},
		{"go_code", "package main\nfunc main() {}", "go"},
		{"unknown_language", "some text", "nonexistent_lang"},
		{"empty_code", "", "javascript"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := HighlightCode(tt.code, tt.language)
			// Result should not panic and should return something
			// Even empty code returns empty string
			if tt.code != "" && result == "" {
				t.Errorf("HighlightCode(%q, %q) returned empty string", tt.code, tt.language)
			}
		})
	}
}

// TestHighlightCodeWithNewlines verifies code with newlines is handled correctly.
func TestHighlightCodeWithNewlines(t *testing.T) {
	t.Parallel()

	code := `function hello() {
    console.log("world");
    return true;
}`
	result := HighlightCode(code, "javascript")

	// Should contain newlines
	if !strings.Contains(result, "\n") {
		t.Error("HighlightCode should preserve newlines")
	}
}

// TestHighlightJavaScript verifies the convenience function.
func TestHighlightJavaScript(t *testing.T) {
	t.Parallel()

	code := "const x = 42;"
	result := HighlightJavaScript(code)

	if result == "" {
		t.Error("HighlightJavaScript returned empty string")
	}
}

// TestHighlightJSON verifies the convenience function.
func TestHighlightJSON(t *testing.T) {
	t.Parallel()

	code := `{"name": "test", "value": 123}`
	result := HighlightJSON(code)

	if result == "" {
		t.Error("HighlightJSON returned empty string")
	}
}

// TestColorToHexSyntax verifies the colorToHex helper in syntax.go.
func TestColorToHexSyntax(t *testing.T) {
	t.Parallel()

	t.Run("nil_color", func(t *testing.T) {
		t.Parallel()
		result := colorToHex(nil)
		if result != "#ffffff" {
			t.Errorf("colorToHex(nil) = %q, want #ffffff", result)
		}
	})

	t.Run("theme_color", func(t *testing.T) {
		t.Parallel()
		c := Green()
		result := colorToHex(c)
		if result == "" {
			t.Error("colorToHex(Green()) returned empty string")
		}
		if !strings.HasPrefix(result, "#") {
			t.Errorf("colorToHex() = %q, expected to start with #", result)
		}
		// Should be 7 characters (#rrggbb)
		if len(result) != 7 {
			t.Errorf("colorToHex() = %q, expected length 7", result)
		}
	})
}

// TestSyntaxColorsStruct verifies SyntaxColors struct has expected fields.
func TestSyntaxColorsStruct(t *testing.T) {
	t.Parallel()

	colors := SyntaxColors{
		Background: Green(),
		Text:       Fg(),
		Keyword:    Purple(),
		String:     Yellow(),
		Number:     Cyan(),
		Comment:    BrightBlack(),
		Function:   Green(),
		Operator:   Red(),
		Type:       Cyan(),
		Constant:   Purple(),
	}

	// Verify all fields are accessible
	if colors.Background == nil {
		t.Error("Background not set")
	}
	if colors.Keyword == nil {
		t.Error("Keyword not set")
	}
}

// TestHighlightCodeAllTokenTypes tests various token types through actual code.
func TestHighlightCodeAllTokenTypes(t *testing.T) {
	t.Parallel()

	// JavaScript code that exercises many token types
	code := `// This is a comment
const PI = 3.14159;
let name = "hello";
function greet(x) {
    if (x > 0) {
        return true;
    }
    return false;
}
class MyClass {
    constructor() {}
}`

	result := HighlightJavaScript(code)

	// Should have content
	if result == "" {
		t.Error("HighlightJavaScript returned empty for complex code")
	}

	// Should preserve structure (has newlines)
	if strings.Count(result, "\n") < 5 {
		t.Error("Expected multiple newlines in highlighted code")
	}
}
