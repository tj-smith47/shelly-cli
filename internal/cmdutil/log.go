// Package cmdutil provides log file utilities.
package cmdutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// GetLogPath returns the path to the CLI log file.
func GetLogPath() (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "shelly.log"), nil
}

// ReadLastLines reads the last N lines from a file.
func ReadLastLines(path string, n int) ([]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Log file path is from config dir
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	// Remove empty last line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) <= n {
		return lines, nil
	}

	return lines[len(lines)-n:], nil
}
