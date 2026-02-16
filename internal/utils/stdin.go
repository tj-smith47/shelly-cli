package utils

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// ReplaceStdinArg replaces the first "-" argument with content read from stdin.
// This enables piping patterns like: echo "kitchen" | shelly status -
//
// Returns the original args unchanged if no "-" is present.
// If stdin is a TTY, the function blocks waiting for input (standard Unix behavior),
// terminated by Ctrl+D (EOF).
func ReplaceStdinArg(args []string) ([]string, error) {
	idx := slices.Index(args, "-")
	if idx == -1 {
		return args, nil
	}

	value, err := readStdin()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	if value == "" {
		return nil, fmt.Errorf("stdin is empty; '-' argument requires input")
	}

	// Replace only the first "-"
	result := make([]string, len(args))
	copy(result, args)
	result[idx] = value
	return result, nil
}

// ResolveStdinFlag checks if the named flag has value "-" and replaces it
// with the full content read from stdin. This enables patterns like:
//
//	shelly mqtt set <device> --server -
//	shelly kvs set <device> <key> --value -
//
// Call this in RunE before using the flag value. Returns nil if the flag
// is not set, not "-", or was successfully resolved.
func ResolveStdinFlag(cmd *cobra.Command, flagName string) error {
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return fmt.Errorf("unknown flag: %s", flagName)
	}

	if !cmd.Flags().Changed(flagName) || flag.Value.String() != "-" {
		return nil
	}

	value, err := readStdin()
	if err != nil {
		return fmt.Errorf("failed to read stdin for --%s: %w", flagName, err)
	}

	if value == "" {
		return fmt.Errorf("stdin is empty; --%s - requires input", flagName)
	}

	return cmd.Flags().Set(flagName, value)
}

// readStdin reads all content from stdin, trims trailing whitespace,
// and returns the result. If stdin is a TTY, this will block until
// the user sends EOF (Ctrl+D), which is standard Unix behavior.
func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n\r \t"), nil
}
