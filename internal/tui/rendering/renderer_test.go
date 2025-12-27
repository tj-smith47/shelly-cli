package rendering

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestNew(t *testing.T) {
	t.Parallel()
	r := New(40, 10)
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.width != 40 {
		t.Errorf("width = %d, want 40", r.width)
	}
	if r.height != 10 {
		t.Errorf("height = %d, want 10", r.height)
	}
}

func TestRenderer_SetTitle(t *testing.T) {
	t.Parallel()
	r := New(40, 10).SetTitle("Test Title")
	if r.title != "Test Title" {
		t.Errorf("title = %q, want %q", r.title, "Test Title")
	}
}

func TestRenderer_SetFocused(t *testing.T) {
	t.Parallel()
	r := New(40, 10).SetFocused(true)
	if !r.focused {
		t.Error("focused should be true")
	}
}

func TestRenderer_AddSection(t *testing.T) {
	t.Parallel()
	r := New(40, 10).
		AddSection("First", "content1").
		AddSection("Second", "content2")

	if len(r.sections) != 2 {
		t.Errorf("sections count = %d, want 2", len(r.sections))
	}
	if r.sections[0].name != "First" {
		t.Errorf("section[0].name = %q, want %q", r.sections[0].name, "First")
	}
}

func TestRenderer_SetContent(t *testing.T) {
	t.Parallel()
	r := New(40, 10).SetContent("Test content")
	if r.content != "Test content" {
		t.Errorf("content = %q, want %q", r.content, "Test content")
	}
}

func TestRenderer_Render_Empty(t *testing.T) {
	t.Parallel()
	r := New(20, 5)
	output := r.Render()

	if output == "" {
		t.Error("Render() returned empty string for valid dimensions")
	}

	// Should have borders
	if !strings.Contains(output, "╭") || !strings.Contains(output, "╯") {
		t.Error("Render() missing rounded border characters")
	}
}

func TestRenderer_Render_WithTitle(t *testing.T) {
	t.Parallel()
	r := New(40, 10).SetTitle("Test")
	output := r.Render()

	// Title should appear in the output
	if !strings.Contains(output, "Test") {
		t.Error("Render() missing title in output")
	}

	// Should have title decorators
	if !strings.Contains(output, "├") || !strings.Contains(output, "┤") {
		t.Error("Render() missing title decorator characters")
	}
}

func TestRenderer_Render_WithContent(t *testing.T) {
	t.Parallel()
	r := New(40, 10).SetContent("Hello World")
	output := r.Render()

	if !strings.Contains(output, "Hello World") {
		t.Error("Render() missing content in output")
	}
}

func TestRenderer_Render_TooSmall(t *testing.T) {
	t.Parallel()
	tests := []struct {
		width  int
		height int
	}{
		{2, 10},
		{40, 1},
		{0, 0},
	}

	for _, tt := range tests {
		r := New(tt.width, tt.height)
		output := r.Render()
		if output != "" {
			t.Errorf("Render(%d, %d) = %q, want empty", tt.width, tt.height, output)
		}
	}
}

func TestRenderer_ContentDimensions(t *testing.T) {
	t.Parallel()
	r := New(40, 20)

	// ContentWidth = width - 4 (2 for borders + 2 for horizontal padding)
	if r.ContentWidth() != 36 {
		t.Errorf("ContentWidth() = %d, want 36", r.ContentWidth())
	}
	// ContentHeight = height - 4 (2 for borders + 2 for vertical padding)
	if r.ContentHeight() != 16 {
		t.Errorf("ContentHeight() = %d, want 16", r.ContentHeight())
	}
}

func TestBuildTopBorder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		width    int
		title    string
		wantLen  int
		contains []string
	}{
		{
			name:     "no title",
			width:    20,
			title:    "",
			wantLen:  20,
			contains: []string{"╭", "╮", "─"},
		},
		{
			name:     "with title",
			width:    30,
			title:    "Test",
			wantLen:  30,
			contains: []string{"╭", "╮", "├", "┤", "Test"},
		},
		{
			name:    "too small",
			width:   3,
			title:   "Test",
			wantLen: 0,
		},
	}

	border := lipgloss.RoundedBorder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := BuildTopBorder(tt.width, tt.title, border)

			if tt.wantLen == 0 {
				if result != "" {
					t.Errorf("BuildTopBorder() = %q, want empty", result)
				}
				return
			}

			resultWidth := ansi.StringWidth(result)
			if resultWidth != tt.wantLen {
				t.Errorf("StringWidth(BuildTopBorder()) = %d, want %d", resultWidth, tt.wantLen)
			}

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("BuildTopBorder() missing %q in %q", s, result)
				}
			}
		})
	}
}

func TestBuildBottomBorder(t *testing.T) {
	t.Parallel()
	border := lipgloss.RoundedBorder()
	result := BuildBottomBorder(20, border)

	resultWidth := ansi.StringWidth(result)
	if resultWidth != 20 {
		t.Errorf("StringWidth(BuildBottomBorder()) = %d, want 20", resultWidth)
	}

	if !strings.Contains(result, "╰") || !strings.Contains(result, "╯") {
		t.Error("BuildBottomBorder() missing corner characters")
	}
}

func TestBuildDivider(t *testing.T) {
	t.Parallel()
	border := lipgloss.RoundedBorder()

	tests := []struct {
		name     string
		width    int
		secName  string
		contains []string
	}{
		{
			name:     "with name",
			width:    30,
			secName:  "Section",
			contains: []string{"├", "┤", "Section"},
		},
		{
			name:     "no name",
			width:    20,
			secName:  "",
			contains: []string{"├", "┤"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := BuildDivider(tt.width, tt.secName, border)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("BuildDivider() missing %q in %q", s, result)
				}
			}
		})
	}
}
