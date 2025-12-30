// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// isQuietMode checks if quiet mode is enabled via viper.
func isQuietMode() bool {
	return viper.GetBool("quiet")
}

// Message output functions that integrate with IOStreams.
// These functions output styled messages to the appropriate streams.

// Info prints an informational message with theme styling.
// Messages are suppressed in quiet mode.
func (s *IOStreams) Info(format string, args ...any) {
	if s.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, theme.StatusInfo().Render("→")+" "+msg)
}

// Success prints a success message with theme styling.
func (s *IOStreams) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, theme.StatusOK().Render("✓")+" "+msg)
}

// Warning prints a warning message with theme styling.
// Warnings go to stderr.
func (s *IOStreams) Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.ErrOut, theme.StatusWarn().Render("⚠")+" "+msg)
}

// Error prints an error message with theme styling.
// Errors go to stderr.
// Note: For actual command errors, return an error from the command instead.
func (s *IOStreams) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.ErrOut, theme.StatusError().Render("✗")+" "+msg)
}

// Plain prints a message without any styling.
func (s *IOStreams) Plain(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, msg)
}

// Hint prints a helpful tip or suggestion with dim styling.
// Hints are suppressed in quiet mode.
func (s *IOStreams) Hint(format string, args ...any) {
	if s.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, theme.Dim().Render("  "+msg))
}

// Title prints a section title with bold styling.
func (s *IOStreams) Title(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, theme.Title().Render(msg))
}

// Subtitle prints a subtitle with subdued styling.
func (s *IOStreams) Subtitle(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(s.Out, theme.Subtitle().Render(msg))
}

// Count prints a count summary (e.g., "Found 5 device(s)").
// Callers should add ios.Println() before Count() if vertical separation is desired.
func (s *IOStreams) Count(noun string, count int) {
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	writeQuietly(s.Out, "Found %d %s%s\n", count, noun, suffix)
}

// NoResults prints a "no results" message with optional hints.
func (s *IOStreams) NoResults(itemType string, hints ...string) {
	writeQuietly(s.Out, "No %s found\n", itemType)
	for _, hint := range hints {
		s.Hint("%s", hint)
	}
}

// Added prints a count of items added.
func (s *IOStreams) Added(noun string, count int) {
	if count == 0 {
		return
	}
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	s.Success("Added %d %s%s", count, noun, suffix)
}

// UpdateInfo displays version information before an update.
func (s *IOStreams) UpdateInfo(currentVersion, availableVersion, releaseNotes string) {
	s.Printf("\nCurrent version: %s\n", currentVersion)
	s.Printf("Available version: %s\n", availableVersion)
	if releaseNotes != "" {
		s.Printf("\nRelease notes:\n%s\n", releaseNotes)
	}
}

// RollbackInfo displays version information before a rollback.
func (s *IOStreams) RollbackInfo(currentVersion, targetVersion string) {
	s.Printf("Rolling back from %s to %s\n", currentVersion, targetVersion)
}

// Static message functions that write to specific writers.
// These are useful when you don't have an IOStreams instance but need styled output.

// InfoTo prints an informational message to the specified writer.
func InfoTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.StatusInfo().Render("→")+" "+msg)
}

// SuccessTo prints a success message to the specified writer.
func SuccessTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.StatusOK().Render("✓")+" "+msg)
}

// WarningTo prints a warning message to the specified writer.
func WarningTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.StatusWarn().Render("⚠")+" "+msg)
}

// ErrorTo prints an error message to the specified writer.
func ErrorTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.StatusError().Render("✗")+" "+msg)
}

// PlainTo prints a plain message to the specified writer.
func PlainTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, msg)
}

// HintTo prints a hint message to the specified writer.
func HintTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.Dim().Render("  "+msg))
}

// TitleTo prints a title to the specified writer.
func TitleTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.Title().Render(msg))
}

// SubtitleTo prints a subtitle to the specified writer.
func SubtitleTo(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writelnQuietly(w, theme.Subtitle().Render(msg))
}

// CountTo prints a count summary to the specified writer.
// Callers should print a blank line before CountTo() if vertical separation is desired.
func CountTo(w io.Writer, noun string, count int) {
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	writeQuietly(w, "Found %d %s%s\n", count, noun, suffix)
}

// Package-level convenience functions that write to stdout/stderr.
// These provide backwards-compatible APIs for code migrating from ui/output packages.

// Info prints an informational message to stdout.
// Messages are suppressed in quiet mode.
func Info(format string, args ...any) {
	if isQuietMode() {
		return
	}
	InfoTo(os.Stdout, format, args...)
}

// Success prints a success message to stdout.
func Success(format string, args ...any) {
	SuccessTo(os.Stdout, format, args...)
}

// Warning prints a warning message to stderr.
func Warning(format string, args ...any) {
	WarningTo(os.Stderr, format, args...)
}

// Error prints an error message to stderr.
// Note: For actual command errors, return an error from the command instead.
func Error(format string, args ...any) {
	ErrorTo(os.Stderr, format, args...)
}

// Plain prints a message without any styling to stdout.
func Plain(format string, args ...any) {
	PlainTo(os.Stdout, format, args...)
}

// Hint prints a helpful tip or suggestion to stdout.
// Hints are suppressed in quiet mode.
func Hint(format string, args ...any) {
	if isQuietMode() {
		return
	}
	HintTo(os.Stdout, format, args...)
}

// Title prints a section title to stdout.
func Title(format string, args ...any) {
	TitleTo(os.Stdout, format, args...)
}

// Subtitle prints a subtitle to stdout.
func Subtitle(format string, args ...any) {
	SubtitleTo(os.Stdout, format, args...)
}

// Count prints a count summary to stdout.
func Count(noun string, count int) {
	CountTo(os.Stdout, noun, count)
}

// NoResults prints a "no results" message to stdout with optional hints.
func NoResults(itemType string, hints ...string) {
	writeQuietly(os.Stdout, "No %s found\n", itemType)
	for _, hint := range hints {
		Hint("%s", hint)
	}
}

// Added prints a count of items added to stdout.
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
