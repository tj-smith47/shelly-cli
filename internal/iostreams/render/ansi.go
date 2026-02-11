// Package render provides terminal rendering primitives for multi-line progress output.
// Adapted from mc-cli's Docker BuildKit-style render engine.
package render

import (
	"fmt"
	"unicode/utf8"
)

// ANSI escape sequence helpers for terminal cursor manipulation.

const (
	esc = "\x1b["

	// Common sequences.
	clearLineSeq    = esc + "2K"
	clearScreenDown = esc + "0J"
	hideCursorSeq   = esc + "?25l"
	showCursorSeq   = esc + "?25h"
)

// MoveUp returns an ANSI escape sequence to move the cursor up n lines.
func MoveUp(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%dA", esc, n)
}

// MoveDown returns an ANSI escape sequence to move the cursor down n lines.
func MoveDown(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%dB", esc, n)
}

// ClearLine returns an ANSI escape sequence to clear the entire current line.
func ClearLine() string {
	return clearLineSeq
}

// ClearDown returns an ANSI escape sequence to clear from cursor to end of screen.
func ClearDown() string {
	return clearScreenDown
}

// HideCursor returns an ANSI escape sequence to hide the cursor.
func HideCursor() string {
	return hideCursorSeq
}

// ShowCursor returns an ANSI escape sequence to show the cursor.
func ShowCursor() string {
	return showCursorSeq
}

// skipCSI advances past an ANSI CSI sequence starting at position i in s.
// Returns the new position after the sequence. Assumes s[i] == '\x1b' and s[i+1] == '['.
func skipCSI(s string, i int) int {
	i += 2 // skip ESC [
	for i < len(s) {
		if s[i] >= 0x40 && s[i] <= 0x7E {
			i++
			break
		}
		i++
	}
	return i
}

// isCSIStart returns true if position i in s begins an ANSI CSI sequence (ESC [).
func isCSIStart(s string, i int) bool {
	return s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '['
}

// hasVisibleAfter returns true if s contains visible (non-ANSI) characters from position i onward.
func hasVisibleAfter(s string, i int) bool {
	for j := i; j < len(s); {
		if isCSIStart(s, j) {
			j = skipCSI(s, j)
			continue
		}
		return true
	}
	return false
}

// VisibleLen returns the number of visible characters (runes) in s,
// skipping ANSI escape sequences.
func VisibleLen(s string) int {
	visible := 0
	for i := 0; i < len(s); {
		if isCSIStart(s, i) {
			i = skipCSI(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		visible++
		i += size
	}
	return visible
}

// TruncateLine truncates a line to maxWidth visible characters (runes),
// preserving ANSI escape sequences. Replaces the last character with "…" if truncated.
func TruncateLine(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	visible := 0
	var result []byte

	for i := 0; i < len(s); {
		if isCSIStart(s, i) {
			end := skipCSI(s, i)
			result = append(result, s[i:end]...)
			i = end
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		visible++
		result = append(result, s[i:i+size]...)
		i += size

		if visible >= maxWidth && hasVisibleAfter(s, i) {
			// Replace last visible rune with ellipsis
			result = result[:len(result)-utf8.RuneLen(r)]
			result = append(result, "\u2026"...)
			break
		}
	}

	return string(result)
}
