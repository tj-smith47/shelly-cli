// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"
	"io"
	"sync"

	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Status represents the status of a line in the multi-writer output.
type Status int

const (
	// StatusPending indicates the operation has not started.
	StatusPending Status = iota
	// StatusRunning indicates the operation is in progress.
	StatusRunning
	// StatusSuccess indicates the operation completed successfully.
	StatusSuccess
	// StatusError indicates the operation failed.
	StatusError
	// StatusSkipped indicates the operation was skipped.
	StatusSkipped
)

// String returns a string representation of the status.
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusError:
		return "error"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Line represents a single output line that can be updated.
type Line struct {
	ID      string
	Status  Status
	Message string
}

// MultiWriter manages multiple concurrent output lines.
// It provides Docker-style multi-line progress output where each target
// gets its own line that updates in place (for TTY) or prints sequentially
// (for non-TTY).
type MultiWriter struct {
	mu    sync.Mutex
	out   io.Writer
	lines map[string]*Line
	order []string // Preserve insertion order
	isTTY bool
}

// NewMultiWriter creates a multi-line writer.
func NewMultiWriter(out io.Writer, isTTY bool) *MultiWriter {
	return &MultiWriter{
		out:   out,
		lines: make(map[string]*Line),
		isTTY: isTTY,
	}
}

// AddLine adds a new tracked line.
func (m *MultiWriter) AddLine(id, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lines[id] = &Line{
		ID:      id,
		Status:  StatusPending,
		Message: message,
	}
	m.order = append(m.order, id)

	// For TTY, print the initial line
	if m.isTTY {
		m.printLine(m.lines[id])
	}
}

// UpdateLine updates an existing line.
func (m *MultiWriter) UpdateLine(id string, status Status, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	line, ok := m.lines[id]
	if !ok {
		return
	}

	line.Status = status
	line.Message = message

	if m.isTTY {
		m.render()
	}
}

// GetLine returns a copy of the line for the given ID.
func (m *MultiWriter) GetLine(id string) (Line, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	line, ok := m.lines[id]
	if !ok {
		return Line{}, false
	}
	return *line, true
}

// LineCount returns the number of tracked lines.
func (m *MultiWriter) LineCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.lines)
}

// render redraws all lines (TTY only).
func (m *MultiWriter) render() {
	if !m.isTTY {
		return
	}

	// Move cursor up to start of our output
	if len(m.order) > 1 {
		writeQuietly(m.out, "\033[%dA", len(m.order))
	} else if len(m.order) == 1 {
		writeQuietly(m.out, "\033[1A")
	}

	for _, id := range m.order {
		line := m.lines[id]
		writeQuietly(m.out, "\033[2K") // Clear line
		m.printLine(line)
	}
}

// printLine prints a single line with status icon and styling.
func (m *MultiWriter) printLine(line *Line) {
	icon := m.statusIcon(line.Status)
	style := m.statusStyle(line.Status)
	idStyle := m.idStyle(line.Status)

	writeQuietly(m.out, "%s %s: %s\n",
		icon,
		idStyle.Render(line.ID),
		style.Render(line.Message),
	)
}

// statusIcon returns the appropriate icon for a status.
func (m *MultiWriter) statusIcon(s Status) string {
	switch s {
	case StatusPending:
		return theme.Dim().Render("○")
	case StatusRunning:
		return theme.StatusWarn().Render("◐")
	case StatusSuccess:
		return theme.StatusOK().Render("✓")
	case StatusError:
		return theme.StatusError().Render("✗")
	case StatusSkipped:
		return theme.Dim().Render("⊘")
	default:
		return "?"
	}
}

// statusStyle returns the appropriate lipgloss style for message text.
func (m *MultiWriter) statusStyle(s Status) lipgloss.Style {
	switch s {
	case StatusSuccess:
		return theme.Dim()
	case StatusError:
		return theme.StatusError()
	case StatusRunning:
		return lipgloss.NewStyle()
	case StatusSkipped:
		return theme.Dim()
	default:
		return theme.Dim()
	}
}

// idStyle returns the appropriate style for the ID/name.
func (m *MultiWriter) idStyle(s Status) lipgloss.Style {
	switch s {
	case StatusSuccess:
		return theme.StatusOK()
	case StatusError:
		return theme.StatusError()
	case StatusRunning:
		return theme.Bold()
	default:
		return lipgloss.NewStyle()
	}
}

// Finalize prints final state (for non-TTY or completion).
// For non-TTY outputs, this prints all lines since they weren't printed during updates.
// For TTY outputs, this is a no-op since lines are already visible.
func (m *MultiWriter) Finalize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isTTY {
		// Non-TTY: print each line once at the end
		for _, id := range m.order {
			line := m.lines[id]
			m.printLine(line)
		}
	}
	// For TTY, lines are already rendered - nothing to do
}

// Summary returns a summary of the operation results.
func (m *MultiWriter) Summary() (success, failed, skipped int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, line := range m.lines {
		switch line.Status {
		case StatusSuccess:
			success++
		case StatusError:
			failed++
		case StatusSkipped:
			skipped++
		case StatusPending, StatusRunning:
			// Not counted in summary - operation still in progress
		}
	}
	return
}

// PrintSummary prints a summary line with counts.
func (m *MultiWriter) PrintSummary() {
	success, failed, skipped := m.Summary()
	total := len(m.lines)

	parts := make([]string, 0, 3)
	if success > 0 {
		parts = append(parts, theme.StatusOK().Render(fmt.Sprintf("%d succeeded", success)))
	}
	if failed > 0 {
		parts = append(parts, theme.StatusError().Render(fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, theme.Dim().Render(fmt.Sprintf("%d skipped", skipped)))
	}

	if len(parts) == 0 {
		writelnQuietly(m.out, fmt.Sprintf("Completed %d operations", total))
		return
	}

	summary := ""
	for i, part := range parts {
		if i > 0 {
			summary += ", "
		}
		summary += part
	}
	writelnQuietly(m.out, summary)
}

// HasErrors returns true if any line has an error status.
func (m *MultiWriter) HasErrors() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, line := range m.lines {
		if line.Status == StatusError {
			return true
		}
	}
	return false
}

// AllSucceeded returns true if all lines have success status.
func (m *MultiWriter) AllSucceeded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, line := range m.lines {
		if line.Status != StatusSuccess && line.Status != StatusSkipped {
			return false
		}
	}
	return true
}
