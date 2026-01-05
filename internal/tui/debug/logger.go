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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

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
	// MaxSessions is the maximum number of debug sessions to keep.
	MaxSessions = 25
)

// Logger writes TUI debug information to a log file.
type Logger struct {
	mu         sync.Mutex
	enabled    bool
	file       afero.File
	sessionDir string
	lastView   string
	startTime  time.Time
}

// globalLogger is the shared logger instance for trace logging.
var (
	globalLogger   *Logger
	globalLoggerMu sync.RWMutex
)

// SetGlobal sets the global logger instance for trace logging.
// This should be called once when the app starts.
func SetGlobal(l *Logger) {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()
	globalLogger = l
}

// getGlobal returns the global logger instance safely.
func getGlobal() *Logger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return globalLogger
}

// TraceLock logs a lock acquisition from any component.
func TraceLock(component, lockType, caller string) {
	if l := getGlobal(); l != nil {
		l.LogLock(component, lockType, caller)
	}
}

// TraceUnlock logs a lock release from any component.
func TraceUnlock(component, lockType, caller string) {
	if l := getGlobal(); l != nil {
		l.LogUnlock(component, lockType, caller)
	}
}

// TraceNetwork logs a network operation from any component.
func TraceNetwork(operation, device, method string, err error) {
	if l := getGlobal(); l != nil {
		l.LogNetwork(operation, device, method, err)
	}
}

// New creates a new Logger.
// If SHELLY_TUI_DEBUG=1 is set, creates an active session immediately.
// Otherwise returns a disabled logger that can be toggled on with Shift+D.
func New() *Logger {
	l := &Logger{
		enabled: false, // Start disabled, can be toggled with Shift+D
	}

	// If env var is set, enable immediately
	if os.Getenv(EnvKey) != "1" {
		return l
	}

	configDir, err := config.Dir()
	if err != nil {
		iostreams.DebugErr("get config dir", err)
		return l // Return disabled logger
	}

	// Create timestamped session directory
	fs := config.Fs()
	l.startTime = time.Now()
	sessionName := l.startTime.Format("2006-01-02_15-04-05")
	debugDir := filepath.Join(configDir, DebugDir)
	l.sessionDir = filepath.Join(debugDir, sessionName)

	if err := fs.MkdirAll(l.sessionDir, 0o700); err != nil {
		iostreams.DebugErr("create debug session dir", err)
		return l
	}

	// Clean up old sessions
	cleanupOldSessions(debugDir)

	logPath := filepath.Join(l.sessionDir, MainLogFile)
	f, err := fs.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		iostreams.DebugErr("open log file", err)
		return l
	}

	l.file = f
	l.enabled = true

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

// LogLock writes a lock acquisition trace message.
// component is the component name (e.g., "cache", "events"), lockType is "Lock" or "RLock".
func (l *Logger) LogLock(component, lockType, caller string) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	entry := fmt.Sprintf("[%s] LOCK: %s.%s() from %s\n", timestamp, component, lockType, caller)

	if _, err := l.file.WriteString(entry); err != nil {
		iostreams.DebugErr("write lock trace", err)
	}
}

// LogUnlock writes a lock release trace message.
func (l *Logger) LogUnlock(component, lockType, caller string) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	entry := fmt.Sprintf("[%s] UNLOCK: %s.%s() from %s\n", timestamp, component, lockType, caller)

	if _, err := l.file.WriteString(entry); err != nil {
		iostreams.DebugErr("write unlock trace", err)
	}
}

// LogNetwork writes a network operation trace message.
// operation is "start" or "end", device is the device name/address, method is the API call.
func (l *Logger) LogNetwork(operation, device, method string, err error) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000")
	var entry string
	if err != nil {
		entry = fmt.Sprintf("[%s] NET %s: %s %s (err: %v)\n", timestamp, operation, device, method, err)
	} else {
		entry = fmt.Sprintf("[%s] NET %s: %s %s\n", timestamp, operation, device, method)
	}

	if _, err := l.file.WriteString(entry); err != nil {
		iostreams.DebugErr("write network trace", err)
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

// Toggle enables or disables debug logging dynamically.
// When enabling, creates a new session directory and log file.
// When disabling, closes the current session.
// Returns the new enabled state and session directory (if enabled).
func (l *Logger) Toggle() (enabled bool, sessionDir string) {
	if l == nil {
		return false, ""
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.enabled {
		// Disable: close current session
		l.closeSessionLocked()
		l.enabled = false
		return false, ""
	}

	// Enable: create new session
	configDir, err := config.Dir()
	if err != nil {
		iostreams.DebugErr("get config dir for toggle", err)
		return false, ""
	}

	// Create timestamped session directory
	fs := config.Fs()
	l.startTime = time.Now()
	sessionName := l.startTime.Format("2006-01-02_15-04-05")
	debugDir := filepath.Join(configDir, DebugDir)
	l.sessionDir = filepath.Join(debugDir, sessionName)

	if err := fs.MkdirAll(l.sessionDir, 0o700); err != nil {
		iostreams.DebugErr("create debug session dir for toggle", err)
		return false, ""
	}

	// Clean up old sessions
	cleanupOldSessions(debugDir)

	logPath := filepath.Join(l.sessionDir, MainLogFile)
	f, err := fs.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		iostreams.DebugErr("open log file for toggle", err)
		return false, ""
	}

	l.file = f
	l.enabled = true
	l.lastView = "" // Reset to ensure first view is logged

	// Write header
	l.writeHeaderLocked()

	return true, l.sessionDir
}

// closeSessionLocked closes the current session without acquiring lock.
// Must be called with lock held.
func (l *Logger) closeSessionLocked() {
	if l.file == nil {
		return
	}

	// Write footer with session duration
	duration := time.Since(l.startTime)
	footer := fmt.Sprintf("\n%s\nSession Ended (toggle): %s\nDuration: %s\n%s\n",
		strings.Repeat("=", 60),
		time.Now().Format(time.RFC3339),
		duration.Round(time.Second),
		strings.Repeat("=", 60))
	if _, err := l.file.WriteString(footer); err != nil {
		iostreams.DebugErr("write log footer on toggle", err)
	}

	if err := l.file.Close(); err != nil {
		iostreams.DebugErr("close log file on toggle", err)
	}
	l.file = nil
}

// writeHeaderLocked writes header without acquiring lock.
// Must be called with lock held.
func (l *Logger) writeHeaderLocked() {
	header := fmt.Sprintf("%s\nTUI Debug Session (toggled on)\n%s\nStarted: %s\nSession: %s\n%s\n\n",
		strings.Repeat("=", 60),
		strings.Repeat("=", 60),
		l.startTime.Format(time.RFC3339),
		l.sessionDir,
		strings.Repeat("=", 60))
	if _, err := l.file.WriteString(header); err != nil {
		iostreams.DebugErr("write log header on toggle", err)
	}
}

// Writer returns an io.Writer that writes to the debug log.
// Returns io.Discard if logging is disabled.
func (l *Logger) Writer() io.Writer {
	if l == nil || !l.enabled {
		return io.Discard
	}
	return l.file
}

// cleanupOldSessions removes debug sessions beyond MaxSessions.
// Sessions are sorted by name (timestamp-based) and oldest are removed.
func cleanupOldSessions(debugDir string) {
	fs := config.Fs()
	entries, err := afero.ReadDir(fs, debugDir)
	if err != nil {
		iostreams.DebugErr("read debug dir for cleanup", err)
		return
	}

	// Filter to directories only (session folders)
	var sessions []string
	for _, e := range entries {
		if e.IsDir() {
			sessions = append(sessions, e.Name())
		}
	}

	if len(sessions) <= MaxSessions {
		return
	}

	// Sort ascending (oldest first) - timestamp format ensures alphabetical = chronological
	slices.Sort(sessions)

	// Remove oldest sessions beyond limit
	toRemove := len(sessions) - MaxSessions
	for i := range toRemove {
		sessionPath := filepath.Join(debugDir, sessions[i])
		if err := fs.RemoveAll(sessionPath); err != nil {
			iostreams.DebugErr("remove old debug session", err)
		}
	}
}
