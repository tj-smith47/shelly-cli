// Package ui provides user interaction components for the CLI.
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Info prints an informational message.
func Info(format string, args ...any) {
	if viper.GetBool("quiet") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.StatusInfo().Render("ℹ") + " " + msg)
}

// InfoTo prints an informational message to the specified writer.
func InfoTo(w io.Writer, format string, args ...any) {
	if viper.GetBool("quiet") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprintln(w, theme.StatusInfo().Render("ℹ")+" "+msg); err != nil {
		DebugErr("writing info message", err)
	}
}

// Success prints a success message.
func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.StatusOK().Render("✓") + " " + msg)
}

// Warning prints a warning message.
func Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprintln(os.Stderr, theme.StatusWarn().Render("⚠")+" "+msg); err != nil {
		DebugErr("writing warning message", err)
	}
}

// Error prints an error message.
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprintln(os.Stderr, theme.StatusError().Render("✗")+" "+msg); err != nil {
		DebugErr("writing error message", err)
	}
}

// Title prints a section title.
func Title(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.Title().Render(msg))
}

// Hint prints a helpful tip with dim styling.
func Hint(format string, args ...any) {
	if viper.GetBool("quiet") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.Dim().Render("  " + msg))
}

// Count prints a count summary (e.g., "Found 5 device(s)").
func Count(noun string, count int) {
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Printf("\nFound %d %s%s\n", count, noun, suffix)
}

// NoResults prints a "no results" message with optional hints.
func NoResults(itemType string, hints ...string) {
	fmt.Printf("No %s found\n", itemType)
	for _, hint := range hints {
		Hint("%s", hint)
	}
}

// Added prints a count of items added.
func Added(noun string, count int) {
	if count == 0 {
		return
	}
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	Success("Added %d %s%s", count, noun, suffix)
}
