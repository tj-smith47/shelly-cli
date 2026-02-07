package editmodal

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestBase_HandleKey_Actions(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3}

	tests := []struct {
		key  string
		want KeyAction
	}{
		{"esc", ActionClose},
		{"ctrl+[", ActionClose},
		{"enter", ActionSave},
		{"ctrl+s", ActionSave},
		{"tab", ActionNext},
		{"shift+tab", ActionPrev},
		{"a", ActionNone},
		{"x", ActionNone},
	}

	for _, tt := range tests {
		msg := tea.KeyPressMsg{}
		// Build a KeyPressMsg that has the right String() return
		// We need to use the actual bubbletea key construction
		switch tt.key {
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "ctrl+[":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape} // ctrl+[ maps to esc
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "ctrl+s":
			msg = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
		case "tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
		case "a":
			msg = tea.KeyPressMsg{Code: 'a', Text: "a"}
		case "x":
			msg = tea.KeyPressMsg{Code: 'x', Text: "x"}
		}
		got := b.HandleKey(msg)
		if got != tt.want {
			t.Errorf("HandleKey(%q) = %d, want %d (key.String()=%q)", tt.key, got, tt.want, msg.String())
		}
	}
}

func TestBase_HandleKey_SuppressedWhenSaving(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3, Saving: true}

	// All keys should return ActionNone when saving
	keys := []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
		{Code: tea.KeyEnter},
		{Code: tea.KeyTab},
		{Code: tea.KeyTab, Mod: tea.ModShift},
	}

	for _, msg := range keys {
		got := b.HandleKey(msg)
		if got != ActionNone {
			t.Errorf("HandleKey(%q) while saving = %d, want ActionNone", msg.String(), got)
		}
	}
}
