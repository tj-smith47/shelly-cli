package keys

import "strings"

// Hint represents a keybinding hint for display in panel footers.
type Hint struct {
	Key  string // e.g., "e", "j/k", "spc"
	Desc string // e.g., "edit", "nav", "sel"
}

// FormatHints formats a list of hints as a compact string that fits within maxWidth.
// Drops hints from the end when they won't fit, ensuring whole hints are preserved.
// Returns a plain string like "e:edit n:new d:del" — caller should style with theme.StyledKeybindings().
func FormatHints(hints []Hint, maxWidth int) string {
	if len(hints) == 0 || maxWidth < 3 {
		return ""
	}

	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = h.Key + ":" + h.Desc
	}

	// Try fitting all hints, then progressively fewer
	for n := len(parts); n > 0; n-- {
		result := strings.Join(parts[:n], " ")
		if len(result) <= maxWidth {
			return result
		}
	}

	// Even the first hint doesn't fit; return it anyway (renderer will truncate)
	return parts[0]
}

// FooterHintWidth returns the max content width for hint text in a panel footer.
// Accounts for border overhead (├─ ... ─┤) and the panel index hint (⇧N) section.
func FooterHintWidth(panelWidth int) int {
	// Border: ╰─├─ hint ─┤───├─ ⇧N ─┤╯ uses ~18 chars overhead
	w := panelWidth - 18
	if w < 10 {
		return 10
	}
	return w
}
