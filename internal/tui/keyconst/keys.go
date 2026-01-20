// Package keyconst provides shared key binding constants for TUI components.
package keyconst

// Panel focus key constants for Shift+N shortcuts.
// On US keyboard layout, Shift+Number produces these characters:
// Shift+1=!, Shift+2=@, Shift+3=#, Shift+4=$, Shift+5=%.
// Shift+6=^, Shift+7=&, Shift+8=*, Shift+9=(.
const (
	Shift1 = "!"
	Shift2 = "@"
	Shift3 = "#"
	Shift4 = "$"
	Shift5 = "%"
	Shift6 = "^"
	Shift7 = "&"
	Shift8 = "*"
	Shift9 = "("
)

// Common key constants.
const (
	KeyEnter    = "enter"
	KeySpace    = "space"
	KeyTab      = "tab"
	KeyShiftTab = "shift+tab"
	KeyEsc      = "esc"
)

// Navigation key constants.
const (
	KeyDown   = "down"
	KeyUp     = "up"
	KeyPgDown = "pgdown"
	KeyPgUp   = "pgup"
	KeyCtrlD  = "ctrl+d"
	KeyCtrlU  = "ctrl+u"
)

// Action key constants.
const (
	KeyCtrlS = "ctrl+s"
	KeyCtrlF = "ctrl+f"
)
