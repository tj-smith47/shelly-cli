// Package debug provides TUI debug logging for development and troubleshooting.
// This logs the rendered view output to a file for debugging TUI layout issues.
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
	// MaxLogSize is the maximum size of a log file before rotation (5MB).
	MaxLogSize = 5 * 1024 * 1024
	// MaxBackups is the number of backup log files to keep.
	MaxBackups = 3
	// EnvKey is the environment variable to enable debug logging.
	EnvKey = "SHELLY_TUI_DEBUG"
)

// Logger writes TUI debug information to a log file.
type Logger struct {
	mu       sync.Mutex
	enabled  bool
	file     *os.File
	path     string
	maxSize  int64
	size     int64
	lastView string
}

// New creates a new Logger if SHELLY_TUI_DEBUG=1 is set.
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

	if err := os.MkdirAll(configDir, 0o700); err != nil {
		iostreams.DebugErr("create config dir", err)
		return nil
	}

	logPath := filepath.Join(configDir, "tui-debug.log")
	//nolint:gosec // Path constructed from config.Dir() + fixed suffix, not user input
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		iostreams.DebugErr("open log file", err)
		return nil
	}

	// Get current file size
	info, err := f.Stat()
	if err != nil {
		iostreams.DebugErr("stat log file", err)
		iostreams.CloseWithDebug("close log file", f)
		return nil
	}

	l := &Logger{
		enabled: true,
		file:    f,
		path:    logPath,
		maxSize: MaxLogSize,
		size:    info.Size(),
	}

	// Write header
	l.writeHeader()

	return l
}

// writeHeader writes a session start header to the log.
func (l *Logger) writeHeader() {
	header := fmt.Sprintf("\n%s\n=== TUI Debug Session Started ===\n%s\n\n",
		strings.Repeat("=", 50),
		time.Now().Format(time.RFC3339))
	n, err := l.file.WriteString(header)
	if err != nil {
		iostreams.DebugErr("write log header", err)
		return
	}
	l.size += int64(n)
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

	// Check if rotation is needed
	if l.size >= l.maxSize {
		l.rotate()
	}

	// Format log entry
	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	entry := fmt.Sprintf("[%s] Tab: %s | Focus: %s | Size: %dx%d\n--- VIEW OUTPUT ---\n%s\n--- END VIEW ---\n\n",
		timestamp, tab, focus, width, height, view)

	n, err := l.file.WriteString(entry)
	if err != nil {
		iostreams.DebugErr("write log entry", err)
		return
	}
	l.size += int64(n)
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

	n, err := l.file.WriteString(entry)
	if err != nil {
		iostreams.DebugErr("write log event", err)
		return
	}
	l.size += int64(n)
}

// rotate rotates the log file, keeping up to MaxBackups old files.
func (l *Logger) rotate() {
	if l.file == nil {
		return
	}

	// Close current file
	iostreams.CloseWithDebug("close for rotation", l.file)

	// Rotate existing backups
	for i := MaxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.path, i)
		newPath := fmt.Sprintf("%s.%d", l.path, i+1)
		if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
			iostreams.DebugErr("rotate backup", err)
		}
	}

	// Rename current to .1
	if err := os.Rename(l.path, l.path+".1"); err != nil && !os.IsNotExist(err) {
		iostreams.DebugErr("rotate current", err)
	}

	// Create new log file
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		iostreams.DebugErr("create new log", err)
		l.enabled = false
		return
	}

	l.file = f
	l.size = 0
	l.lastView = ""
	l.writeHeader()
}

// Close closes the logger.
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Write footer
	footer := fmt.Sprintf("\n=== TUI Debug Session Ended ===\n%s\n%s\n",
		time.Now().Format(time.RFC3339),
		strings.Repeat("=", 50))
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
