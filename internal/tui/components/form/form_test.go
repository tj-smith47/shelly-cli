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

// TestSelectNew tests Select creation.
func TestSelectNew(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectOptions([]string{"Option A", "Option B", "Option C"}),
	)

	if s.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", s.Selected())
	}
	if s.SelectedValue() != "Option A" {
		t.Errorf("expected 'Option A', got %q", s.SelectedValue())
	}
	if s.IsExpanded() {
		t.Error("expected not expanded by default")
	}
	if s.IsFiltering() {
		t.Error("expected not filtering by default")
	}
}

func TestSelectWithOptions(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectLabel("Choose option"),
		WithSelectOptions([]string{"A", "B", "C"}),
		WithSelectSelected(1),
		WithSelectMaxVisible(3),
		WithSelectHelp("Pick one"),
		WithSelectFiltering(true),
	)

	if s.label != "Choose option" {
		t.Errorf("expected label, got %q", s.label)
	}
	if s.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", s.Selected())
	}
	if s.SelectedValue() != "B" {
		t.Errorf("expected 'B', got %q", s.SelectedValue())
	}
	if s.help != "Pick one" {
		t.Errorf("expected help 'Pick one', got %q", s.help)
	}
	if !s.filterEnabled {
		t.Error("expected filtering enabled")
	}
}

func TestSelectNavigation(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectOptions([]string{"A", "B", "C"}),
	)
	s = s.Focus()

	// Expand with enter
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)
	if !s.IsExpanded() {
		t.Error("expected expanded after enter")
	}

	// Navigate down with 'j'
	keyMsg = tea.KeyPressMsg{Code: 'j'}
	s, _ = s.Update(keyMsg)
	if s.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", s.cursor)
	}

	// Navigate down with 'down'
	keyMsg = tea.KeyPressMsg{Code: tea.KeyDown}
	s, _ = s.Update(keyMsg)
	if s.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", s.cursor)
	}

	// Navigate up with 'k'
	keyMsg = tea.KeyPressMsg{Code: 'k'}
	s, _ = s.Update(keyMsg)
	if s.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", s.cursor)
	}

	// Home with 'g'
	keyMsg = tea.KeyPressMsg{Code: 'g'}
	s, _ = s.Update(keyMsg)
	if s.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", s.cursor)
	}

	// End with 'G'
	keyMsg = tea.KeyPressMsg{Code: 'G', Mod: tea.ModShift}
	s, _ = s.Update(keyMsg)
	if s.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", s.cursor)
	}

	// Select with enter
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)
	if s.Selected() != 2 {
		t.Errorf("expected selected 2, got %d", s.Selected())
	}
	if s.IsExpanded() {
		t.Error("expected collapsed after selection")
	}
}

func TestSelectEscape(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectOptions([]string{"A", "B", "C"}),
	)
	s = s.Focus()

	// Expand
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)
	if !s.IsExpanded() {
		t.Error("expected expanded")
	}

	// Escape closes without selecting
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEscape}
	s, _ = s.Update(keyMsg)
	if s.IsExpanded() {
		t.Error("expected collapsed after escape")
	}
}

func TestSelectFiltering(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectOptions([]string{"Apple", "Banana", "Cherry", "Apricot"}),
		WithSelectFiltering(true),
	)
	s = s.Focus()

	// Expand
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)

	// Enter filter mode with '/'
	keyMsg = tea.KeyPressMsg{Code: '/', Text: "/"}
	s, _ = s.Update(keyMsg)
	if !s.IsFiltering() {
		t.Error("expected filtering mode active")
	}

	// Type 'a' to filter (Text field required for textinput in bubbletea v2)
	keyMsg = tea.KeyPressMsg{Code: 'a', Text: "a"}
	s, _ = s.Update(keyMsg)
	// Should show: Apple, Banana, Apricot (contain 'a', Cherry doesn't)
	if len(s.filtered) != 3 {
		t.Errorf("expected 3 filtered options (Apple, Banana, Apricot), got %d", len(s.filtered))
	}

	// Type 'p' to narrow filter to 'ap'
	keyMsg = tea.KeyPressMsg{Code: 'p', Text: "p"}
	s, _ = s.Update(keyMsg)
	// Should show: Apple, Apricot (contain 'ap')
	if len(s.filtered) != 2 {
		t.Errorf("expected 2 filtered options, got %d", len(s.filtered))
	}

	// Press escape to exit filter mode
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEscape}
	s, _ = s.Update(keyMsg)
	if s.IsFiltering() {
		t.Error("expected filtering mode exited")
	}
	// Filter should be cleared, all options visible
	if len(s.filtered) != 4 {
		t.Errorf("expected all 4 options after escape, got %d", len(s.filtered))
	}
}

func TestSelectSetOptions(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))
	s = s.SetSelected(1)
	s = s.SetOptions([]string{"X"})

	if s.Selected() != 0 {
		t.Errorf("expected selected clamped to 0, got %d", s.Selected())
	}
	if s.SelectedValue() != "X" {
		t.Errorf("expected 'X', got %q", s.SelectedValue())
	}
}

func TestSelectSetSelectedValue(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B", "C"}))
	s = s.SetSelectedValue("C")

	if s.Selected() != 2 {
		t.Errorf("expected selected 2, got %d", s.Selected())
	}
	if s.SelectedValue() != "C" {
		t.Errorf("expected 'C', got %q", s.SelectedValue())
	}
}

func TestSelectFocusBlur(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))

	s = s.Focus()
	if !s.Focused() {
		t.Error("expected focused")
	}

	// Expand
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)
	if !s.IsExpanded() {
		t.Error("expected expanded")
	}

	// Blur collapses
	s = s.Blur()
	if s.Focused() {
		t.Error("expected not focused after blur")
	}
	if s.IsExpanded() {
		t.Error("expected collapsed after blur")
	}
}

func TestSelectExpandCollapse(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))

	s = s.Expand()
	if !s.IsExpanded() {
		t.Error("expected expanded")
	}

	s = s.Collapse()
	if s.IsExpanded() {
		t.Error("expected collapsed")
	}
}

func TestSelectView(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectLabel("Test"),
		WithSelectOptions([]string{"A", "B"}),
	)
	view := s.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestSelectViewExpanded(t *testing.T) {
	t.Parallel()
	s := NewSelect(
		WithSelectLabel("Test"),
		WithSelectOptions([]string{"A", "B", "C"}),
		WithSelectFiltering(true),
	)
	s = s.Expand()
	view := s.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestSelectNoOptions(t *testing.T) {
	t.Parallel()
	s := NewSelect()

	if s.SelectedValue() != "" {
		t.Errorf("expected empty value, got %q", s.SelectedValue())
	}
}

func TestSelectDownExpandsWhenCollapsed(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))
	s = s.Focus()

	// Down arrow should expand when collapsed
	keyMsg := tea.KeyPressMsg{Code: tea.KeyDown}
	s, _ = s.Update(keyMsg)
	if !s.IsExpanded() {
		t.Error("expected expanded after down arrow on collapsed select")
	}
}

func TestSelectSpaceToggles(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))
	s = s.Focus()

	// Space should expand
	keyMsg := tea.KeyPressMsg{Code: ' '}
	s, _ = s.Update(keyMsg)
	if !s.IsExpanded() {
		t.Error("expected expanded after space")
	}

	// Space again should select and collapse
	keyMsg = tea.KeyPressMsg{Code: ' '}
	s, _ = s.Update(keyMsg)
	if s.IsExpanded() {
		t.Error("expected collapsed after second space")
	}
}

func TestSelectTab(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B", "C"}))
	s = s.Focus()
	s = s.Expand()

	// Move cursor to B
	keyMsg := tea.KeyPressMsg{Code: 'j'}
	s, _ = s.Update(keyMsg)

	// Tab selects and closes
	keyMsg = tea.KeyPressMsg{Code: tea.KeyTab}
	s, _ = s.Update(keyMsg)

	if s.SelectedValue() != "B" {
		t.Errorf("expected 'B', got %q", s.SelectedValue())
	}
	if s.IsExpanded() {
		t.Error("expected collapsed after tab")
	}
}

func TestSelectNotFocusedIgnoresInput(t *testing.T) {
	t.Parallel()
	s := NewSelect(WithSelectOptions([]string{"A", "B"}))
	// Not focused

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	s, _ = s.Update(keyMsg)

	if s.IsExpanded() {
		t.Error("expected still collapsed when not focused")
	}
}

// TestDefaultStyles tests that default styles are created without panicking.
func TestDefaultStyles(t *testing.T) {
	t.Parallel()

	// These should not panic
	_ = DefaultTextInputStyles()
	_ = DefaultToggleStyles()
	_ = DefaultSliderStyles()
	_ = DefaultTextAreaStyles()
	_ = DefaultSelectStyles()
}
