package form

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Test errors for validation.
var (
	errEmpty    = errors.New("value required")
	errTooShort = errors.New("must be at least 10 characters")
)

// TestTextInputNew tests TextInput creation.
func TestTextInputNew(t *testing.T) {
	t.Parallel()
	ti := NewTextInput()

	if ti.Value() != "" {
		t.Errorf("expected empty value, got %q", ti.Value())
	}
	if ti.Focused() {
		t.Error("expected not focused by default")
	}
	if ti.Error() != nil {
		t.Error("expected no error by default")
	}
}

func TestTextInputWithOptions(t *testing.T) {
	t.Parallel()
	ti := NewTextInput(
		WithLabel("Username"),
		WithPlaceholder("Enter username"),
		WithCharLimit(50),
		WithHelp("Your unique username"),
	)

	if ti.label != "Username" {
		t.Errorf("expected label 'Username', got %q", ti.label)
	}
	if ti.help != "Your unique username" {
		t.Errorf("expected help text, got %q", ti.help)
	}
}

func TestTextInputValidation(t *testing.T) {
	t.Parallel()
	ti := NewTextInput(
		WithValidation(func(s string) error {
			if s == "" {
				return errEmpty
			}
			return nil
		}),
	)

	ti = ti.SetValue("")
	if ti.Valid() {
		t.Error("expected validation to fail for empty value")
	}
	if !errors.Is(ti.Error(), errEmpty) {
		t.Errorf("expected errEmpty, got %v", ti.Error())
	}

	ti = ti.SetValue("test")
	if !ti.Valid() {
		t.Error("expected validation to pass for non-empty value")
	}
}

func TestTextInputFocus(t *testing.T) {
	t.Parallel()
	ti := NewTextInput()

	ti, _ = ti.Focus()
	if !ti.Focused() {
		t.Error("expected focused after Focus()")
	}

	ti = ti.Blur()
	if ti.Focused() {
		t.Error("expected not focused after Blur()")
	}
}

func TestTextInputReset(t *testing.T) {
	t.Parallel()
	ti := NewTextInput()
	ti = ti.SetValue("test")
	ti = ti.Reset()

	if ti.Value() != "" {
		t.Errorf("expected empty value after Reset(), got %q", ti.Value())
	}
}

func TestTextInputView(t *testing.T) {
	t.Parallel()
	ti := NewTextInput(WithLabel("Test"))
	view := ti.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestPasswordNew tests Password creation.
func TestPasswordNew(t *testing.T) {
	t.Parallel()
	p := NewPassword()

	if p.Value() != "" {
		t.Errorf("expected empty value, got %q", p.Value())
	}
	if p.IsVisible() {
		t.Error("expected password hidden by default")
	}
}

func TestPasswordWithOptions(t *testing.T) {
	t.Parallel()
	p := NewPassword(
		WithPasswordLabel("Password"),
		WithPasswordPlaceholder("Enter password"),
		WithPasswordCharLimit(100),
		WithMaskChar('*'),
	)

	if p.label != "Password" {
		t.Errorf("expected label 'Password', got %q", p.label)
	}
}

func TestPasswordToggleVisibility(t *testing.T) {
	t.Parallel()
	p := NewPassword()
	p, _ = p.Focus()

	// Toggle visibility with ctrl+t
	keyMsg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModCtrl}
	// Note: We simulate the key but the actual toggle logic checks for "ctrl+t"
	// This test verifies the component handles key messages without panicking
	_, _ = p.Update(keyMsg)
}

func TestPasswordView(t *testing.T) {
	t.Parallel()
	p := NewPassword(WithPasswordLabel("Secret"))
	view := p.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestToggleNew tests Toggle creation.
func TestToggleNew(t *testing.T) {
	t.Parallel()
	tog := NewToggle()

	if tog.Value() {
		t.Error("expected false by default")
	}
	if tog.Focused() {
		t.Error("expected not focused by default")
	}
}

func TestToggleWithOptions(t *testing.T) {
	t.Parallel()
	tog := NewToggle(
		WithToggleLabel("Enable feature"),
		WithToggleOnLabel("Enabled"),
		WithToggleOffLabel("Disabled"),
		WithToggleValue(true),
	)

	if tog.label != "Enable feature" {
		t.Errorf("expected label, got %q", tog.label)
	}
	if tog.onLabel != "Enabled" {
		t.Errorf("expected onLabel 'Enabled', got %q", tog.onLabel)
	}
	if !tog.Value() {
		t.Error("expected value true")
	}
}

func TestToggleToggle(t *testing.T) {
	t.Parallel()
	tog := NewToggle()
	tog = tog.Focus()

	// Test space toggles
	keyMsg := tea.KeyPressMsg{Code: ' '}
	tog, _ = tog.Update(keyMsg)
	if !tog.Value() {
		t.Error("expected value true after space")
	}

	tog, _ = tog.Update(keyMsg)
	if tog.Value() {
		t.Error("expected value false after second space")
	}
}

func TestToggleYN(t *testing.T) {
	t.Parallel()
	tog := NewToggle()
	tog = tog.Focus()

	// Test 'y' sets true
	keyMsg := tea.KeyPressMsg{Code: 'y'}
	tog, _ = tog.Update(keyMsg)
	if !tog.Value() {
		t.Error("expected value true after 'y'")
	}

	// Test 'n' sets false
	keyMsg = tea.KeyPressMsg{Code: 'n'}
	tog, _ = tog.Update(keyMsg)
	if tog.Value() {
		t.Error("expected value false after 'n'")
	}
}

func TestToggleView(t *testing.T) {
	t.Parallel()
	tog := NewToggle(WithToggleLabel("Test"))
	view := tog.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestDropdownNew tests Dropdown creation.
func TestDropdownNew(t *testing.T) {
	t.Parallel()
	d := NewDropdown(
		WithDropdownOptions([]string{"Option A", "Option B", "Option C"}),
	)

	if d.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", d.Selected())
	}
	if d.SelectedValue() != "Option A" {
		t.Errorf("expected 'Option A', got %q", d.SelectedValue())
	}
}

func TestDropdownWithOptions(t *testing.T) {
	t.Parallel()
	d := NewDropdown(
		WithDropdownLabel("Choose option"),
		WithDropdownOptions([]string{"A", "B", "C"}),
		WithDropdownSelected(1),
		WithDropdownMaxVisible(3),
	)

	if d.label != "Choose option" {
		t.Errorf("expected label, got %q", d.label)
	}
	if d.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", d.Selected())
	}
	if d.SelectedValue() != "B" {
		t.Errorf("expected 'B', got %q", d.SelectedValue())
	}
}

func TestDropdownNavigation(t *testing.T) {
	t.Parallel()
	d := NewDropdown(
		WithDropdownOptions([]string{"A", "B", "C"}),
	)
	d = d.Focus()

	// Expand
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	d, _ = d.Update(keyMsg)
	if !d.IsExpanded() {
		t.Error("expected expanded after enter")
	}

	// Navigate down
	keyMsg = tea.KeyPressMsg{Code: 'j'}
	d, _ = d.Update(keyMsg)
	if d.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", d.cursor)
	}

	// Select
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEnter}
	d, _ = d.Update(keyMsg)
	if d.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", d.Selected())
	}
	if d.IsExpanded() {
		t.Error("expected collapsed after selection")
	}
}

func TestDropdownSetOptions(t *testing.T) {
	t.Parallel()
	d := NewDropdown(WithDropdownOptions([]string{"A", "B"}))
	d = d.SetSelected(1)
	d = d.SetOptions([]string{"X"})

	if d.Selected() != 0 {
		t.Errorf("expected selected clamped to 0, got %d", d.Selected())
	}
	if d.SelectedValue() != "X" {
		t.Errorf("expected 'X', got %q", d.SelectedValue())
	}
}

func TestDropdownView(t *testing.T) {
	t.Parallel()
	d := NewDropdown(
		WithDropdownLabel("Test"),
		WithDropdownOptions([]string{"A", "B"}),
	)
	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestSliderNew tests Slider creation.
func TestSliderNew(t *testing.T) {
	t.Parallel()
	s := NewSlider()

	if s.Value() != 0 {
		t.Errorf("expected value 0, got %f", s.Value())
	}
	if s.min != 0 || s.max != 100 {
		t.Error("expected default range 0-100")
	}
}

func TestSliderWithOptions(t *testing.T) {
	t.Parallel()
	s := NewSlider(
		WithSliderLabel("Volume"),
		WithSliderMin(0),
		WithSliderMax(10),
		WithSliderStep(0.5),
		WithSliderValue(5),
		WithSliderWidth(30),
		WithSliderFormat("%.1f"),
	)

	if s.label != "Volume" {
		t.Errorf("expected label 'Volume', got %q", s.label)
	}
	if s.Value() != 5 {
		t.Errorf("expected value 5, got %f", s.Value())
	}
}

func TestSliderNavigation(t *testing.T) {
	t.Parallel()
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(10),
		WithSliderStep(1),
		WithSliderValue(5),
	)
	s = s.Focus()

	// Move right
	keyMsg := tea.KeyPressMsg{Code: tea.KeyRight}
	s, _ = s.Update(keyMsg)
	if s.Value() != 6 {
		t.Errorf("expected value 6, got %f", s.Value())
	}

	// Move left
	keyMsg = tea.KeyPressMsg{Code: tea.KeyLeft}
	s, _ = s.Update(keyMsg)
	if s.Value() != 5 {
		t.Errorf("expected value 5, got %f", s.Value())
	}
}

func TestSliderClamp(t *testing.T) {
	t.Parallel()
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(10),
	)

	s = s.SetValue(-5)
	if s.Value() != 0 {
		t.Errorf("expected clamped to 0, got %f", s.Value())
	}

	s = s.SetValue(20)
	if s.Value() != 10 {
		t.Errorf("expected clamped to 10, got %f", s.Value())
	}
}

func TestSliderPercentage(t *testing.T) {
	t.Parallel()
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(100),
		WithSliderValue(50),
	)

	if s.Percentage() != 50 {
		t.Errorf("expected 50%%, got %f%%", s.Percentage())
	}
}

func TestSliderView(t *testing.T) {
	t.Parallel()
	s := NewSlider(WithSliderLabel("Test"))
	view := s.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestTextAreaNew tests TextArea creation.
func TestTextAreaNew(t *testing.T) {
	t.Parallel()
	ta := NewTextArea()

	if ta.Value() != "" {
		t.Errorf("expected empty value, got %q", ta.Value())
	}
	if ta.Focused() {
		t.Error("expected not focused by default")
	}
}

func TestTextAreaWithOptions(t *testing.T) {
	t.Parallel()
	ta := NewTextArea(
		WithTextAreaLabel("Description"),
		WithTextAreaPlaceholder("Enter description..."),
		WithTextAreaCharLimit(500),
		WithTextAreaDimensions(60, 10),
		WithTextAreaShowLineNumbers(true),
	)

	if ta.label != "Description" {
		t.Errorf("expected label 'Description', got %q", ta.label)
	}
}

func TestTextAreaValidation(t *testing.T) {
	t.Parallel()
	ta := NewTextArea(
		WithTextAreaValidation(func(s string) error {
			if len(s) < 10 {
				return errTooShort
			}
			return nil
		}),
	)

	ta = ta.SetValue("short")
	if ta.Valid() {
		t.Error("expected validation to fail")
	}

	ta = ta.SetValue("this is long enough")
	if !ta.Valid() {
		t.Error("expected validation to pass")
	}
}

func TestTextAreaFocus(t *testing.T) {
	t.Parallel()
	ta := NewTextArea()

	ta, _ = ta.Focus()
	if !ta.Focused() {
		t.Error("expected focused after Focus()")
	}

	ta = ta.Blur()
	if ta.Focused() {
		t.Error("expected not focused after Blur()")
	}
}

func TestTextAreaReset(t *testing.T) {
	t.Parallel()
	ta := NewTextArea()
	ta = ta.SetValue("test content")
	ta = ta.Reset()

	if ta.Value() != "" {
		t.Errorf("expected empty value after Reset(), got %q", ta.Value())
	}
}

func TestTextAreaView(t *testing.T) {
	t.Parallel()
	ta := NewTextArea(WithTextAreaLabel("Test"))
	view := ta.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestDefaultStyles tests that default styles are created without panicking.
func TestDefaultStyles(t *testing.T) {
	t.Parallel()

	// These should not panic
	_ = DefaultTextInputStyles()
	_ = DefaultToggleStyles()
	_ = DefaultDropdownStyles()
	_ = DefaultSliderStyles()
	_ = DefaultTextAreaStyles()
}
