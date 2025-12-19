// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// printTable prints a table to ios.Out with standard error handling.
func printTable(ios *iostreams.IOStreams, table *output.Table) {
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// formatTemp formats a temperature value with color based on severity.
func formatTemp(c float64) string {
	s := fmt.Sprintf("%.1f°C", c)
	if c >= 70 {
		return theme.StatusError().Render(s)
	} else if c >= 50 {
		return theme.StatusWarn().Render(s)
	}
	return theme.StatusOK().Render(s)
}

// joinStrings joins string parts with a separator.
func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// repeatChar repeats a rune n times and returns the resulting string.
func repeatChar(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}

// valueOrEmpty returns a placeholder if the string is empty.
func valueOrEmpty(s string) string {
	if s == "" {
		return "<not connected>"
	}
	return s
}
