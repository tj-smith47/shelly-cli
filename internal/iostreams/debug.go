// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Debug functions for verbose/diagnostic output.
// These are suppressed unless verbose mode is enabled.

// Debug prints a debug message to stderr when verbose mode is enabled.
func Debug(format string, args ...any) {
	if viper.GetBool("verbose") {
		msg := fmt.Sprintf(format, args...)
		writeQuietly(os.Stderr, "debug: %s\n", msg)
	}
}

// DebugErr logs an error to stderr when verbose mode is enabled.
// Use this for non-critical errors that should not cause failure.
func DebugErr(context string, err error) {
	if err != nil && viper.GetBool("verbose") {
		writeQuietly(os.Stderr, "debug: %s: %v\n", context, err)
	}
}

// CloseWithDebug closes an io.Closer and logs any error in verbose mode.
// Use this in defer statements where close errors are not critical.
func CloseWithDebug(context string, closer interface{ Close() error }) {
	if closer != nil {
		if err := closer.Close(); err != nil {
			DebugErr(context, err)
		}
	}
}

// IOStreams debug methods

// Debug prints a debug message when verbose mode is enabled.
func (s *IOStreams) Debug(format string, args ...any) {
	if viper.GetBool("verbose") {
		msg := fmt.Sprintf(format, args...)
		writeQuietly(s.ErrOut, "debug: %s\n", msg)
	}
}

// DebugErr logs an error when verbose mode is enabled.
// Use this for non-critical errors that should not cause failure.
func (s *IOStreams) DebugErr(context string, err error) {
	if err != nil && viper.GetBool("verbose") {
		writeQuietly(s.ErrOut, "debug: %s: %v\n", context, err)
	}
}
