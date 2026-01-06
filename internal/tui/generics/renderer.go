// Package generics provides generic utilities for TUI components.
package generics

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// RenderScrollableItems renders visible items from a scroller using the provided render function.
// It handles the visible range loop and newlines between items.
// The renderItem function receives the item, its index, and whether it's at the cursor position.
// Bounds are automatically clamped to prevent index out of range errors.
func RenderScrollableItems[T any](
	items []T,
	scroller *panel.Scroller,
	renderItem func(item T, index int, isCursor bool) string,
) string {
	if len(items) == 0 {
		return ""
	}

	var content strings.Builder
	start, end := scroller.VisibleRange()

	// Clamp bounds to slice length for safety
	if end > len(items) {
		end = len(items)
	}

	for i := start; i < end; i++ {
		isCursor := scroller.IsCursorAt(i)
		content.WriteString(renderItem(items[i], i, isCursor))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

// ScrollInfoMode determines when to display scroll info.
type ScrollInfoMode int

const (
	// ScrollInfoAlways always displays scroll info.
	ScrollInfoAlways ScrollInfoMode = iota
	// ScrollInfoWhenNeeded only displays scroll info when there's content to scroll.
	ScrollInfoWhenNeeded
)

// AppendScrollInfo appends scroll info to the content builder.
// Mode determines whether to always show or only when there's content to scroll.
func AppendScrollInfo(content *strings.Builder, scroller *panel.Scroller, style lipgloss.Style, mode ScrollInfoMode) {
	switch mode {
	case ScrollInfoAlways:
		content.WriteString("\n")
		content.WriteString(style.Render(scroller.ScrollInfo()))
	case ScrollInfoWhenNeeded:
		if scroller.HasMore() || scroller.HasPrevious() {
			content.WriteString("\n")
			content.WriteString(style.Render(scroller.ScrollInfo()))
		}
	}
}

// ListRenderConfig holds configuration for rendering a scrollable list.
type ListRenderConfig[T any] struct {
	Items          []T
	Scroller       *panel.Scroller
	RenderItem     func(item T, index int, isCursor bool) string
	ScrollStyle    lipgloss.Style
	ScrollInfoMode ScrollInfoMode
}

// RenderScrollableList renders a complete scrollable list with items and scroll info.
// This is a convenience function that combines RenderScrollableItems and AppendScrollInfo.
func RenderScrollableList[T any](cfg ListRenderConfig[T]) string {
	if len(cfg.Items) == 0 {
		return ""
	}

	var content strings.Builder
	content.WriteString(RenderScrollableItems(cfg.Items, cfg.Scroller, cfg.RenderItem))
	AppendScrollInfo(&content, cfg.Scroller, cfg.ScrollStyle, cfg.ScrollInfoMode)

	return content.String()
}
