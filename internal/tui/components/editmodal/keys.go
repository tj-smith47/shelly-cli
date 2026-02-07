package editmodal

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
)

// KeyAction represents the result of translating a key press.
type KeyAction int

const (
	// ActionNone means the key was not recognized as a modal action.
	ActionNone KeyAction = iota
	// ActionClose means the modal should close (Esc, Ctrl+[).
	ActionClose
	// ActionSave means the modal should save (Enter or Ctrl+S).
	ActionSave
	// ActionNext means move to the next field (Tab).
	ActionNext
	// ActionPrev means move to the previous field (Shift+Tab).
	ActionPrev
	// ActionNavUp means navigate up (from NavigationMsg).
	ActionNavUp
	// ActionNavDown means navigate down (from NavigationMsg).
	ActionNavDown
)

// HandleKey translates common edit modal keys into actions.
// Returns ActionNone for unrecognized keys, allowing the modal to handle them
// or delegate to the focused input. All actions are suppressed when Saving is true.
func (b *Base) HandleKey(msg tea.KeyPressMsg) KeyAction {
	if b.Saving {
		return ActionNone
	}

	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		return ActionClose
	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		return ActionSave
	case keyconst.KeyTab:
		return ActionNext
	case keyconst.KeyShiftTab:
		return ActionPrev
	default:
		return ActionNone
	}
}
