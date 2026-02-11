// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/tj-smith47/shelly-cli/internal/iostreams/render"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

const (
	// tickerInterval is the animation frame interval, aligned with Docker BuildKit's TTY_DISPLAY_RATE.
	tickerInterval = 150 * time.Millisecond

	// minRenderInterval prevents excessive redraws during rapid concurrent updates.
	minRenderInterval = 50 * time.Millisecond
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

// statusToRender maps the public Status enum to render package status strings.
func statusToRender(s Status) string {
	switch s {
	case StatusPending:
		return render.StatusPending
	case StatusRunning:
		return render.StatusRunning
	case StatusSuccess:
		return render.StatusSuccess
	case StatusError:
		return render.StatusError
	case StatusSkipped:
		return render.StatusSkipped
	default:
		return render.StatusPending
	}
}

// renderToStatus maps render package status strings back to the public Status enum.
func renderToStatus(s string) Status {
	switch s {
	case render.StatusPending:
		return StatusPending
	case render.StatusRunning:
		return StatusRunning
	case render.StatusSuccess:
		return StatusSuccess
	case render.StatusError:
		return StatusError
	case render.StatusSkipped:
		return StatusSkipped
	default:
		return StatusPending
	}
}

// Line represents a single output line that can be updated.
type Line struct {
	ID      string
	Status  Status
	Message string
}

// MultiWriter manages multiple concurrent output lines.
// It provides Docker BuildKit-style multi-line progress output where each target
// gets its own line that updates in place (for TTY) or prints sequentially
// (for non-TTY). Uses animated spinners, dirty-line tracking, and rate-limited
// rendering adapted from mc-cli's render engine.
type MultiWriter struct {
	mu        sync.Mutex
	out       io.Writer
	regions   []*render.LineRegion
	regionMap map[string]*render.LineRegion
	order     []string // preserve insertion order
	isTTY     bool
	colorFn   render.ColorFuncs

	// Render state (TTY only)
	lastHeight int
	lastLines  []string
	lastRender time.Time
	finalized  bool

	// Terminal size
	termWidth int
	getSizeFn func() (int, int) // nil disables polling

	// Ticker for spinner animation
	ticker  *time.Ticker
	stopCh  chan struct{}
	stopped bool

	// Plain mode state (non-TTY)
	plainState map[string]string // last printed status per ID
}

// NewMultiWriter creates a multi-line writer.
func NewMultiWriter(out io.Writer, isTTY bool) *MultiWriter {
	m := &MultiWriter{
		out:        out,
		regionMap:  make(map[string]*render.LineRegion),
		isTTY:      isTTY,
		plainState: make(map[string]string),
	}

	if isTTY {
		m.colorFn = buildColorFuncs()
		m.getSizeFn = getTerminalSize
		if w, _ := getTerminalSize(); w > 0 {
			m.termWidth = w
		}
	} else {
		m.colorFn = render.NoColor()
	}

	return m
}

// AddLine adds a new tracked line.
func (m *MultiWriter) AddLine(id, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r := &render.LineRegion{
		ID:      id,
		Status:  render.StatusPending,
		Message: message,
	}
	m.regions = append(m.regions, r)
	m.regionMap[id] = r
	m.order = append(m.order, id)

	// Start ticker on first line (TTY only)
	m.startTickerLocked()

	if m.isTTY {
		m.forceRenderLocked()
	}
}

// UpdateLine updates an existing line.
func (m *MultiWriter) UpdateLine(id string, status Status, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r := m.regionMap[id]
	if r == nil {
		return
	}

	newStatus := statusToRender(status)

	// Track timing on status transitions
	if newStatus == render.StatusRunning && r.Status != render.StatusRunning {
		r.StartTime = time.Now()
	}
	if newStatus != render.StatusRunning && r.Status == render.StatusRunning && !r.StartTime.IsZero() {
		r.Duration = render.FormatElapsed(time.Since(r.StartTime))
	}

	r.Status = newStatus
	r.Message = message

	if m.isTTY {
		m.forceRenderLocked()
	} else {
		m.renderPlainLocked(id)
	}
}

// GetLine returns a copy of the line for the given ID.
func (m *MultiWriter) GetLine(id string) (Line, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r := m.regionMap[id]
	if r == nil {
		return Line{}, false
	}
	return Line{
		ID:      r.ID,
		Status:  renderToStatus(r.Status),
		Message: r.Message,
	}, true
}

// LineCount returns the number of tracked lines.
func (m *MultiWriter) LineCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.regions)
}

// Finalize writes the final state and cleans up.
func (m *MultiWriter) Finalize() {
	// Stop ticker first (outside lock to avoid deadlock with tick loop)
	m.stopTicker()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.finalized {
		return
	}
	m.finalized = true

	if m.isTTY {
		m.forceRenderLocked()
		writeQuietly(m.out, "%s", render.ShowCursor())
	} else {
		// Non-TTY: print any remaining lines not yet printed
		for _, id := range m.order {
			r := m.regionMap[id]
			if m.plainState[id] != r.Status {
				line := render.FormatLine(r, m.colorFn)
				writelnQuietly(m.out, line)
				m.plainState[id] = r.Status
			}
		}
	}
}

// Summary returns a summary of the operation results.
func (m *MultiWriter) Summary() (success, failed, skipped int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.regions {
		switch r.Status {
		case render.StatusSuccess:
			success++
		case render.StatusError:
			failed++
		case render.StatusSkipped:
			skipped++
		}
	}
	return
}

// PrintSummary prints a summary line with counts.
func (m *MultiWriter) PrintSummary() {
	success, failed, skipped := m.Summary()
	total := m.LineCount()

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

	writelnQuietly(m.out, strings.Join(parts, ", "))
}

// HasErrors returns true if any line has an error status.
func (m *MultiWriter) HasErrors() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.regions {
		if r.Status == render.StatusError {
			return true
		}
	}
	return false
}

// AllSucceeded returns true if all lines have success status.
func (m *MultiWriter) AllSucceeded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.regions {
		if r.Status != render.StatusSuccess && r.Status != render.StatusSkipped {
			return false
		}
	}
	return true
}

// startTickerLocked starts the background animation ticker. Must be called with m.mu held.
func (m *MultiWriter) startTickerLocked() {
	if !m.isTTY || m.ticker != nil {
		return
	}
	m.stopCh = make(chan struct{})
	m.ticker = time.NewTicker(tickerInterval)
	go m.tickLoop()
}

// stopTicker stops the background animation ticker. Idempotent.
func (m *MultiWriter) stopTicker() {
	m.mu.Lock()
	if m.stopped || m.ticker == nil {
		m.mu.Unlock()
		return
	}
	m.stopped = true
	m.mu.Unlock()

	m.ticker.Stop()
	close(m.stopCh)
}

// tickLoop is the background goroutine that drives spinner animation.
func (m *MultiWriter) tickLoop() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.ticker.C:
			m.mu.Lock()
			if m.finalized {
				m.mu.Unlock()
				return
			}
			m.advanceSpinners()
			m.renderLocked()
			m.mu.Unlock()
		}
	}
}

// advanceSpinners increments SpinnerFrame on all running regions.
func (m *MultiWriter) advanceSpinners() {
	for _, r := range m.regions {
		if r.Status == render.StatusRunning {
			r.SpinnerFrame++
		}
	}
}

// renderLocked performs a rate-limited render. Must be called with m.mu held.
func (m *MultiWriter) renderLocked() {
	if m.finalized || !m.isTTY {
		return
	}
	if time.Since(m.lastRender) < minRenderInterval {
		return
	}
	m.doRender()
}

// forceRenderLocked renders regardless of rate limiting. Must be called with m.mu held.
func (m *MultiWriter) forceRenderLocked() {
	if m.finalized && !m.isTTY {
		return
	}
	m.doRender()
}

// doRender performs the actual TTY render with dirty-line tracking. Must be called with m.mu held.
func (m *MultiWriter) doRender() {
	m.pollSize()
	lines := m.buildLines()

	// Apply terminal width truncation
	truncated := make([]string, len(lines))
	for i, line := range lines {
		if m.termWidth > 0 {
			line = render.TruncateLine(line, m.termWidth)
		}
		truncated[i] = line
	}

	var buf strings.Builder

	// Hide cursor during redraw to prevent flicker
	buf.WriteString(render.HideCursor())

	// Move cursor back to start of previous output
	if m.lastHeight > 0 {
		buf.WriteString(render.MoveUp(m.lastHeight))
	}

	// Write lines, skipping unchanged ones (dirty-line tracking)
	for i, line := range truncated {
		if i < len(m.lastLines) && line == m.lastLines[i] {
			buf.WriteString(render.MoveDown(1))
		} else {
			buf.WriteString("\r")
			buf.WriteString(render.ClearLine())
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	// Clear leftover lines from previous render
	if len(truncated) < m.lastHeight {
		extra := m.lastHeight - len(truncated)
		for range extra {
			buf.WriteString(render.ClearLine())
			buf.WriteString("\n")
		}
		buf.WriteString(render.MoveUp(extra))
	}

	// Show cursor after redraw
	buf.WriteString(render.ShowCursor())

	writeQuietly(m.out, "%s", buf.String())
	m.lastHeight = len(truncated)
	m.lastLines = truncated
	m.lastRender = time.Now()
}

// buildLines collects formatted lines from all regions.
func (m *MultiWriter) buildLines() []string {
	lines := make([]string, 0, len(m.regions))
	for _, r := range m.regions {
		lines = append(lines, render.FormatLine(r, m.colorFn))
	}
	return lines
}

// pollSize re-queries terminal dimensions.
func (m *MultiWriter) pollSize() {
	if m.getSizeFn == nil {
		return
	}
	w, _ := m.getSizeFn()
	if w > 0 {
		m.termWidth = w
	}
}

// renderPlainLocked prints a status line for non-TTY output when status changes.
// Must be called with m.mu held.
func (m *MultiWriter) renderPlainLocked(id string) {
	r := m.regionMap[id]
	if r == nil {
		return
	}
	if m.plainState[id] == r.Status {
		return // Status unchanged, skip
	}
	line := render.FormatLine(r, m.colorFn)
	writelnQuietly(m.out, line)
	m.plainState[id] = r.Status
}

// buildColorFuncs creates ColorFuncs from the shelly-cli theme package.
func buildColorFuncs() render.ColorFuncs {
	wrap := func(style func(...string) string) func(string) string {
		return func(s string) string { return style(s) }
	}
	return render.ColorFuncs{
		Primary:   wrap(theme.Highlight().Render),
		Secondary: wrap(theme.Dim().Render),
		Tertiary:  wrap(theme.Subtitle().Render),
		Faint:     wrap(theme.Dim().Render),
		Error:     wrap(theme.StatusError().Render),
		Yellow:    wrap(theme.StatusWarn().Render),
		Green:     wrap(theme.StatusOK().Render),
	}
}

// getTerminalSize returns the current terminal width and height.
func getTerminalSize() (width, height int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, 0
	}
	return w, h
}
