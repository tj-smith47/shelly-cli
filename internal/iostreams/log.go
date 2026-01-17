// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// LogLevel represents the severity of a log message.
// Lower values are more verbose (trace < debug < info < warn < error).
type LogLevel int

const (
	// LevelTrace is for maximum verbosity (internal details, loops, etc.).
	LevelTrace LogLevel = iota
	// LevelDebug is for detailed diagnostic information.
	LevelDebug
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning conditions.
	LevelWarn
	// LevelError is for error conditions.
	LevelError
	// LevelNone disables all logging.
	LevelNone
)

// Level name constants for consistent string representation.
const (
	levelNameTrace   = "trace"
	levelNameDebug   = "debug"
	levelNameInfo    = "info"
	levelNameWarn    = "warn"
	levelNameError   = "error"
	levelNameNone    = "none"
	levelNameUnknown = "unknown"
)

// String returns the string representation of a log level.
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return levelNameTrace
	case LevelDebug:
		return levelNameDebug
	case LevelInfo:
		return levelNameInfo
	case LevelWarn:
		return levelNameWarn
	case LevelError:
		return levelNameError
	case LevelNone:
		return levelNameNone
	default:
		return levelNameUnknown
	}
}

// ParseLogLevel parses a string into a LogLevel.
func ParseLogLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case levelNameTrace:
		return LevelTrace
	case levelNameDebug:
		return LevelDebug
	case levelNameInfo:
		return LevelInfo
	case levelNameWarn, "warning":
		return LevelWarn
	case levelNameError:
		return LevelError
	case levelNameNone, "off", "silent":
		return LevelNone
	default:
		return LevelDebug
	}
}

// VerbosityToLevel converts a verbosity count to a log level.
//
//	0: LevelNone (no debug output)
//	1 (-v): LevelInfo (informational messages)
//	2 (-vv): LevelDebug (detailed diagnostics)
//	3+ (-vvv): LevelTrace (maximum verbosity)
func VerbosityToLevel(verbosity int) LogLevel {
	switch {
	case verbosity <= 0:
		return LevelNone
	case verbosity == 1:
		return LevelInfo
	case verbosity == 2:
		return LevelDebug
	default:
		return LevelTrace
	}
}

// LogCategory represents a logging domain for filtering.
type LogCategory string

const (
	// CategoryNetwork is for network-related operations.
	CategoryNetwork LogCategory = "network"
	// CategoryConfig is for configuration operations.
	CategoryConfig LogCategory = "config"
	// CategoryDevice is for device operations.
	CategoryDevice LogCategory = "device"
	// CategoryAPI is for API/RPC operations.
	CategoryAPI LogCategory = "api"
	// CategoryAuth is for authentication operations.
	CategoryAuth LogCategory = "auth"
	// CategoryPlugin is for plugin operations.
	CategoryPlugin LogCategory = "plugin"
	// CategoryDiscovery is for device discovery operations.
	CategoryDiscovery LogCategory = "discovery"
	// CategoryFirmware is for firmware operations.
	CategoryFirmware LogCategory = "firmware"
)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Category string    `json:"category,omitempty"`
	Message  string    `json:"message"`
	Error    string    `json:"error,omitempty"`
	Data     any       `json:"data,omitempty"`
}

// Logger provides structured logging capabilities.
type Logger struct {
	out        io.Writer
	minLevel   LogLevel
	categories map[LogCategory]bool // nil means all categories
	jsonMode   bool
}

// NewLogger creates a new Logger with default settings.
func NewLogger(out io.Writer) *Logger {
	return &Logger{
		out:      out,
		minLevel: LevelDebug,
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.minLevel = level
}

// SetJSONMode enables or disables JSON output.
func (l *Logger) SetJSONMode(enabled bool) {
	l.jsonMode = enabled
}

// SetCategories sets which categories to log. Pass nil to log all categories.
func (l *Logger) SetCategories(categories []LogCategory) {
	if categories == nil {
		l.categories = nil
		return
	}
	l.categories = make(map[LogCategory]bool)
	for _, c := range categories {
		l.categories[c] = true
	}
}

// shouldLog returns true if the message should be logged.
func (l *Logger) shouldLog(level LogLevel, category LogCategory) bool {
	if level < l.minLevel {
		return false
	}
	if l.categories != nil && category != "" && !l.categories[category] {
		return false
	}
	return true
}

// Log writes a log entry.
func (l *Logger) Log(level LogLevel, category LogCategory, format string, args ...any) {
	if !l.shouldLog(level, category) {
		return
	}

	msg := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Time:     time.Now(),
		Level:    level.String(),
		Category: string(category),
		Message:  msg,
	}

	l.writeEntry(entry)
}

// LogErr writes a log entry with an error.
func (l *Logger) LogErr(level LogLevel, category LogCategory, context string, err error) {
	if err == nil || !l.shouldLog(level, category) {
		return
	}

	entry := LogEntry{
		Time:     time.Now(),
		Level:    level.String(),
		Category: string(category),
		Message:  context,
		Error:    err.Error(),
	}

	l.writeEntry(entry)
}

// LogWithData writes a log entry with additional data.
func (l *Logger) LogWithData(level LogLevel, category LogCategory, msg string, data any) {
	if !l.shouldLog(level, category) {
		return
	}

	entry := LogEntry{
		Time:     time.Now(),
		Level:    level.String(),
		Category: string(category),
		Message:  msg,
		Data:     data,
	}

	l.writeEntry(entry)
}

// writeEntry writes the log entry in the configured format.
func (l *Logger) writeEntry(entry LogEntry) {
	if l.jsonMode {
		l.writeJSONEntry(entry)
		return
	}
	l.writeTextEntry(entry)
}

// writeJSONEntry writes the entry as JSON.
func (l *Logger) writeJSONEntry(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	writeQuietly(l.out, "%s\n", data)
}

// writeTextEntry writes the entry in human-readable format.
func (l *Logger) writeTextEntry(entry LogEntry) {
	prefix := entry.Level
	if entry.Category != "" {
		prefix = fmt.Sprintf("%s:%s", entry.Level, entry.Category)
	}
	if entry.Error != "" {
		writeQuietly(l.out, "%s: %s: %s\n", prefix, entry.Message, entry.Error)
		return
	}
	writeQuietly(l.out, "%s: %s\n", prefix, entry.Message)
}

// Package-level structured logging

var defaultLogger = NewLogger(os.Stderr)

// ConfigureLogger configures the default logger from viper settings.
// Call this after viper is initialized.
//
// Configuration keys:
//   - verbosity (int): Verbosity level (0=none, 1=info, 2=debug, 3=trace)
//   - log.level (string): Override level by name (trace, debug, info, warn, error)
//   - log.json (bool): Enable JSON output
//   - log.categories (string): Comma-separated category filter
func ConfigureLogger() {
	// First check verbosity count (-v/-vv/-vvv)
	verbosity := viper.GetInt("verbosity")
	if verbosity > 0 {
		defaultLogger.SetLevel(VerbosityToLevel(verbosity))
	}

	// log.level overrides verbosity if set explicitly
	if levelStr := viper.GetString("log.level"); levelStr != "" {
		defaultLogger.SetLevel(ParseLogLevel(levelStr))
	}

	defaultLogger.SetJSONMode(viper.GetBool("log.json"))

	if catStr := viper.GetString("log.categories"); catStr != "" {
		cats := strings.Split(catStr, ",")
		categories := make([]LogCategory, 0, len(cats))
		for _, c := range cats {
			categories = append(categories, LogCategory(strings.TrimSpace(c)))
		}
		defaultLogger.SetCategories(categories)
	}
}

// GetVerbosity returns the current verbosity level from viper.
func GetVerbosity() int {
	return viper.GetInt("verbosity")
}

// IsVerbose returns true if any verbosity is enabled.
func IsVerbose() bool {
	return GetVerbosity() > 0
}

// Log writes a structured log entry to the default logger.
// Only outputs when verbosity is sufficient for the log level.
func Log(level LogLevel, category LogCategory, format string, args ...any) {
	if GetVerbosity() <= 0 {
		return
	}
	defaultLogger.Log(level, category, format, args...)
}

// LogErr writes a structured log entry with an error to the default logger.
// Only outputs when verbosity is sufficient for the log level.
func LogErr(level LogLevel, category LogCategory, context string, err error) {
	if GetVerbosity() <= 0 {
		return
	}
	defaultLogger.LogErr(level, category, context, err)
}

// LogWithData writes a structured log entry with data to the default logger.
// Only outputs when verbosity is sufficient for the log level.
func LogWithData(level LogLevel, category LogCategory, msg string, data any) {
	if GetVerbosity() <= 0 {
		return
	}
	defaultLogger.LogWithData(level, category, msg, data)
}

// IOStreams structured logging methods

// Logger returns a Logger for this IOStreams instance.
// The logger is configured based on current viper settings.
func (s *IOStreams) Logger() *Logger {
	logger := NewLogger(s.ErrOut)
	verbosity := GetVerbosity()
	if verbosity > 0 {
		logger.SetLevel(VerbosityToLevel(verbosity))
	} else {
		logger.SetLevel(LevelNone)
	}
	logger.SetJSONMode(viper.GetBool("log.json"))
	return logger
}

// Log writes a structured log entry.
// Only outputs when verbosity is sufficient for the log level.
func (s *IOStreams) Log(level LogLevel, category LogCategory, format string, args ...any) {
	if GetVerbosity() <= 0 {
		return
	}
	s.Logger().Log(level, category, format, args...)
}

// LogErr writes a structured log entry with an error.
// Only outputs when verbosity is sufficient for the log level.
func (s *IOStreams) LogErr(level LogLevel, category LogCategory, context string, err error) {
	if GetVerbosity() <= 0 {
		return
	}
	s.Logger().LogErr(level, category, context, err)
}

// LogWithData writes a structured log entry with additional data.
// Only outputs when verbosity is sufficient for the log level.
func (s *IOStreams) LogWithData(level LogLevel, category LogCategory, msg string, data any) {
	if GetVerbosity() <= 0 {
		return
	}
	s.Logger().LogWithData(level, category, msg, data)
}
