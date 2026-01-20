package keys

import (
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// ScrollNavigator defines the interface for scrollable components.
// This matches the *panel.Scroller methods used for list navigation.
type ScrollNavigator interface {
	CursorUp() bool
	CursorDown() bool
	CursorToStart()
	CursorToEnd()
	PageUp()
	PageDown()
}

// HandleScrollNavigation handles common navigation keys for scrollable lists.
// Returns true if the key was handled, false otherwise.
//
// Supported keys:
//   - j, down: Move cursor down
//   - k, up: Move cursor up
//   - g: Jump to start
//   - G: Jump to end
//   - ctrl+d, pgdown: Page down
//   - ctrl+u, pgup: Page up
//
// Example usage:
//
//	func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
//	    if keys.HandleScrollNavigation(msg.String(), m.scroller) {
//	        return m, nil
//	    }
//	    // handle other keys
//	}
func HandleScrollNavigation(key string, nav ScrollNavigator) bool {
	switch key {
	case "j", "down":
		nav.CursorDown()
	case "k", "up":
		nav.CursorUp()
	case "g":
		nav.CursorToStart()
	case "G":
		nav.CursorToEnd()
	case keyconst.KeyCtrlD, keyconst.KeyPgDown:
		nav.PageDown()
	case keyconst.KeyCtrlU, keyconst.KeyPgUp:
		nav.PageUp()
	default:
		return false
	}
	return true
}

// HandleScrollNavigationScroller is a convenience wrapper for *panel.Scroller.
// It's equivalent to HandleScrollNavigation but with a concrete type.
func HandleScrollNavigationScroller(key string, scroller *panel.Scroller) bool {
	if scroller == nil {
		return false
	}
	return HandleScrollNavigation(key, scroller)
}
