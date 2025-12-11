// Package output provides output formatting utilities for the CLI.
package output

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Message types for consistent output formatting

// Info prints an informational message with theme styling.
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
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.StatusInfo().Render("ℹ")+" "+msg)
}

// Success prints a success message with theme styling.
func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.StatusOK().Render("✓") + " " + msg)
}

// SuccessTo prints a success message to the specified writer.
func SuccessTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.StatusOK().Render("✓")+" "+msg)
}

// Warning prints a warning message with theme styling.
func Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, theme.StatusWarn().Render("⚠")+" "+msg)
}

// WarningTo prints a warning message to the specified writer.
func WarningTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.StatusWarn().Render("⚠")+" "+msg)
}

// Error prints an error message with theme styling.
// Note: This is for display purposes. For actual errors, return an error from the command.
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, theme.StatusError().Render("✗")+" "+msg)
}

// ErrorTo prints an error message to the specified writer.
func ErrorTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.StatusError().Render("✗")+" "+msg)
}

// Plain prints a message without any styling.
func Plain(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// PlainTo prints a plain message to the specified writer.
func PlainTo(w io.Writer, format string, args ...any) {
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintf(w, format+"\n", args...)
}

// Hint prints a helpful tip or suggestion with dim styling.
func Hint(format string, args ...any) {
	if viper.GetBool("quiet") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.Dim().Render("  " + msg))
}

// HintTo prints a hint message to the specified writer.
func HintTo(w io.Writer, format string, args ...any) {
	if viper.GetBool("quiet") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.Dim().Render("  "+msg))
}

// Title prints a section title with bold styling.
func Title(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.Title().Render(msg))
}

// TitleTo prints a title to the specified writer.
func TitleTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.Title().Render(msg))
}

// Subtitle prints a subtitle with subdued styling.
func Subtitle(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(theme.Subtitle().Render(msg))
}

// SubtitleTo prints a subtitle to the specified writer.
func SubtitleTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintln(w, theme.Subtitle().Render(msg))
}

// Count prints a count summary (e.g., "Found 5 device(s)").
func Count(noun string, count int) {
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Printf("\nFound %d %s%s\n", count, noun, suffix)
}

// CountTo prints a count summary to the specified writer.
func CountTo(w io.Writer, noun string, count int) {
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	//nolint:errcheck // Best-effort output to terminal
	fmt.Fprintf(w, "\nFound %d %s%s\n", count, noun, suffix)
}

// NoResults prints a "no results" message with an optional hint.
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
