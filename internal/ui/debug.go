// Package ui provides user interaction components for the CLI.
package ui

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Debug prints a debug message to stderr when verbose mode is enabled.
func Debug(format string, args ...any) {
	if viper.GetBool("verbose") {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "debug: %s\n", msg)
	}
}

// DebugErr logs an error to stderr when verbose mode is enabled.
// Use this for non-critical errors that should not cause failure.
func DebugErr(context string, err error) {
	if err != nil && viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "debug: %s: %v\n", context, err)
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
