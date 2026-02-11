package render

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
)

// Status constants for line regions.
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusError   = "error"
	StatusSkipped = "skipped"
)

// LineRegion tracks a single line's display state within multi-line progress output.
type LineRegion struct {
	ID           string
	Status       string
	Message      string
	StartTime    time.Time // when the line transitioned to running
	Duration     string    // formatted duration, set on completion
	SpinnerFrame int       // current spinner animation frame
}

// ColorFuncs holds theme coloring functions used for rendering.
type ColorFuncs struct {
	Primary   func(string) string
	Secondary func(string) string
	Tertiary  func(string) string
	Faint     func(string) string
	Error     func(string) string
	Yellow    func(string) string
	Green     func(string) string
}

// NoColor returns a ColorFuncs that applies no coloring.
func NoColor() ColorFuncs {
	identity := func(s string) string { return s }
	return ColorFuncs{
		Primary:   identity,
		Secondary: identity,
		Tertiary:  identity,
		Faint:     identity,
		Error:     identity,
		Yellow:    identity,
		Green:     identity,
	}
}

// spinnerChar returns the spinner character for the given animation frame.
// Uses charmbracelet/bubbles MiniDot frames for consistency with TUI spinners.
func spinnerChar(frame int) string {
	frames := spinner.MiniDot.Frames
	return frames[frame%len(frames)]
}

// FormatElapsed formats a completed duration without decimals: "12s", "1m30s".
func FormatElapsed(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	return fmt.Sprintf("%dm%ds", secs/60, secs%60)
}

// formatRunningElapsed formats a live running duration with one decimal: "5.2s", "1m30.5s".
func formatRunningElapsed(d time.Duration) string {
	total := d.Seconds()
	if total < 60 {
		return fmt.Sprintf("%.1fs", total)
	}
	mins := int(total) / 60
	secs := total - float64(mins*60)
	return fmt.Sprintf("%dm%.1fs", mins, secs)
}

// FormatLine renders a single status line for a LineRegion.
func FormatLine(r *LineRegion, colorFn ColorFuncs) string {
	switch r.Status {
	case StatusPending:
		return formatPendingLine(r, colorFn)
	case StatusRunning:
		return formatRunningLine(r, colorFn)
	case StatusSuccess:
		return formatSuccessLine(r, colorFn)
	case StatusError:
		return formatErrorLine(r, colorFn)
	case StatusSkipped:
		return formatSkippedLine(r, colorFn)
	default:
		return fmt.Sprintf("? %s: %s", r.ID, r.Message)
	}
}

func formatPendingLine(r *LineRegion, colorFn ColorFuncs) string {
	return fmt.Sprintf("%s %s: %s",
		colorFn.Faint("○"),
		colorFn.Faint(r.ID),
		colorFn.Faint(r.Message),
	)
}

func formatRunningLine(r *LineRegion, colorFn ColorFuncs) string {
	icon := colorFn.Primary(spinnerChar(r.SpinnerFrame))
	line := fmt.Sprintf("%s %s: %s", icon, colorFn.Primary(r.ID), r.Message)
	if !r.StartTime.IsZero() {
		elapsed := formatRunningElapsed(time.Since(r.StartTime))
		line += " " + colorFn.Tertiary(elapsed)
	}
	return line
}

func formatSuccessLine(r *LineRegion, colorFn ColorFuncs) string {
	line := fmt.Sprintf("%s %s: %s",
		colorFn.Green("\u2714"),
		colorFn.Green(r.ID),
		colorFn.Secondary(r.Message),
	)
	if r.Duration != "" {
		line += " " + colorFn.Tertiary("("+r.Duration+")")
	}
	return line
}

func formatErrorLine(r *LineRegion, colorFn ColorFuncs) string {
	line := fmt.Sprintf("%s %s: %s",
		colorFn.Error("\u2718"),
		colorFn.Error(r.ID),
		colorFn.Error(r.Message),
	)
	if r.Duration != "" {
		line += " " + colorFn.Tertiary("("+r.Duration+")")
	}
	return line
}

func formatSkippedLine(r *LineRegion, colorFn ColorFuncs) string {
	return fmt.Sprintf("%s %s: %s",
		colorFn.Faint("\u2298"),
		colorFn.Faint(r.ID),
		colorFn.Faint(r.Message),
	)
}
