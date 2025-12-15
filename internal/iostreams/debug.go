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

// Debug functions for verbose/diagnostic output.
// Organized in three tiers:
//
// Tier 1 - Simple (most common, no category):
//   - Debug(format, args...)     - prints at -vv level
//   - DebugErr(context, err)     - logs error at -v level
//   - Trace(format, args...)     - prints at -vvv level
//
// Tier 2 - Category-aware (for --log-categories filtering):
//   - DebugCat(cat, format, args...)   - categorized debug at -vv
//   - DebugErrCat(cat, context, err)   - categorized error at -v
//   - TraceCat(cat, format, args...)   - categorized trace at -vvv
//
// Tier 3 - Full control (explicit level, see log.go):
//   - Log(level, cat, format, args...)
//   - LogErr(level, cat, context, err)

// =============================================================================
// Tier 1: Simple Functions (no category required)
// =============================================================================

// Debug prints a debug message to stderr when verbosity >= 2 (-vv).
func Debug(format string, args ...any) {
	if GetVerbosity() >= 2 {
		writeDebugOutput(os.Stderr, "debug", "", format, args...)
	}
}

// DebugErr logs an error to stderr when verbosity >= 1 (-v).
// Safe to call with nil error (no output).
func DebugErr(context string, err error) {
	if err != nil && GetVerbosity() >= 1 {
		writeDebugErrOutput(os.Stderr, "", context, err)
	}
}

// Trace prints a trace message to stderr when verbosity >= 3 (-vvv).
// Use for maximum verbosity output (internal details, loops, etc.).
func Trace(format string, args ...any) {
	if GetVerbosity() >= 3 {
		writeDebugOutput(os.Stderr, "trace", "", format, args...)
	}
}

// CloseWithDebug closes an io.Closer and logs any error when verbosity >= 1.
// Use this in defer statements where close errors are not critical.
func CloseWithDebug(context string, closer interface{ Close() error }) {
	if closer != nil {
		if err := closer.Close(); err != nil {
			DebugErr(context, err)
		}
	}
}

// =============================================================================
// Tier 2: Category-Aware Functions (for --log-categories filtering)
// =============================================================================

// DebugCat prints a categorized debug message when verbosity >= 2 (-vv).
// Message is filtered if --log-categories is set and doesn't include this category.
func DebugCat(cat LogCategory, format string, args ...any) {
	if GetVerbosity() >= 2 && shouldLogCategory(cat) {
		writeDebugOutput(os.Stderr, "debug", cat, format, args...)
	}
}

// DebugErrCat logs a categorized error when verbosity >= 1 (-v).
// Message is filtered if --log-categories is set and doesn't include this category.
// Safe to call with nil error (no output).
func DebugErrCat(cat LogCategory, context string, err error) {
	if err != nil && GetVerbosity() >= 1 && shouldLogCategory(cat) {
		writeDebugErrOutput(os.Stderr, cat, context, err)
	}
}

// TraceCat prints a categorized trace message when verbosity >= 3 (-vvv).
// Message is filtered if --log-categories is set and doesn't include this category.
func TraceCat(cat LogCategory, format string, args ...any) {
	if GetVerbosity() >= 3 && shouldLogCategory(cat) {
		writeDebugOutput(os.Stderr, "trace", cat, format, args...)
	}
}

// =============================================================================
// IOStreams Methods
// =============================================================================

// Debug prints a debug message when verbosity >= 2 (-vv).
func (s *IOStreams) Debug(format string, args ...any) {
	if GetVerbosity() >= 2 {
		writeDebugOutput(s.ErrOut, "debug", "", format, args...)
	}
}

// DebugErr logs an error when verbosity >= 1 (-v).
// Safe to call with nil error (no output).
func (s *IOStreams) DebugErr(context string, err error) {
	if err != nil && GetVerbosity() >= 1 {
		writeDebugErrOutput(s.ErrOut, "", context, err)
	}
}

// Trace prints a trace message when verbosity >= 3 (-vvv).
func (s *IOStreams) Trace(format string, args ...any) {
	if GetVerbosity() >= 3 {
		writeDebugOutput(s.ErrOut, "trace", "", format, args...)
	}
}

// DebugCat prints a categorized debug message when verbosity >= 2 (-vv).
func (s *IOStreams) DebugCat(cat LogCategory, format string, args ...any) {
	if GetVerbosity() >= 2 && shouldLogCategory(cat) {
		writeDebugOutput(s.ErrOut, "debug", cat, format, args...)
	}
}

// DebugErrCat logs a categorized error when verbosity >= 1 (-v).
func (s *IOStreams) DebugErrCat(cat LogCategory, context string, err error) {
	if err != nil && GetVerbosity() >= 1 && shouldLogCategory(cat) {
		writeDebugErrOutput(s.ErrOut, cat, context, err)
	}
}

// TraceCat prints a categorized trace message when verbosity >= 3 (-vvv).
func (s *IOStreams) TraceCat(cat LogCategory, format string, args ...any) {
	if GetVerbosity() >= 3 && shouldLogCategory(cat) {
		writeDebugOutput(s.ErrOut, "trace", cat, format, args...)
	}
}

// =============================================================================
// Internal Helpers
// =============================================================================

// shouldLogCategory checks if a category should be logged based on --log-categories filter.
// Returns true if no filter is set (empty = show all) or if the category is in the filter.
func shouldLogCategory(cat LogCategory) bool {
	filter := viper.GetString("log.categories")
	if filter == "" {
		return true // No filter = show all categories
	}
	for _, c := range strings.Split(filter, ",") {
		if strings.TrimSpace(c) == string(cat) {
			return true
		}
	}
	return false
}

// writeDebugOutput writes a debug message in the appropriate format.
func writeDebugOutput(w io.Writer, level string, cat LogCategory, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)

	if viper.GetBool("log.json") {
		writeJSONLog(w, level, cat, msg, nil)
		return
	}

	// Human-readable format
	prefix := level
	if cat != "" {
		prefix = fmt.Sprintf("%s:%s", level, cat)
	}
	writeQuietly(w, "%s: %s\n", prefix, msg)
}

// writeDebugErrOutput writes an error message in the appropriate format.
// Always uses "debug" as the level name since all debug error output is at debug level.
func writeDebugErrOutput(w io.Writer, cat LogCategory, context string, err error) {
	const level = "debug"
	if viper.GetBool("log.json") {
		writeJSONLog(w, level, cat, context, err)
		return
	}

	// Human-readable format
	prefix := level
	if cat != "" {
		prefix = fmt.Sprintf("%s:%s", level, cat)
	}
	writeQuietly(w, "%s: %s: %v\n", prefix, context, err)
}

// debugLogEntry is a simplified log entry for debug output.
type debugLogEntry struct {
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Category string    `json:"category,omitempty"`
	Message  string    `json:"message"`
	Error    string    `json:"error,omitempty"`
}

// writeJSONLog writes a JSON-formatted log entry.
func writeJSONLog(w io.Writer, level string, cat LogCategory, msg string, err error) {
	entry := debugLogEntry{
		Time:     time.Now(),
		Level:    level,
		Category: string(cat),
		Message:  msg,
	}
	if err != nil {
		entry.Error = err.Error()
	}

	data, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		return // Silently fail on marshal error
	}
	writeQuietly(w, "%s\n", data)
}
