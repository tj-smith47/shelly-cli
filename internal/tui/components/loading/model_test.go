package loading

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if !m.visible {
		t.Error("expected loading to be visible by default")
	}
	if m.message != "Loading..." {
		t.Errorf("expected default message 'Loading...', got %q", m.message)
	}
	if !m.centerH || !m.centerV {
		t.Error("expected centering to be enabled by default")
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Parallel()
	m := New(
		WithMessage("Fetching data..."),
		WithStyle(StyleLine),
		WithCentered(false, true),
	)

	if m.message != "Fetching data..." {
		t.Errorf("expected message 'Fetching data...', got %q", m.message)
	}
	if m.centerH {
		t.Error("expected horizontal centering to be disabled")
	}
	if !m.centerV {
		t.Error("expected vertical centering to be enabled")
	}
}

func TestSetMessage(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetMessage("New message")

	if m.message != "New message" {
		t.Errorf("expected message 'New message', got %q", m.message)
	}
	if m.Message() != "New message" {
		t.Errorf("expected Message() to return 'New message', got %q", m.Message())
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

func TestVisibility(t *testing.T) {
	t.Parallel()
	m := New()

	if !m.IsVisible() {
		t.Error("expected loading to be visible by default")
	}

	m = m.SetVisible(false)
	if m.IsVisible() {
		t.Error("expected loading to be hidden after SetVisible(false)")
	}

	m = m.Hide()
	if m.IsVisible() {
		t.Error("expected loading to be hidden after Hide()")
	}

	m, _ = m.Show()
	if !m.IsVisible() {
		t.Error("expected loading to be visible after Show()")
	}
}

func TestViewWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetVisible(false)

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when hidden, got %q", view)
	}
}

func TestViewWhenVisible(t *testing.T) {
	t.Parallel()
	m := New(WithMessage("Test message"))
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when visible")
	}
	if !strings.Contains(view, "Test message") {
		t.Error("expected view to contain the message")
	}
}

func TestInit(t *testing.T) {
	t.Parallel()
	m := New()
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected Init to return a tick command")
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	m := New()

	// Test that spinner tick is handled
	tickMsg := spinner.TickMsg{}
	newM, cmd := m.Update(tickMsg)

	// Should return the model and possibly another tick command
	if newM.message != m.message {
		t.Error("expected message to remain unchanged")
	}
	// cmd may be nil or a tick command depending on spinner state
	_ = cmd
}

func TestUpdateWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetVisible(false)

	tickMsg := spinner.TickMsg{}
	newM, cmd := m.Update(tickMsg)

	if cmd != nil {
		t.Error("expected no command when hidden")
	}
	if newM.visible {
		t.Error("expected model to remain hidden")
	}
}

func TestSetSpinnerStyle(t *testing.T) {
	t.Parallel()
	m := New()

	// Change to line style
	m = m.SetSpinnerStyle(StyleLine)

	// Verify the spinner was updated (we can't directly compare spinners,
	// but we can verify it doesn't panic and accepts valid styles)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after style change")
	}
}

func TestAllSpinnerStyles(t *testing.T) {
	t.Parallel()
	styles := []Style{
		StyleDot,
		StyleLine,
		StyleMiniDot,
		StylePulse,
		StylePoints,
		StyleGlobe,
		StyleMoon,
		StyleMonkey,
	}

	for _, style := range styles {
		m := New(WithStyle(style))
		view := m.View()
		if view == "" {
			t.Errorf("expected non-empty view for style %d", style)
		}
	}
}

func TestSpinnersMap(t *testing.T) {
	t.Parallel()
	// Verify all defined styles have corresponding spinners
	expectedStyles := []Style{
		StyleDot,
		StyleLine,
		StyleMiniDot,
		StylePulse,
		StylePoints,
		StyleGlobe,
		StyleMoon,
		StyleMonkey,
	}

	for _, style := range expectedStyles {
		s, ok := Spinners[style]
		if !ok {
			t.Errorf("missing spinner for style %d", style)
			continue
		}
		if len(s.Frames) == 0 {
			t.Errorf("spinner for style %d has no frames", style)
		}
		if s.FPS == 0 {
			t.Errorf("spinner for style %d has no FPS", style)
		}
	}
}

func TestTick(t *testing.T) {
	t.Parallel()
	m := New()
	cmd := m.Tick()

	if cmd == nil {
		t.Error("expected Tick to return a command")
	}
}

func TestShowReturnsTickCommand(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetVisible(false)

	newM, cmd := m.Show()

	if !newM.IsVisible() {
		t.Error("expected Show to make model visible")
	}
	if cmd == nil {
		t.Error("expected Show to return a tick command")
	}
}

func TestKeyPressDoesNotAffectSpinner(t *testing.T) {
	t.Parallel()
	m := New()

	// Key presses should not affect the spinner
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newM, _ := m.Update(keyMsg)

	if newM.message != m.message {
		t.Error("expected key press to not affect spinner state")
	}
}
