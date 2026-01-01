// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

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

// DisplayIntegratorCredentialHelp prints help for configuring integrator credentials.
func DisplayIntegratorCredentialHelp(ios *iostreams.IOStreams) {
	ios.Warning("Integrator credentials not configured")
	ios.Println()
	ios.Info("Set the following environment variables:")
	ios.Printf("  export SHELLY_INTEGRATOR_TAG=your-integrator-tag\n")
	ios.Printf("  export SHELLY_INTEGRATOR_TOKEN=your-integrator-token\n")
	ios.Println()
	ios.Info("Or add to config file (~/.config/shelly/config.yaml):")
	ios.Printf("  integrator:\n")
	ios.Printf("    tag: your-integrator-tag\n")
	ios.Printf("    token: your-integrator-token\n")
}
