package errorview

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")
	m := New(err)

	if m.Message() != "test error" {
		t.Errorf("expected message 'test error', got %q", m.Message())
	}
	if !errors.Is(m.Error(), err) {
		t.Error("expected error to match")
	}
	if m.mode != ModeInline {
		t.Error("expected default mode to be ModeInline")
	}
	if m.severity != SeverityError {
		t.Error("expected default severity to be SeverityError")
	}
}

func TestNewFromMessage(t *testing.T) {
	t.Parallel()
	m := NewFromMessage("direct message")

	if m.Message() != "direct message" {
		t.Errorf("expected message 'direct message', got %q", m.Message())
	}
	if m.Error() != nil {
		t.Error("expected nil error")
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err,
		WithMode(ModeBanner),
		WithSeverity(SeverityWarning),
		WithDismissible(true),
		WithDetails("Additional details"),
		WithWidth(100),
	)

	if m.mode != ModeBanner {
		t.Error("expected mode ModeBanner")
	}
	if m.severity != SeverityWarning {
		t.Error("expected severity SeverityWarning")
	}
	if !m.dismissible {
		t.Error("expected dismissible")
	}
	if m.details != "Additional details" {
		t.Errorf("expected details, got %q", m.details)
	}
	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
}

func TestViewModes(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")

	modes := []DisplayMode{ModeBanner, ModeInline, ModeCompact, ModeDetailed}
	for _, mode := range modes {
		m := New(err, WithMode(mode))
		view := m.View()
		if view == "" {
			t.Errorf("expected non-empty view for mode %d", mode)
		}
		if !strings.Contains(view, "test error") {
			t.Errorf("expected view to contain error message for mode %d", mode)
		}
	}
}

func TestSeverityStyles(t *testing.T) {
	t.Parallel()
	err := errors.New("test")

	severities := []Severity{SeverityError, SeverityWarning, SeverityCritical}
	for _, severity := range severities {
		m := New(err, WithSeverity(severity))
		view := m.View()
		if view == "" {
			t.Errorf("expected non-empty view for severity %d", severity)
		}
	}
}

func TestDismissible(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err, WithDismissible(true))

	if m.IsDismissed() {
		t.Error("expected not dismissed initially")
	}

	// Press Esc to dismiss
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m, _ = m.Update(keyMsg)

	if !m.IsDismissed() {
		t.Error("expected dismissed after Esc")
	}

	view := m.View()
	if view != "" {
		t.Error("expected empty view when dismissed")
	}
}

func TestDismissWithEnter(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err, WithDismissible(true))

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m, _ = m.Update(keyMsg)

	if !m.IsDismissed() {
		t.Error("expected dismissed after Enter")
	}
}

func TestNotDismissibleByDefault(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err) // Not dismissible

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m, _ = m.Update(keyMsg)

	if m.IsDismissed() {
		t.Error("expected not dismissible by default")
	}
}

func TestSetters(t *testing.T) {
	t.Parallel()
	m := New(nil)

	newErr := errors.New("new error")
	m = m.SetError(newErr)
	if m.Message() != "new error" {
		t.Error("SetError failed")
	}

	m = m.SetMessage("direct")
	if m.Message() != "direct" {
		t.Error("SetMessage failed")
	}

	m = m.SetDetails("details")
	if m.details != "details" {
		t.Error("SetDetails failed")
	}

	m = m.SetMode(ModeBanner)
	if m.mode != ModeBanner {
		t.Error("SetMode failed")
	}

	m = m.SetSeverity(SeverityCritical)
	if m.severity != SeverityCritical {
		t.Error("SetSeverity failed")
	}

	m = m.SetWidth(200)
	if m.width != 200 {
		t.Error("SetWidth failed")
	}
}

func TestClear(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err, WithDetails("details"))

	m = m.Clear()

	if m.HasError() {
		t.Error("expected no error after Clear")
	}
	if m.Message() != "" {
		t.Error("expected empty message after Clear")
	}
	if m.details != "" {
		t.Error("expected empty details after Clear")
	}
}

func TestHasError(t *testing.T) {
	t.Parallel()
	m := New(nil)
	if m.HasError() {
		t.Error("expected no error for nil")
	}

	m = New(errors.New("test"))
	if !m.HasError() {
		t.Error("expected error")
	}

	m = m.Dismiss()
	if m.HasError() {
		t.Error("expected no error when dismissed")
	}
}

func TestEmptyView(t *testing.T) {
	t.Parallel()
	m := New(nil)
	view := m.View()
	if view != "" {
		t.Errorf("expected empty view for nil error, got %q", view)
	}
}

func TestInit(t *testing.T) {
	t.Parallel()
	m := New(errors.New("test"))
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are not empty
	if styles.Banner.String() == "" {
		t.Error("expected non-empty Banner style")
	}
	if styles.Inline.String() == "" {
		t.Error("expected non-empty Inline style")
	}
}

func TestWithHint(t *testing.T) {
	t.Parallel()
	err := errors.New("network error")
	m := New(err, WithHint("Check connectivity"))

	if m.Hint() != "Check connectivity" {
		t.Errorf("expected hint 'Check connectivity', got %q", m.Hint())
	}
	view := m.View()
	if !strings.Contains(view, "Check connectivity") {
		t.Error("expected hint in inline view")
	}
}

func TestHintInBanner(t *testing.T) {
	t.Parallel()
	m := New(errors.New("test"), WithMode(ModeBanner), WithHint("Helpful hint"))
	view := m.View()
	if !strings.Contains(view, "Helpful hint") {
		t.Error("expected hint in banner view")
	}
}

func TestSetHint(t *testing.T) {
	t.Parallel()
	m := New(errors.New("test"))
	m = m.SetHint("new hint")
	if m.Hint() != "new hint" {
		t.Errorf("SetHint failed, got %q", m.Hint())
	}
}

func TestClearResetsHint(t *testing.T) {
	t.Parallel()
	m := New(errors.New("test"), WithHint("some hint"))
	m = m.Clear()
	if m.Hint() != "" {
		t.Error("expected empty hint after Clear")
	}
}

func TestNewCategorized(t *testing.T) {
	t.Parallel()
	// Use a generic error — categorizer will produce "Error: ..." message
	err := errors.New("something broke")
	m := NewCategorized(err)

	if m.Message() == "" {
		t.Error("expected non-empty categorized message")
	}
	if m.Hint() == "" {
		t.Error("expected non-empty categorized hint")
	}
	if !errors.Is(m.Error(), err) {
		t.Error("expected original error preserved")
	}
}

func TestRenderInline(t *testing.T) {
	t.Parallel()

	t.Run("nil error returns empty", func(t *testing.T) {
		t.Parallel()
		if RenderInline(nil) != "" {
			t.Error("expected empty for nil error")
		}
	})

	t.Run("renders error with icon", func(t *testing.T) {
		t.Parallel()
		result := RenderInline(errors.New("test failure"))
		if !strings.Contains(result, "✗") {
			t.Error("expected error icon in render")
		}
		if result == "" {
			t.Error("expected non-empty render")
		}
	})
}

func TestDetailedViewWithDetails(t *testing.T) {
	t.Parallel()
	err := errors.New("main error")
	m := New(err,
		WithMode(ModeDetailed),
		WithDetails("Stack trace:\n  at line 42"),
	)

	view := m.View()
	if !strings.Contains(view, "main error") {
		t.Error("expected view to contain main error")
	}
	if !strings.Contains(view, "Stack trace") {
		t.Error("expected view to contain details")
	}
}

func TestDismissibleHint(t *testing.T) {
	t.Parallel()
	err := errors.New("test")
	m := New(err, WithMode(ModeBanner), WithDismissible(true))

	view := m.View()
	if !strings.Contains(view, "Esc") {
		t.Error("expected dismissible hint in banner")
	}
}

func TestGetIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityError, "✗"},
		{SeverityWarning, "⚠"},
		{SeverityCritical, "🔴"},
	}

	for _, test := range tests {
		m := New(errors.New("test"), WithSeverity(test.severity))
		icon := m.getIcon()
		if icon != test.expected {
			t.Errorf("expected icon %q for severity %d, got %q", test.expected, test.severity, icon)
		}
	}
}
