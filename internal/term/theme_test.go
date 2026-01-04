package term

import (
	"strings"
	"testing"
)

func TestDisplayValidationResult_WithThemeName(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayValidationResult(ios, "nord", nil)

	output := out.String()
	if !strings.Contains(output, "validated successfully") {
		t.Error("expected success message")
	}
	if !strings.Contains(output, "Base theme: nord") {
		t.Error("expected base theme name")
	}
	if !strings.Contains(output, "--apply") {
		t.Error("expected apply hint")
	}
}

func TestDisplayValidationResult_WithColors(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	colors := map[string]string{
		"red":   "#ff0000",
		"green": "#00ff00",
	}
	DisplayValidationResult(ios, "", colors)

	output := out.String()
	if !strings.Contains(output, "Color overrides: 2") {
		t.Error("expected color override count")
	}
}

func TestDisplayValidationResult_Minimal(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayValidationResult(ios, "", nil)

	output := out.String()
	if !strings.Contains(output, "validated successfully") {
		t.Error("expected success message")
	}
	// Should not show base theme when empty
	if strings.Contains(output, "Base theme:") {
		t.Error("should not show base theme when empty")
	}
}
