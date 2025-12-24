// Package debug provides TUI debug logging for development and troubleshooting.
// This logs the rendered view output to a file for debugging TUI layout issues.
//
// Debug logs are stored in ~/.config/shelly/debug/<timestamp>/ where each TUI
// session gets its own timestamped folder containing:
//   - tui.log: Main debug log with view renders and events
package debug

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const (
	// EnvKey is the environment variable to enable debug logging.
	EnvKey = "SHELLY_TUI_DEBUG"
	// DebugDir is the subdirectory under config for debug logs.
	DebugDir = "debug"
	// MainLogFile is the main log file name.
	MainLogFile = "tui.log"
)

// Logger writes TUI debug information to a log file.
type Logger struct {
	mu         sync.Mutex
	enabled    bool
	file       *os.File
	sessionDir string
	lastView   string
	startTime  time.Time
}

// New creates a new Logger if SHELLY_TUI_DEBUG=1 is set.
// Creates a timestamped session folder under ~/.config/shelly/debug/.
// Returns nil if debug logging is disabled.
func New() *Logger {
	if os.Getenv(EnvKey) != "1" {
		return nil
	}

	configDir, err := config.Dir()
	if err != nil {
		iostreams.DebugErr("get config dir", err)
		return nil
	}

	// Create timestamped session directory
	startTime := time.Now()
	sessionName := startTime.Format("2006-01-02_15-04-05")
	sessionDir := filepath.Join(configDir, DebugDir, sessionName)

	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		iostreams.DebugErr("create debug session dir", err)
		return nil
	}

	logPath := filepath.Join(sessionDir, MainLogFile)
	//nolint:gosec // Path constructed from config.Dir() + fixed constants, not user input
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		iostreams.DebugErr("open log file", err)
		return nil
	}

	l := &Logger{
		enabled:    true,
		file:       f,
		sessionDir: sessionDir,
		startTime:  startTime,
	}

	// Write header
	l.writeHeader()

	return l
}

// SessionDir returns the path to the current debug session directory.
func (l *Logger) SessionDir() string {
	if l == nil {
		return ""
	}
	return l.sessionDir
}

// writeHeader writes a session start header to the log.
func (l *Logger) writeHeader() {
	header := fmt.Sprintf("%s\nTUI Debug Session\n%s\nStarted: %s\nSession: %s\n%s\n\n",
		strings.Repeat("=", 60),
		strings.Repeat("=", 60),
		l.startTime.Format(time.RFC3339),
		l.sessionDir,
		strings.Repeat("=", 60))
	if _, err := l.file.WriteString(header); err != nil {
		iostreams.DebugErr("write log header", err)
	}
}

// Log writes a debug entry for the current view state.
func (l *Logger) Log(tab, focus string, width, height int, view string) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Skip if view hasn't changed
	if view == l.lastView {
		return
	}
	l.lastView = view

	// Format log entry
	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	entry := fmt.Sprintf("[%s] Tab: %s | Focus: %s | Size: %dx%d\n--- VIEW OUTPUT ---\n%s\n--- END VIEW ---\n\n",
		timestamp, tab, focus, width, height, view)

	if _, err := l.file.WriteString(entry); err != nil {
		iostreams.DebugErr("write log entry", err)
	}
}

// LogEvent writes a simple event message.
func (l *Logger) LogEvent(event string) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	entry := fmt.Sprintf("[%s] EVENT: %s\n", timestamp, event)

	if _, err := l.file.WriteString(entry); err != nil {
		iostreams.DebugErr("write log event", err)
	}
}

// Close closes the logger and writes session summary.
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Write footer with session duration
	duration := time.Since(l.startTime)
	footer := fmt.Sprintf("\n%s\nSession Ended: %s\nDuration: %s\n%s\n",
		strings.Repeat("=", 60),
		time.Now().Format(time.RFC3339),
		duration.Round(time.Second),
		strings.Repeat("=", 60))
	if _, err := l.file.WriteString(footer); err != nil {
		iostreams.DebugErr("write log footer", err)
	}

	return l.file.Close()
}

// Enabled returns whether debug logging is active.
func (l *Logger) Enabled() bool {
	return l != nil && l.enabled
}

// Writer returns an io.Writer that writes to the debug log.
// Returns io.Discard if logging is disabled.
func (l *Logger) Writer() io.Writer {
	if l == nil || !l.enabled {
		return io.Discard
	}
	return l.file
}
