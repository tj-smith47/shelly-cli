package modal

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.visible {
		t.Error("expected modal to be hidden by default")
	}
	if m.title != "Dialog" {
		t.Errorf("expected default title 'Dialog', got %q", m.title)
	}
	if !m.closeOnEsc {
		t.Error("expected closeOnEsc to be true by default")
	}
	if !m.confirmOnEnter {
		t.Error("expected confirmOnEnter to be true by default")
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Parallel()
	m := New(
		WithTitle("Test Modal"),
		WithContent("Test content"),
		WithFooter("Press Enter to confirm"),
		WithCloseOnEsc(false),
		WithConfirmOnEnter(false),
	)

	if m.title != "Test Modal" {
		t.Errorf("expected title 'Test Modal', got %q", m.title)
	}
	if m.content != "Test content" {
		t.Errorf("expected content 'Test content', got %q", m.content)
	}
	if m.footer != "Press Enter to confirm" {
		t.Errorf("expected footer 'Press Enter to confirm', got %q", m.footer)
	}
	if m.closeOnEsc {
		t.Error("expected closeOnEsc to be false")
	}
	if m.confirmOnEnter {
		t.Error("expected confirmOnEnter to be false")
	}
}

func TestSetters(t *testing.T) {
	t.Parallel()
	m := New()

	m = m.SetTitle("New Title")
	if m.Title() != "New Title" {
		t.Errorf("expected title 'New Title', got %q", m.Title())
	}

	m = m.SetContent("New Content")
	if m.Content() != "New Content" {
		t.Errorf("expected content 'New Content', got %q", m.Content())
	}

	m = m.SetFooter("New Footer")
	if m.Footer() != "New Footer" {
		t.Errorf("expected footer 'New Footer', got %q", m.Footer())
	}
}

func TestVisibility(t *testing.T) {
	t.Parallel()
	m := New()

	if m.IsVisible() {
		t.Error("expected modal to be hidden by default")
	}

	m = m.Show()
	if !m.IsVisible() {
		t.Error("expected modal to be visible after Show()")
	}

	m = m.Hide()
	if m.IsVisible() {
		t.Error("expected modal to be hidden after Hide()")
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestViewWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when hidden, got %q", view)
	}
}

func TestViewWhenVisible(t *testing.T) {
	t.Parallel()
	m := New(
		WithTitle("Test"),
		WithContent("Content here"),
	)
	m = m.SetSize(80, 24)
	m = m.Show()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when visible")
	}
	if !strings.Contains(view, "Test") {
		t.Error("expected view to contain title")
	}
	if !strings.Contains(view, "Content here") {
		t.Error("expected view to contain content")
	}
}

func TestEscapeClosesModal(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	newM, cmd := m.Update(keyMsg)

	if newM.IsVisible() {
		t.Error("expected modal to be hidden after Esc")
	}
	if cmd == nil {
		t.Error("expected CloseMsg command")
	}
}

func TestEscapeDisabled(t *testing.T) {
	t.Parallel()
	m := New(WithCloseOnEsc(false))
	m = m.Show()

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	newM, cmd := m.Update(keyMsg)

	if !newM.IsVisible() {
		t.Error("expected modal to remain visible when closeOnEsc is false")
	}
	if cmd != nil {
		t.Error("expected no command when closeOnEsc is false")
	}
}

func TestEnterConfirms(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newM, cmd := m.Update(keyMsg)

	if newM.IsVisible() {
		t.Error("expected modal to be hidden after Enter")
	}
	if cmd == nil {
		t.Error("expected CloseMsg command")
	}
}

func TestEnterDisabled(t *testing.T) {
	t.Parallel()
	m := New(WithConfirmOnEnter(false))
	m = m.Show()

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newM, cmd := m.Update(keyMsg)

	if !newM.IsVisible() {
		t.Error("expected modal to remain visible when confirmOnEnter is false")
	}
	if cmd != nil {
		t.Error("expected no command when confirmOnEnter is false")
	}
}

func TestScrolling(t *testing.T) {
	t.Parallel()
	longContent := strings.Repeat("Line\n", 100)
	m := New(WithContent(longContent))
	m = m.SetSize(80, 24)
	m = m.Show()

	// Scroll down
	keyMsg := tea.KeyPressMsg{Code: tea.KeyDown}
	newM, _ := m.Update(keyMsg)
	if newM.scrollOffset == 0 {
		t.Error("expected scrollOffset to increase after down key")
	}

	// Scroll up
	keyMsg = tea.KeyPressMsg{Code: tea.KeyUp}
	newM, _ = newM.Update(keyMsg)
	if newM.scrollOffset != 0 {
		t.Error("expected scrollOffset to return to 0 after up key")
	}
}

func TestScrollToEnd(t *testing.T) {
	t.Parallel()
	longContent := strings.Repeat("Line\n", 100)
	m := New(WithContent(longContent))
	m = m.SetSize(80, 24)
	m = m.Show()

	// Go to end
	keyMsg := tea.KeyPressMsg{Code: 'G'}
	newM, _ := m.Update(keyMsg)
	if newM.scrollOffset == 0 {
		t.Error("expected scrollOffset to be at end after G key")
	}

	// Go to start
	keyMsg = tea.KeyPressMsg{Code: 'g'}
	newM, _ = newM.Update(keyMsg)
	if newM.scrollOffset != 0 {
		t.Error("expected scrollOffset to be at start after g key")
	}
}

func TestInit(t *testing.T) {
	t.Parallel()
	m := New()
	cmd := m.Init()

	if cmd != nil {
		t.Error("expected Init to return nil command")
	}
}

func TestUpdateWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newM, cmd := m.Update(keyMsg)

	if cmd != nil {
		t.Error("expected no command when modal is hidden")
	}
	if newM.visible {
		t.Error("expected modal to remain hidden")
	}
}

func TestDefaultSize(t *testing.T) {
	t.Parallel()
	size := DefaultSize()

	if size.WidthPct != 60 {
		t.Errorf("expected WidthPct 60, got %d", size.WidthPct)
	}
	if size.HeightPct != 50 {
		t.Errorf("expected HeightPct 50, got %d", size.HeightPct)
	}
	if size.MinWidth != 40 {
		t.Errorf("expected MinWidth 40, got %d", size.MinWidth)
	}
	if size.MinHeight != 10 {
		t.Errorf("expected MinHeight 10, got %d", size.MinHeight)
	}
}

func TestWithSize(t *testing.T) {
	t.Parallel()
	customSize := Size{
		Width:  80,
		Height: 30,
	}
	m := New(WithSize(customSize))

	if m.size.Width != 80 {
		t.Errorf("expected Width 80, got %d", m.size.Width)
	}
	if m.size.Height != 30 {
		t.Errorf("expected Height 30, got %d", m.size.Height)
	}
}

func TestOverlayWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()
	base := "Base content"

	result := m.Overlay(base)
	if result != base {
		t.Error("expected Overlay to return base when modal is hidden")
	}
}

func TestOverlayWhenVisible(t *testing.T) {
	t.Parallel()
	m := New(WithTitle("Modal"))
	m = m.SetSize(80, 24)
	m = m.Show()
	base := "Base content"

	result := m.Overlay(base)
	if result == base {
		t.Error("expected Overlay to return modal view when visible")
	}
	if !strings.Contains(result, "Modal") {
		t.Error("expected Overlay result to contain modal title")
	}
}

func TestCalculateWidthWithFixedSize(t *testing.T) {
	t.Parallel()
	m := New(WithSize(Size{Width: 60}))
	m = m.SetSize(100, 50)

	width := m.calculateWidth()
	if width != 60 {
		t.Errorf("expected width 60, got %d", width)
	}
}

func TestCalculateWidthWithPercentage(t *testing.T) {
	t.Parallel()
	m := New(WithSize(Size{WidthPct: 50, MinWidth: 20, MaxWidth: 100}))
	m = m.SetSize(100, 50)

	width := m.calculateWidth()
	if width != 50 {
		t.Errorf("expected width 50 (50%% of 100), got %d", width)
	}
}

func TestCalculateHeightWithFixedSize(t *testing.T) {
	t.Parallel()
	m := New(WithSize(Size{Height: 25}))
	m = m.SetSize(100, 50)

	height := m.calculateHeight()
	if height != 25 {
		t.Errorf("expected height 25, got %d", height)
	}
}

func TestClampWidth(t *testing.T) {
	t.Parallel()
	m := New(WithSize(Size{
		Width:    200,
		MaxWidth: 80,
	}))
	m = m.SetSize(100, 50)

	width := m.calculateWidth()
	if width != 80 {
		t.Errorf("expected width clamped to 80, got %d", width)
	}
}

func TestClampHeight(t *testing.T) {
	t.Parallel()
	m := New(WithSize(Size{
		Height:    100,
		MaxHeight: 30,
	}))
	m = m.SetSize(100, 50)

	height := m.calculateHeight()
	if height != 30 {
		t.Errorf("expected height clamped to 30, got %d", height)
	}
}

func TestContentHeightTracking(t *testing.T) {
	t.Parallel()
	content := "Line 1\nLine 2\nLine 3"
	m := New(WithContent(content))

	if m.contentHeight != 3 {
		t.Errorf("expected contentHeight 3, got %d", m.contentHeight)
	}

	m = m.SetContent("Single line")
	if m.contentHeight != 1 {
		t.Errorf("expected contentHeight 1, got %d", m.contentHeight)
	}
}

func TestShowResetsScroll(t *testing.T) {
	t.Parallel()
	m := New(WithContent(strings.Repeat("Line\n", 100)))
	m = m.SetSize(80, 24)
	m = m.Show()
	m.scrollOffset = 50

	m = m.Show()
	if m.scrollOffset != 0 {
		t.Errorf("expected Show to reset scrollOffset, got %d", m.scrollOffset)
	}
}
