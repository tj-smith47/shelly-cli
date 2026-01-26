package term

import (
	"strings"
	"testing"
)

func TestDisplayDiagram(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diagram := "  Shelly Plus 1 (Gen2)\n  ─────────────────────\n\n  ┌────────────────┐\n  │ L            O │\n  └────────────────┘\n"

	DisplayDiagram(ios, diagram)

	output := out.String()
	if output == "" {
		t.Error("DisplayDiagram should produce output")
	}
	if !strings.Contains(output, "Shelly Plus 1") {
		t.Error("output should contain the rendered diagram content")
	}
}

func TestDisplayDiagram_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDiagram(ios, "")

	output := out.String()
	if !strings.Contains(output, "\n") {
		t.Error("DisplayDiagram with empty string should still call Println")
	}
}

func TestDisplayDiagram_Multiline(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diagram := "line 1\nline 2\nline 3\n"

	DisplayDiagram(ios, diagram)

	output := out.String()
	if !strings.Contains(output, "line 1") {
		t.Error("output should contain first line")
	}
	if !strings.Contains(output, "line 3") {
		t.Error("output should contain last line")
	}
}

func TestDisplayDiagramGenerationNote(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDiagramGenerationNote(ios, 4, []int{1, 3, 4})

	output := out.String()
	if !strings.Contains(output, "Gen4") {
		t.Error("output should mention selected generation")
	}
	if !strings.Contains(output, "Gen1") {
		t.Error("output should mention alternative Gen1")
	}
	if !strings.Contains(output, "Gen3") {
		t.Error("output should mention alternative Gen3")
	}
	if !strings.Contains(output, "-g") {
		t.Error("output should mention -g flag")
	}
}

func TestDisplayDiagramGenerationNote_TwoGens(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDiagramGenerationNote(ios, 4, []int{1, 4})

	output := out.String()
	if !strings.Contains(output, "Gen1") {
		t.Error("output should mention alternative Gen1")
	}
	if !strings.Contains(output, "latest") {
		t.Error("output should mention 'latest'")
	}
}

func TestDisplayDiagramGenerationNote_NoOthers(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDiagramGenerationNote(ios, 2, []int{2})

	output := out.String()
	if output != "" {
		t.Errorf("should produce no output when only one generation, got: %q", output)
	}
}
