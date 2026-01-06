package generics

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

func TestRenderScrollableItems(t *testing.T) {
	t.Parallel()

	items := []string{"a", "b", "c", "d", "e"}
	scroller := panel.NewScroller(len(items), 3)

	render := func(item string, index int, isCursor bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return prefix + item
	}

	result := RenderScrollableItems(items, scroller, render)

	// Should render first 3 items (visible rows = 3)
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), result)
	}
	if lines[0] != "> a" {
		t.Errorf("expected first line '> a', got %q", lines[0])
	}
	if lines[1] != "  b" {
		t.Errorf("expected second line '  b', got %q", lines[1])
	}
	if lines[2] != "  c" {
		t.Errorf("expected third line '  c', got %q", lines[2])
	}
}

func TestRenderScrollableItems_Empty(t *testing.T) {
	t.Parallel()

	var items []string
	scroller := panel.NewScroller(0, 3)

	render := func(item string, index int, isCursor bool) string {
		return item
	}

	result := RenderScrollableItems(items, scroller, render)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderScrollableItems_ScrolledDown(t *testing.T) {
	t.Parallel()

	items := []string{"a", "b", "c", "d", "e"}
	scroller := panel.NewScroller(len(items), 3)

	// Move cursor down twice
	scroller.CursorDown()
	scroller.CursorDown()

	render := func(item string, index int, isCursor bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return prefix + item
	}

	result := RenderScrollableItems(items, scroller, render)

	// Cursor at index 2 (c), should show a, b, c
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), result)
	}
	if lines[2] != "> c" {
		t.Errorf("expected cursor at 'c', got %q", lines[2])
	}
}

func TestAppendScrollInfo_Always(t *testing.T) {
	t.Parallel()

	scroller := panel.NewScroller(10, 3)
	style := lipgloss.NewStyle()

	var content strings.Builder
	content.WriteString("items")
	AppendScrollInfo(&content, scroller, style, ScrollInfoAlways)

	result := content.String()
	if !strings.Contains(result, "items\n") {
		t.Errorf("expected newline before scroll info, got %q", result)
	}
	// ScrollInfo should be appended
	if !strings.Contains(result, "/") {
		t.Errorf("expected scroll info with / separator, got %q", result)
	}
}

func TestAppendScrollInfo_WhenNeeded_HasMore(t *testing.T) {
	t.Parallel()

	// 10 items, 3 visible = has more
	scroller := panel.NewScroller(10, 3)
	style := lipgloss.NewStyle()

	var content strings.Builder
	content.WriteString("items")
	AppendScrollInfo(&content, scroller, style, ScrollInfoWhenNeeded)

	result := content.String()
	// Should have scroll info because there's more content
	if !strings.Contains(result, "/") {
		t.Errorf("expected scroll info when has more, got %q", result)
	}
}

func TestAppendScrollInfo_WhenNeeded_NoMore(t *testing.T) {
	t.Parallel()

	// 2 items, 5 visible = no scrolling needed
	scroller := panel.NewScroller(2, 5)
	style := lipgloss.NewStyle()

	var content strings.Builder
	content.WriteString("items")
	AppendScrollInfo(&content, scroller, style, ScrollInfoWhenNeeded)

	result := content.String()
	// Should NOT have scroll info because everything fits
	if result != "items" {
		t.Errorf("expected no scroll info when not needed, got %q", result)
	}
}

func TestRenderScrollableList(t *testing.T) {
	t.Parallel()

	items := []string{"a", "b", "c", "d", "e"}
	scroller := panel.NewScroller(len(items), 3)
	style := lipgloss.NewStyle()

	render := func(item string, index int, isCursor bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return prefix + item
	}

	result := RenderScrollableList(ListRenderConfig[string]{
		Items:          items,
		Scroller:       scroller,
		RenderItem:     render,
		ScrollStyle:    style,
		ScrollInfoMode: ScrollInfoAlways,
	})

	lines := strings.Split(result, "\n")
	// 3 items + scroll info line = 4 lines
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d: %q", len(lines), result)
	}
}

func TestRenderScrollableList_Empty(t *testing.T) {
	t.Parallel()

	var items []string
	scroller := panel.NewScroller(0, 3)
	style := lipgloss.NewStyle()

	render := func(item string, index int, isCursor bool) string {
		return item
	}

	result := RenderScrollableList(ListRenderConfig[string]{
		Items:          items,
		Scroller:       scroller,
		RenderItem:     render,
		ScrollStyle:    style,
		ScrollInfoMode: ScrollInfoAlways,
	})

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
