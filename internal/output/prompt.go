// Package output provides pure formatters (data → string).
package output

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// FormatREPLPrompt creates the REPL prompt string.
func FormatREPLPrompt(activeDevice string) string {
	if activeDevice != "" {
		return fmt.Sprintf("shelly [%s]> ", theme.Highlight().Render(activeDevice))
	}
	return "shelly> "
}

// FormatShellPrompt creates the shell prompt string for a device.
func FormatShellPrompt(device string) string {
	return fmt.Sprintf("%s> ", theme.Highlight().Render(device))
}
